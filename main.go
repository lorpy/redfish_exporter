package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"regexp"
	"server_exporter/collector"
	"server_exporter/config"
	"server_exporter/tools"
)

var (
	Dict       string
	Prometheus string
	Conf       string
)

func main() {
	flag.StringVar(&Dict, "dict", "", "the map file of name and ip")
	flag.StringVar(&Conf, "conf", "", "")
	flag.StringVar(&Prometheus, "prometheus", "", "prometheus server")
	flag.Parse()

	client := gin.Default()
	var data tools.Data
	conf := config.Init(Conf)
	if conf.Basic.Port == "" {
		conf.Basic.Port = "9234"
	}
	client.GET("/metrics", func(c *gin.Context) {
		target := c.Query("target")
		if target == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "param 'target' is required"})
			return
		}
		isValid, _ := regexp.MatchString("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$", target)
		if !isValid {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "param 'target' is invalid"})
			return
		}

		for ip, auth := range conf.Hosts {
			if ip == target {
				data.Dat.Host = target
				data.Dat.Account = auth.Username
				data.Dat.Password = auth.Password
			}
		}

		deviceName := config.GetMapOfNameAndIp(Dict, target)

		if data.Dat.Account == "" || data.Dat.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "params 'account' or 'password' are empty"})
			return
		}
		ch := make(chan string, 100)
		go func() {
			collectorClient := collector.NewCollector(data, deviceName, conf, ch)
			collectorClient.Collect()
			close(ch)
		}()
		for {
			msg, ok := <-ch
			if !ok {
				log.Println("channel closed")
				return
			}
			if Prometheus != "" {
				tools.RemoteWrite(msg, Prometheus)
			} else {
				c.String(http.StatusOK, "%s\n", msg)
			}
		}
	})
	log.Fatal(client.Run(conf.Basic.BindIp + ":" + conf.Basic.Port))
}
