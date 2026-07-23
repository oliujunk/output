package main

import (
	"log"
	"oliujunk/output/lianshi"
)

func init() {
	// 日志信息添加文件名行号
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {

	lianshi.Start()

	select {}
}
