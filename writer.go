package main

import (
	"log"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
)

type Writer interface {
	Write(wc chan *Message)
}

type WriteToInfluxDB struct {
	// influx data source
	influxdbDsn string
}

func (w *WriteToInfluxDB) Write(wc chan *Message) {
	// Create a new HTTPClient
	influxdbConf := strings.Split(w.influxdbDsn, "@")
	influxdbClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxdbConf[0],
		Username: influxdbConf[1],
		Password: influxdbConf[2],
	})
	if err != nil {
		log.Fatal(err)
	}

	for v := range wc {
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database: influxdbConf[3],
			// 精度 秒 s
			Precision: influxdbConf[4],
		})
		if err != nil {
			log.Fatal(err)
		}
		// Tags: Path, Method, Scheme, Status
		tags := map[string]string{
			"Path":   v.Path,
			"Method": v.Method,
			"Scheme": v.Scheme,
			"Status": v.Status,
		}
		// Fields: UpstreamTime, RequestTime, BytesSent
		fields := map[string]interface{}{
			"UpstreamTime": v.UpstreamTime,
			"RequestTime":  v.RequestTime,
			"BytesSent":    v.BytesSent,
		}
		// Create a point and add to batch
		// 表名nginx_log
		pt, err := client.NewPoint("nginx_log", tags, fields, v.TimeLocal)
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		// Write the batch
		if err := influxdbClient.Write(bp); err != nil {
			log.Fatal(err)
		}

		log.Println("write Message to influxdb success!")
	}
}
