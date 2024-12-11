package collector

import (
	"log"
	"server_exporter/config"
	"server_exporter/tools"
	"sync"
)

type Collector struct {
	client *Client
	config config.Config
	ch     chan string
}

func NewCollector(authInfo tools.Data, deviceName string, conf config.Config, ch chan string) *Collector {
	client, err := NewClient(authInfo, deviceName)
	if err != nil {
		log.Fatalf("NewClient err: %v", err)
		return nil
	}

	return &Collector{
		client: client,
		config: conf,
		ch:     ch,
	}
}

func (collector *Collector) Collect() {
	var wg sync.WaitGroup

	if collector.config.Metrics.System {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshSystem(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect system metrics-%s", err)
			}
			wg.Done()
		}()
	}

	if collector.config.Metrics.Sensors {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshSensors(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect sensors metrics-%s", err)
			}
			wg.Done()
		}()
	}

	if collector.config.Metrics.Power {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshPower(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect power metrics-%s", err)
			}
			wg.Done()
		}()
	}

	if collector.config.Metrics.Network {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshNetwork(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect network metrics-%s", err)
			}
			wg.Done()
		}()
	}

	if collector.config.Metrics.Sel {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshIdracSel(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect sel metrics-%s", err)
			}
			wg.Done()
		}()
	}

	if collector.config.Metrics.Storage {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshStorage(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect storage metrics-%s", err)
			}
			wg.Done()
		}()
	}

	if collector.config.Metrics.Memory {
		wg.Add(1)
		go func() {
			err := collector.client.RefreshMemory(collector, collector.ch)
			if err != nil {
				log.Printf("fail to collect memory metrics-%s", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

}
