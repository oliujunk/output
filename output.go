package main

import (
	"log"
	"oliujunk/output/hangjingqi"
	"oliujunk/output/houtuyun"
	"oliujunk/output/jiamusi"
	"oliujunk/output/soilmoisture"
	"oliujunk/output/zhongrui"
)

func init() {
	// 日志信息添加文件名行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {

	soilmoisture.Start()

	jiamusi.Start()

	hangjingqi.Start()

	zhongrui.Start()

	houtuyun.Start()

	select {}
}
