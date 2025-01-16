package main

import (
	"log"
	"oliujunk/output/bjhlc"
)

func init() {
	// 日志信息添加文件名行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {

	//soilmoisture.Start()
	//
	//jiamusi.Start()
	//
	//hangjingqi.Start()
	//
	//zhongrui.Start()
	//
	//houtuyun.Start()
	//
	////houtuyun_t.Start()
	//
	//shangma.Start()
	//
	//zjtpyun.Start()
	//
	//zhongyaocai.Start()
	//
	//jiangsushengnywlw.Start()
	//
	//taicangnywlw.Start()
	//
	//
	//
	//ruinong_houtuyun.Start()
	//
	//zhiling.Start()

	bjhlc.Start()

	select {}
}
