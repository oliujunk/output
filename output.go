package main

import (
	"log"
	"oliujunk/output/bjhlc"
	"oliujunk/output/hangjingqi"
	"oliujunk/output/houtuyun"
	"oliujunk/output/jiamusi"
	"oliujunk/output/jiangsushengnywlw"
	"oliujunk/output/ruinong_houtuyun"
	"oliujunk/output/shangma"
	"oliujunk/output/soilmoisture"
	"oliujunk/output/taicangnywlw"
	"oliujunk/output/zhongrui"
	"oliujunk/output/zhongyaocai"
	"oliujunk/output/zjtpyun"
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

	//houtuyun_t.Start()

	shangma.Start()

	zjtpyun.Start()

	zhongyaocai.Start()

	jiangsushengnywlw.Start()

	taicangnywlw.Start()

	bjhlc.Start()

	ruinong_houtuyun.Start()

	select {}
}
