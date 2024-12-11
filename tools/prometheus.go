package tools

import (
	"context"
	"github.com/castai/promwrite"
	"log"
	"regexp"
	"strconv"
	"time"
)

func RemoteWrite(metrics, prometheusServer string) bool {
	data := make([]promwrite.Label, 0)
	// 提取指标名、标签和标签值
	re := regexp.MustCompile(`^(\w+)\{(.+?)\}\s+([\d.]+)$`)
	matches := re.FindStringSubmatch(metrics)
	if len(matches) < 4 {
		log.Println("无法提取指标信息")
		return false
	}

	// 提取指标名、标签部分和指标值
	metricName := matches[1]
	labels := matches[2]
	metricValue := matches[3]
	val, _ := strconv.ParseFloat(metricValue, 0)
	data = append(data, promwrite.Label{Name: "__name__", Value: metricName})

	// 处理tag
	labelRe := regexp.MustCompile(`(\w+)="([^"]+)"`)
	labelMatches := labelRe.FindAllStringSubmatch(labels, -1)
	for _, lm := range labelMatches {
		if len(lm) == 3 {
			data = append(data, promwrite.Label{Name: lm[1], Value: lm[2]})
		}
	}
	client := promwrite.NewClient(prometheusServer + "/api/v1/write")
	_, err := client.Write(context.Background(), &promwrite.WriteRequest{
		TimeSeries: []promwrite.TimeSeries{
			{
				Labels: data,
				Sample: promwrite.Sample{
					Time:  time.Now(),
					Value: val,
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
