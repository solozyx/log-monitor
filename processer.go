package main

import (
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// 2006-01-02 15:04:05 +0000
	GoTimeParseSeed = "02/Jan/2006:15:04:05 +0000"
)

type LogProcess struct {
	// 读取数据channel ReadFromFile->Process
	rc chan []byte
	// 写入数据channel Process->WriteToInfluxDBDns
	wc     chan *Message
	reader Reader
	writer Writer
}

type Message struct {
	TimeLocal time.Time
	// 流量
	BytesSent                    int
	Method, Path, Scheme, Status string
	UpstreamTime, RequestTime    float64
}

/**
172.0.0.12 - - [04/Mar/2018:13:49:52 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
*/
func (l *LogProcess) Process() {
	r := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)
	loc, _ := time.LoadLocation("Asia/Shanghai")

	// 消费 read channel 获取逐行日志数据
	for v := range l.rc {
		// 正则13个括号 把所有匹配的括号规则数据返回出来
		ret := r.FindStringSubmatch(string(v))
		//
		if len(ret) != 14 {
			TypeMonitorChan <- TypeErrNum
			log.Println("processor FindStringSubmatch fail:", string(v))
			continue
		}

		message := &Message{}
		// 04/Mar/2018:13:49:52 +0000
		t, err := time.ParseInLocation(GoTimeParseSeed, ret[4], loc)
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			log.Println("processor ParseInLocation fail:", err.Error(), ret[4])
			continue
		}
		message.TimeLocal = t
		// 2133
		byteSent, _ := strconv.Atoi(ret[8])
		message.BytesSent = byteSent
		// GET /foo?query=t HTTP/1.0
		reqSlice := strings.Split(ret[6], " ")
		if len(reqSlice) != 3 {
			TypeMonitorChan <- TypeErrNum
			log.Println("processor strings.Split fail", ret[6])
			continue
		}
		// GET
		message.Method = reqSlice[0]
		// /foo
		u, err := url.Parse(reqSlice[1])
		if err != nil {
			log.Println("processor url parse fail:", err)
			TypeMonitorChan <- TypeErrNum
			continue
		}
		message.Path = u.Path
		// http
		message.Scheme = ret[5]
		// 200
		message.Status = ret[7]
		// 1.005
		upstreamTime, _ := strconv.ParseFloat(ret[12], 64)
		message.UpstreamTime = upstreamTime
		// 1.854
		requestTime, _ := strconv.ParseFloat(ret[13], 64)
		message.RequestTime = requestTime
		// 读取日志文件1行内容 投递 wc channel
		l.wc <- message
	}
}
