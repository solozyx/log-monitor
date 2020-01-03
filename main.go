package main

import (
	"flag"
	"time"
)

func main() {
	var logPath, influxdbConf string
	flag.StringVar(&logPath, "path", "./access.log", "read log file path")
	flag.StringVar(&influxdbConf, "influxdbDsn",
		"http://192.168.174.134:8086@user_root@pass_root@db_log_monitor@s",
		"influxdb data source server")
	flag.Parse()

	r := &ReadFromFile{path: logPath}
	w := &WriteToInfluxDB{influxdbDsn: influxdbConf}

	lp := &LogProcess{
		rc:     make(chan []byte, 200),
		wc:     make(chan *Message, 200),
		reader: r,
		writer: w,
	}

	// 读取日志文件
	go lp.reader.Read(lp.rc)
	// 解析日志文件 正则匹配
	for i := 0; i < 2; i++ {
		go lp.Process()
	}
	// 监控指标写入influxdb走网络 相对最慢
	for i := 0; i < 4; i++ {
		go lp.writer.Write(lp.wc)
	}

	m := &Monitor{
		startTime: time.Now(),
		data:      SystemInfo{},
	}
	m.start(lp)
}
