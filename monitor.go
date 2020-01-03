package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// 系统状态监控
type SystemInfo struct {
	HandleLine   int     `json:"handleLine"`   // 总处理日志行数
	Tps          float64 `json:"tps"`          // 系统吞出量
	ReadChanLen  int     `json:"readChanLen"`  // read channel 长度
	WriteChanLen int     `json:"writeChanLen"` // write channel 长度
	RunTime      string  `json:"runTime"`      // 运行总时间
	ErrNum       int     `json:"errNum"`       // 错误数
}

const (
	TypeHandleLine = 0
	TypeErrNum     = 1
)

var (
	TypeMonitorChan = make(chan int, 200)
	MonitorPort     = ":9193"
)

type Monitor struct {
	// 开始运行时间
	startTime time.Time
	data      SystemInfo
	tpsSlice  []int
}

func (m *Monitor) start(lp *LogProcess) {
	// 日志处理 处理行数统计 错误数统计
	go func() {
		for n := range TypeMonitorChan {
			switch n {
			case TypeErrNum:
				m.data.ErrNum += 1
			case TypeHandleLine:
				m.data.HandleLine += 1
			}
		}
	}()
	// TPS 计算
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			// 每5秒 执行1次计算代码
			<-ticker.C
			// 追加当前处理日志行数
			m.tpsSlice = append(m.tpsSlice, m.data.HandleLine)
			if len(m.tpsSlice) > 2 {
				// 始终保持slice只保持2个数据
				m.tpsSlice = m.tpsSlice[1:]
			}
		}
	}()

	http.HandleFunc("/monitor", func(writer http.ResponseWriter, request *http.Request) {
		m.data.RunTime = time.Now().Sub(m.startTime).String()
		m.data.ReadChanLen = len(lp.rc)
		m.data.WriteChanLen = len(lp.wc)
		// TPS 计算
		if len(m.tpsSlice) >= 2 {
			m.data.Tps = float64(m.tpsSlice[1]-m.tpsSlice[0]) / 5
		}
		// 带缩进 Indent
		ret, _ := json.MarshalIndent(m.data, "", "\t")
		// 向 writer 输出string
		io.WriteString(writer, string(ret))
	})
	// 阻塞 程序持续运行
	http.ListenAndServe(MonitorPort, nil)
}
