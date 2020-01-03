package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

type Reader interface {
	Read(rc chan []byte)
}

type ReadFromFile struct {
	// 读取日志文件的路径
	path string
}

// func (r *ReadFromFile) Read(rc chan string) {
func (r *ReadFromFile) Read(rc chan []byte) {
	// TODO : 生产环境 日志按 天 周 切割
	//  打开新的日志文件,重新给 f rd 赋值
	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error:%s", err.Error()))
	}
	// 把文件的字符指针移动到文件末尾 从文件末尾开始逐行读取文件内容
	// Seek(0,2) and 2 means relative to the end
	f.Seek(0, 2)
	// f 只能逐个读取字符 不能逐行读取 使用bufio包
	rd := bufio.NewReader(f)

	for {
		// 逐行读取文件内容 直到遇到 \n 为止 即读取1行内容
		line, err := rd.ReadBytes('\n')
		// 读取到文件末尾返回的err 应该继续读取 读取太快日志文件还没有增加
		if err == io.EOF {
			// 等待新日志行产生
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("ReadBytes error:%s", err.Error()))
		}
		// 读取到1行日志数据
		TypeMonitorChan <- TypeHandleLine
		// 去掉读取内容最后的\n符 投递channel
		rc <- line[:len(line)-1]
	}
}
