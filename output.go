package main

import (
	"log"
	"oliujunk/output/jiangsushengnywlw"
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

	jiangsushengnywlw.Start()

	select {}
}
