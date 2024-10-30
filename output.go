package main

import (
	"log"
	"oliujunk/output/yunfei"
	yunfei_dashuju "oliujunk/output/yunfei-dashuju"
)

func init() {
	// 日志信息添加文件名行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {

	yunfei.Start()
	yunfei_dashuju.Start()

	select {}
}
