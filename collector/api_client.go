package collector

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"server_exporter/tools"
	"strconv"
	"strings"
	"sync"
	"time"
)

const redfishRootPath = "/redfish/v1"

const (
	UNKNOWN = iota
	DELL
	HPE
	LENOVO
	INSPUR
	H3C
	HUAWEI
)

type Client struct {
	deviceName  string
	host        string
	username    string
	password    string
	httpClient  *http.Client
	vendor      int
	version     int
	systemPath  string
	thermalPath string
	powerPath   string
	storagePath string
	memoryPath  string
	networkPath string
}

func newHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(120) * time.Second,
	}
}

func NewClient(authInfo tools.Data, deviceName string) (*Client, error) {
	client := &Client{
		deviceName: deviceName,
		host:       authInfo.Dat.Host,
		username:   authInfo.Dat.Account,
		password:   authInfo.Dat.Password,
		httpClient: newHttpClient(),
	}

	err := client.findAllEndpoints()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (client *Client) findAllEndpoints() error {
	var root V1Response
	var group GroupResponse
	var chassis ChassisResponse
	var system SystemResponse
	var err error

	// Root
	err = client.redfishGet(redfishRootPath, &root)
	if err != nil {
		return err
	}
	// System

	err = client.redfishGet(root.Systems.OdataId, &group)
	if err != nil {
		return err
	}

	client.systemPath = group.Members[0].OdataId

	// Chassis
	err = client.redfishGet(root.Chassis.OdataId, &group)
	if err != nil {
		return err
	}

	// Thermal and Power
	err = client.redfishGet(group.Members[0].OdataId, &chassis)
	if err != nil {
		return err
	}

	err = client.redfishGet(client.systemPath, &system)
	if err != nil {
		return err
	}

	client.storagePath = system.Storage.OdataId
	client.memoryPath = system.Memory.OdataId
	client.networkPath = system.NetworkInterfaces.OdataId
	client.thermalPath = chassis.Thermal.OdataId
	client.powerPath = chassis.Power.OdataId

	// Vendor
	m := strings.ToLower(system.Manufacturer)
	log.Printf("vender is %q", system.Manufacturer)
	if strings.Contains(m, "dell") {
		client.vendor = DELL
	} else if strings.Contains(m, "hpe") {
		client.vendor = HPE
	} else if strings.Contains(m, "lenovo") {
		client.vendor = LENOVO
	} else if strings.Contains(m, "inspur") {
		client.vendor = INSPUR
	} else if strings.Contains(m, "h3c") {
		client.vendor = H3C
	} else if strings.Contains(m, "huawei") {
		client.vendor = HUAWEI
	}

	// Fix for Inspur bug
	if client.vendor == INSPUR {
		client.storagePath = strings.ReplaceAll(client.storagePath, "Storages", "Storage")
	}

	// Fix for iLO 4 machines
	if client.vendor == HPE {
		if strings.Contains(root.Name, "HP RESTful") {
			client.memoryPath = "/redfish/v1/Systems/1/Memory/"
			client.storagePath = "/redfish/v1/Systems/1/SmartStorage/ArrayControllers/"
			client.version = 4
		}
	}

	return nil
}

func (client *Client) RefreshSensors(mc *Collector, ch chan string) error {
	var resp ThermalResponse

	err := client.redfishGet(client.thermalPath, &resp)
	if err != nil {
		return err
	}

	for n, t := range resp.Temperatures {
		if t.Status.State != StateEnabled && t.Status.State != "ENABLE" {
			continue
		}

		if t.ReadingCelsius < 0 {
			continue
		}

		id := t.GetId(n)
		ch <- mc.NewSensorsTemperature(t.ReadingCelsius, id, t.Name, "celsius")
	}

	for n, f := range resp.Fans {
		if f.Status.State != StateEnabled && f.Status.State != "ENABLE" {
			continue
		}

		name := f.GetName()
		if name == "" {
			continue
		}

		units := f.GetUnits()
		if units == "" {
			continue
		}

		id := f.GetId(n)
		ch <- mc.NewSensorsFanHealth(id, name, f.Status.Health)
		ch <- mc.NewSensorsFanSpeed(f.GetReading(), "", name, strings.ToLower(units))
	}

	return nil
}

func (client *Client) RefreshSystem(mc *Collector, ch chan string) error {
	var resp SystemResponse
	err := client.redfishGet(client.systemPath, &resp)
	if err != nil {
		return err
	}
	ch <- mc.NewSystemPowerOn(resp.PowerState)
	ch <- mc.NewSystemHealth(resp.Status.Health, client.deviceName)
	ch <- mc.NewSystemIndicatorLED(resp.IndicatorLED)
	if resp.MemorySummary != nil {
		ch <- mc.NewSystemMemorySize(resp.MemorySummary.TotalSystemMemoryGiB * 1073741824)
	}
	if resp.ProcessorSummary != nil {
		ch <- mc.NewSystemCpuCount(resp.ProcessorSummary.Count, resp.ProcessorSummary.Model)
	}
	ch <- mc.NewSystemBiosInfo(resp.BiosVersion)
	ch <- mc.NewSystemMachineInfo(resp.Manufacturer, resp.Model, resp.SerialNumber, resp.SKU)

	return nil
}

func (client *Client) RefreshNetwork(mc *Collector, ch chan string) error {
	var wg sync.WaitGroup
	group := GroupResponse{}
	err := client.redfishGet(client.networkPath, &group)
	if err != nil {
		return err
	}

	for _, c := range group.Members {
		ni := NetworkInterface{}
		err = client.redfishGet(c.OdataId, &ni)
		if err != nil {
			return err
		}

		if ni.Status.State != StateEnabled {
			continue
		}

		ch <- mc.NewNetworkInterfaceHealth(ni.Id, ni.Status.Health)

		ports := GroupResponse{}
		err = client.redfishGet(ni.GetPorts(), &ports)
		if err != nil {
			return err
		}

		port := NetworkPort{}
		for _, c := range ports.Members {
			wg.Add(1)
			go func(c Odata) {
				err = client.redfishGet(c.OdataId, &port)
				if err != nil {
					log.Println(err)
				}
				if client.vendor == HUAWEI && port.Status.State == "Enabled" {
					ch <- mc.NewNetworkPortHealth(ni.Id, port.Id, "OK")
				} else {
					ch <- mc.NewNetworkPortHealth(ni.Id, port.Id, port.Status.Health)
				}
				ch <- mc.NewNetworkPortSpeed(ni.Id, port.Id, port.GetSpeed())
				ch <- mc.NewNetworkPortLinkUp(ni.Id, port.Id, port.LinkStatus)
				wg.Done()
			}(c)
		}
		wg.Wait()
	}

	return nil
}

func (client *Client) RefreshPower(mc *Collector, ch chan string) error {
	var resp PowerResponse

	err := client.redfishGet(client.powerPath, &resp)
	if err != nil {
		return err
	}
	for i, psu := range resp.PowerSupplies {
		if psu.Status.State != StateEnabled && psu.Status.State != "ENABLE" {
			continue
		}

		id := strconv.Itoa(i)
		ch <- mc.NewPowerSupplyHealth(psu.Status.Health, id)
		ch <- mc.NewPowerSupplyInputWatts(psu.PowerInputWatts, id)
		ch <- mc.NewPowerSupplyInputVoltage(psu.LineInputVoltage, id)
		ch <- mc.NewPowerSupplyOutputWatts(psu.GetOutputPower(), id)
		ch <- mc.NewPowerSupplyCapacityWatts(psu.PowerCapacityWatts, id)
		ch <- mc.NewPowerSupplyEfficiencyPercent(psu.EfficiencyPercent, id)
	}

	for i, pc := range resp.PowerControl {
		id := strconv.Itoa(i)
		ch <- mc.NewPowerControlConsumedWatts(pc.PowerConsumedWatts, id, pc.Name)
		ch <- mc.NewPowerControlCapacityWatts(pc.PowerCapacityWatts, id, pc.Name)

		if pc.PowerMetrics == nil {
			continue
		}

		pm := pc.PowerMetrics
		ch <- mc.NewPowerControlMinConsumedWatts(pm.MinConsumedWatts, id, pc.Name)
		ch <- mc.NewPowerControlMaxConsumedWatts(pm.MaxConsumedWatts, id, pc.Name)
		ch <- mc.NewPowerControlAvgConsumedWatts(pm.AverageConsumedWatts, id, pc.Name)
		ch <- mc.NewPowerControlInterval(pm.IntervalInMinutes, id, pc.Name)
	}

	return nil
}

func (client *Client) RefreshIdracSel(mc *Collector, ch chan string) error {
	if client.vendor == HUAWEI {
		var resp LogRes
		err := client.redfishGetHuaweiLog(&resp)
		if err != nil {
			return err
		}

		for _, e := range resp.Error.ExtendedInfo[0].Oem.Huawei.SelLogEntries {
			if e.AlertTime == "" {
				continue
			}
			t, _ := time.Parse("2006-01-02 15:04:05", e.AlertTime)
			switch {
			case e.Level == "0":
				e.Level = "normal"
			case e.Level == "1":
				e.Level = "warning"
			case e.Level == "2":
				e.Level = "error"
			case e.Level == "3":
				e.Level = "critical"
			default:
				e.Level = "unknown"
			}
			ch <- mc.NewSelEntry(e.EventID, e.EventDesc, "Unknown", e.Level, t)
		}
	} else if client.vendor == INSPUR {
		var resp InspurSelResponse
		err := client.redfishGet(redfishRootPath+"/Managers/1/LogServices/Log/Entries", &resp)
		if err != nil {
			return err
		}
		for _, e := range resp.Members {
			st := e.SensorType
			if st == "" {
				st = "Unknown"
			}
			// translate origin time
			finalTime, _ := time.Parse("06-01-02 15:04:05", e.Created)
			ch <- mc.NewSelEntry(e.ID, e.Message, st, e.Severity, finalTime)
		}
	} else {
		var resp IdracSelResponse
		err := client.redfishGet(redfishRootPath+"/Managers/iDRAC.Embedded.1/Logs/Sel", &resp)
		if err != nil {
			return err
		}
		log.Printf("log is %s", resp.Members)
		for _, e := range resp.Members {
			st := string(e.SensorType)
			if st == "" {
				st = "Unknown"
			}
			ch <- mc.NewSelEntry(e.Id, e.Message, st, e.Severity, e.Created)
		}
	}

	return nil
}

type inspurDriveResponse struct {
	Members []Odata `json:"Members"`
}

func (client *Client) RefreshStorage(mc *Collector, ch chan string) error {
	var wg sync.WaitGroup
	group := GroupResponse{}
	sPath := client.storagePath
	if client.vendor == INSPUR {
		sPath = "/redfish/v1/Systems/1/Storages"
	}
	err := client.redfishGet(sPath, &group)
	if err != nil {
		return err
	}
	if client.vendor == INSPUR {
		grp := inspurDriveResponse{}
		ctlr := StorageController{}
		err = client.redfishGet("/redfish/v1/Chassis/1/Drives", &grp)
		if err != nil {
			return err
		}

		ctlr.Drives = grp.Members
		drive := Drive{}
		limitChan := make(chan int, 8)

		for _, d := range ctlr.Drives {
			wg.Add(1)
			go func(d Odata) {
				limitChan <- 1
				err = client.redfishGet(d.OdataId, &drive)
				if err != nil {
					log.Printf("磁盘信息采集错误-%s", err.Error())
				}
				// iLO 4
				if (client.vendor == HPE) && (client.version == 4) {
					drive.CapacityBytes = 1024 * 1024 * drive.CapacityMiB
					drive.Protocol = drive.InterfaceType
					drive.PredictedLifeLeft = 100 - drive.SSDEnduranceUtilizationPercentage
				}

				var id = drive.Id
				if strings.Contains(id, "HDDPlaneDisk") {
					id = strings.ReplaceAll(id, "HDDPlaneDisk", "磁盘槽")
				} else if strings.Contains(id, "mainboardSDCard") {
					id = strings.ReplaceAll(id, "mainboardSDCard", "主板SD卡")
				} else if strings.Contains(id, ":Enclosure.Internal") {
					id = strings.Split(id, ":")[0]
					id = strings.ReplaceAll(id, "Disk.Bay.", "磁盘槽")
				} else if strings.Contains(id, "HDDPlaneDisk00000") {
					id = strings.ReplaceAll(id, "HDDPlaneDisk00000", "磁盘槽")
				}

				ch <- mc.NewDriveInfo(id, drive.Name, drive.Manufacturer, strings.Trim(drive.Model, " "), strings.Trim(drive.SerialNumber, " "), drive.MediaType, drive.Protocol, drive.GetSlot())
				ch <- mc.NewDriveHealth(id, drive.Status.Health)
				ch <- mc.NewDriveCapacity(id, drive.CapacityBytes)
				ch <- mc.NewDriveLifeLeft(id, drive.PredictedLifeLeft)
				<-limitChan
				wg.Done()
			}(d)
		}
		wg.Wait()
	} else {
		for _, c := range group.Members {
			wg.Add(1)
			go func(d Odata) {
				ctlr := StorageController{}
				err = client.redfishGet(c.OdataId, &ctlr)
				if err != nil {
					log.Printf("磁盘信息采集错误-%s", err.Error())
				}
				// iLO 4
				if (client.vendor == HPE) && (client.version == 4) {
					grp := GroupResponse{}
					err = client.redfishGet(c.OdataId+"DiskDrives/", &grp)
					if err != nil {
						log.Printf("磁盘信息采集错误-%s", err.Error())
					}
					ctlr.Drives = grp.Members
				}
				var wg2 sync.WaitGroup
				for _, d := range ctlr.Drives {
					wg2.Add(1)
					go func(d Odata) {
						drive := Drive{}
						err = client.redfishGet(d.OdataId, &drive)
						if err != nil {
							log.Printf("磁盘信息采集错误-%s", err.Error())
						}

						// iLO 4
						if (client.vendor == HPE) && (client.version == 4) {
							drive.CapacityBytes = 1024 * 1024 * drive.CapacityMiB
							drive.Protocol = drive.InterfaceType
							drive.PredictedLifeLeft = 100 - drive.SSDEnduranceUtilizationPercentage
						}

						var id = drive.Id
						if strings.Contains(id, "HDDPlaneDisk") {
							id = strings.ReplaceAll(id, "HDDPlaneDisk", "磁盘槽")
						} else if strings.Contains(id, "mainboardSDCard") {
							id = strings.ReplaceAll(id, "mainboardSDCard", "主板SD卡")
						} else if strings.Contains(id, ":Enclosure.Internal") {
							id = strings.Split(id, ":")[0]
							id = strings.ReplaceAll(id, "Disk.Bay.", "磁盘槽")
						} else if strings.Contains(id, "HDDPlaneDisk00000") {
							id = strings.ReplaceAll(id, "HDDPlaneDisk00000", "磁盘槽")
						}

						ch <- mc.NewDriveInfo(id, drive.Name, drive.Manufacturer, drive.Model, drive.SerialNumber, drive.MediaType, drive.Protocol, drive.GetSlot())
						ch <- mc.NewDriveHealth(id, drive.Status.Health)
						ch <- mc.NewDriveCapacity(id, drive.CapacityBytes)
						ch <- mc.NewDriveLifeLeft(id, drive.PredictedLifeLeft)
						wg2.Done()
					}(d)
				}
				wg2.Wait()
			}(c)
		}
		wg.Wait()
	}
	return nil
}

func (client *Client) RefreshMemory(mc *Collector, ch chan string) error {
	var group GroupResponse
	var m Memory

	err := client.redfishGet(client.memoryPath, &group)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	limitChan := make(chan int, 7)
	for _, c := range group.Members {
		if m.Status.State == StateAbsent {
			continue
		}
		wg.Add(1)
		go func() {
			limitChan <- 1
			err = client.redfishGet(c.OdataId, &m)
			if err != nil {
				log.Printf(err.Error())
			}
			// iLO 4
			if (client.vendor == HPE) && (client.version == 4) {
				m.Manufacturer = strings.TrimSpace(m.Manufacturer)
				m.RankCount = m.Rank
				m.MemoryDeviceType = m.DIMMType
				m.Status.Health = m.DIMMStatus
				m.CapacityMiB = m.SizeMB
			}
			var id = m.Id
			if strings.Contains(id, "mainboardDIMM") {
				id = strings.ReplaceAll(id, "mainboardDIMM", "内存槽")
				id = strings.ReplaceAll(id, ".Socket.", "-")
			} else if strings.Contains(id, "DIMM.Socket") {
				id = strings.ReplaceAll(id, "DIMM.Socket", "内存槽")
			} else if strings.Contains(id, "DIMM") {
				id = strings.ReplaceAll(id, "DIMM", "内存槽")
			}

			ch <- mc.NewMemoryModuleInfo(id, m.Name, m.Manufacturer, m.MemoryDeviceType, m.SerialNumber, m.ErrorCorrection, m.RankCount)
			ch <- mc.NewMemoryModuleHealth(id, m.Status.Health)
			ch <- mc.NewMemoryModuleCapacity(id, m.CapacityMiB*1048576)
			ch <- mc.NewMemoryModuleSpeed(id, m.OperatingSpeedMhz)
			<-limitChan
			wg.Done()
		}()
	}
	wg.Wait()
	return nil
}

func (client *Client) redfishGet(path string, res interface{}) error {
	if !strings.HasPrefix(path, redfishRootPath) {
		return fmt.Errorf("invalid url for redfish request")
	}

	url := "https://" + client.host + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(client.username, client.password)

	resp, err := client.httpClient.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Printf("Failed to query url %q: %v", url, err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("Query to url %q returned unexpected status code: %d (%s)", url, resp.StatusCode, resp.Status)
		return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		log.Printf("Error decoding response from url %q: %v", url, err)
		type new struct {
			Name        string   `json:"Name"`
			Description string   `json:"Description"`
			Members     []string `json:"Members"`
		}
		var orin = new{Members: []string{"/redfish/v1/Chassis/1"}}
		var correctFormat = GroupResponse{Members: []Odata{{OdataId: "0"}}}
		correctFormat.Members[0].OdataId = orin.Members[0]
		jsonData, _ := json.Marshal(correctFormat)
		json.NewDecoder(strings.NewReader(string(jsonData))).Decode(&res)

		return nil
	}

	return nil
}

type LogRes struct {
	Error struct {
		Code         string         `json:"code"`
		Message      string         `json:"message"`
		ExtendedInfo []ExtendedInfo `json:"@Message.ExtendedInfo"`
	} `json:"error"`
}

type ExtendedInfo struct {
	Oem Oem `json:"Oem"`
}

type Oem struct {
	Huawei Huawei `json:"Huawei"`
}

type Huawei struct {
	SelLogEntries []SelLogEntry `json:"SelLogEntries"`
}

type SelLogEntry struct {
	Level        string          `json:"level,omitempty"`
	EventID      string          `json:"eventid,omitempty"`
	SubjectType  string          `json:"subjecttype,omitempty"`
	EventDesc    string          `json:"eventdesc,omitempty"`
	TrigMode     string          `json:"trigmode,omitempty"`
	AlertTime    string          `json:"alerttime,omitempty"`
	Status       string          `json:"status,omitempty"`
	EventCode    string          `json:"eventcode,omitempty"`
	OldEventCode string          `json:"oldeventcode,omitempty"`
	EventSubject string          `json:"eventsubject,omitempty"`
	EventSugg    string          `json:"eventsugg,omitempty"`
	Number       json.RawMessage `json:"number,omitempty"`
}

func (client *Client) redfishGetHuaweiLog(res *LogRes) error {
	url := "https://" + client.host + "/redfish/v1/Systems/1/LogServices/Log1/Actions/Oem/Huawei/LogService.QuerySelLogEntries"
	data := map[string]interface{}{
		"StartEntryId": 1,
		"EntriesCount": 50,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(client.username, client.password)

	resp, err := client.httpClient.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Printf("Failed to query url %q: %v", url, err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Printf("Query to url %q returned unexpected status code: %d (%s)", url, resp.StatusCode, resp.Status)
		return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}
	re, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(re, &res)
	if err != nil {
		log.Printf("Error decoding response from url %v", err)
	}

	return nil
}
