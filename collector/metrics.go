package collector

import (
	"strconv"
	"time"
)

func health2value(health string) float64 {
	switch health {
	case "OK":
		return 0
	case "Warning":
		return 1
	case "Critical":
		return 2
	}
	return 10
}

func linkstatus2value(status string) float64 {
	switch status {
	case "Up", "LinkUp":
		return 1
	}
	return 0
}

func (collector *Collector) NewSystemPowerOn(state string) string {
	var value float64
	if state == "On" {
		value = 1
	}
	return "idrac_system_power_on{ip=\"" + collector.client.host + "\", } " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewSystemHealth(health, deviceName string) string {
	value := health2value(health)

	return "idrac_system_health{ip=\"" + collector.client.host + "\", device_name=\"" + deviceName + "\",status=\"" + health + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewSystemIndicatorLED(state string) string {
	var value float64
	if state != "Off" {
		value = 1
	}
	return "idrac_system_indicator_led_on{ip=\"" + collector.client.host + "\", state=\"" + state + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewSystemMemorySize(memory float64) string {
	return "idrac_system_memory_size_bytes{ip=\"" + collector.client.host + "\"} " + strconv.FormatFloat(memory, 'f', 0, 64)
}

func (collector *Collector) NewSystemCpuCount(cpus int, model string) string {
	return "idrac_system_cpu_count{ip=\"" + collector.client.host + "\", model=\"" + model + "\"} " + strconv.Itoa(cpus)
}

func (collector *Collector) NewSystemBiosInfo(version string) string {
	return "idrac_system_bios_info{ip=\"" + collector.client.host + "\", version=\"" + version + "\"} 1"
}

func (collector *Collector) NewSystemMachineInfo(manufacturer, model, serial, sku string) string {
	return "idrac_system_machine_info{ip=\"" + collector.client.host + "\", manufacturer=\"" + manufacturer + "\", model=\"" + model + "\", serial=\"" + serial + "\", sku=\"" + sku + "\"} 1"
}

func (collector *Collector) NewSensorsTemperature(temperature float64, id, name, units string) string {
	return "idrac_sensors_temperature{ip=\"" + collector.client.host + "\", id=\"" + id + "\", name=\"" + name + "\", units=\"" + units + "\"} " + strconv.FormatFloat(temperature, 'f', 0, 64)
}

func (collector *Collector) NewSensorsFanHealth(id, name, health string) string {
	value := health2value(health)
	return "idrac_sensors_fan_health{ip=\"" + collector.client.host + "\", id=\"" + id + "\", name=\"" + name + "\", status=\"" + health + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewSensorsFanSpeed(speed float64, id, name, units string) string {
	return "idrac_sensors_fan_speed{ip=\"" + collector.client.host + "\", id=\"" + id + "\", name=\"" + name + "\", units=\"" + units + "\"} " + strconv.FormatFloat(speed, 'f', 0, 64)
}

func (collector *Collector) NewPowerSupplyHealth(health, id string) string {
	value := health2value(health)
	return "idrac_power_supply_health{ip=\"" + collector.client.host + "\", id=\"" + id + "\", status=\"" + health + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerSupplyInputWatts(value float64, id string) string {
	return "idrac_power_supply_input_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerSupplyInputVoltage(value float64, id string) string {
	return "idrac_power_supply_input_voltage{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerSupplyOutputWatts(value float64, id string) string {
	return "idrac_power_supply_output_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerSupplyCapacityWatts(value float64, id string) string {
	return "idrac_power_supply_capacity_watts{ip=\"" + collector.client.host + ", id=\"" + id + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerSupplyEfficiencyPercent(value float64, id string) string {
	return "idrac_power_supply_efficiency_percent{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerControlConsumedWatts(value float64, id, name string) string {
	return "idrac_power_control_consumed_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\" , name=\"" + name + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerControlCapacityWatts(value float64, id, name string) string {
	return "idrac_power_control_capacity_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\" , name=\"" + name + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerControlMinConsumedWatts(value float64, id, name string) string {
	return "idrac_power_control_min_consumed_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\" , name=\"" + name + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerControlMaxConsumedWatts(value float64, id, name string) string {
	return "idrac_power_control_max_consumed_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\" , name=\"" + name + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerControlAvgConsumedWatts(value float64, id, name string) string {
	return "idrac_power_control_avg_consumed_watts{ip=\"" + collector.client.host + "\", id=\"" + id + "\" , name=\"" + name + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewPowerControlInterval(interval int, id, name string) string {
	return "idrac_power_control_interval_in_minutes{ip=\"" + collector.client.host + "\", id=\"" + id + "\", name=\"" + name + "\"} " + strconv.Itoa(interval)
}

func (collector *Collector) NewSelEntry(id string, message string, component string, severity string, created time.Time) string {
	return "idrac_sel_entry{ip=\"" + collector.client.host + "\", component=\"" + component + "\", id=\"" + id + "\", message=\"" + message + "\", severity=\"" + severity + "\", time=\"" + created.String() + "\"} " + strconv.Itoa(int(created.Unix()))
}

func (collector *Collector) NewDriveInfo(id, name, manufacturer, model, serial, mediatype, protocol string, slot int) string {
	//var slotstr string
	//
	//if slot < 0 {
	//	slotstr = ""
	//} else {
	//	slotstr = fmt.Sprint(slot)
	//}

	return "idrac_drive_info{ip=\"" + collector.client.host + "\", id=\"" + id + "\", manufacturer=\"" + manufacturer + "\", mediatype=\"" + mediatype + "\", model=\"" + model + "\", name=\"" + name + "\", protocol=\"" + protocol + "\", serial=\"" + serial + "\"} 1"
}

func (collector *Collector) NewDriveHealth(id, health string) string {
	value := health2value(health)
	return "idrac_drive_health{ip=\"" + collector.client.host + "\", id=\"" + id + "\", status=\"" + health + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewDriveCapacity(id string, capacity int) string {
	return "idrac_drive_capacity_bytes{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.Itoa(capacity)
}

func (collector *Collector) NewDriveLifeLeft(id string, lifeLeft int) string {
	return "\nidrac_drive_life_left_percent{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.Itoa(lifeLeft)
}

func (collector *Collector) NewMemoryModuleInfo(id, name, manufacturer, memtype, serial, ecc string, rank int) string {
	return "idrac_memory_module_info{ip=\"" + collector.client.host + "\", ecc=\"" + ecc + "\", id=\"" + id + "\", manufacturer=\"" + manufacturer + "\", name=\"" + name + "\", rank=\"" + strconv.Itoa(rank) + "\", serial=\"" + serial + "\", type=\"" + memtype + "\"} 1"
}

func (collector *Collector) NewMemoryModuleHealth(id, health string) string {
	value := health2value(health)
	return "idrac_memory_module_health{ip=\"" + collector.client.host + "\", id=\"" + id + "\", status=\"" + health + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewMemoryModuleCapacity(id string, capacity int) string {
	return "idrac_memory_module_capacity_bytes{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.Itoa(capacity)
}

func (collector *Collector) NewMemoryModuleSpeed(id string, speed int) string {
	return "idrac_memory_module_speed_mhz{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.Itoa(speed)
}

func (collector *Collector) NewNetworkInterfaceHealth(id, health string) string {
	value := health2value(health)
	return "idrac_network_interface_health{ip=\"" + collector.client.host + "\", id=\"" + id + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewNetworkPortHealth(iface, id, health string) string {
	value := health2value(health)
	return "idrac_network_port_health{ip=\"" + collector.client.host + "\", id=\"" + id + "\", interface=\"" + iface + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}

func (collector *Collector) NewNetworkPortSpeed(iface, id string, speed int) string {
	return "idrac_network_port_speed_mbps{ip=\"" + collector.client.host + "\", id=\"" + id + "\", interface=\"" + iface + "\"} " + strconv.Itoa(speed)
}

func (collector *Collector) NewNetworkPortLinkUp(iface, id, status string) string {
	value := linkstatus2value(status)
	return "idrac_network_port_link_up{ip=\"" + collector.client.host + "\", id=\"" + id + "\", interface=\"" + iface + "\", status=\"" + status + "\"} " + strconv.FormatFloat(value, 'f', 0, 64)
}
