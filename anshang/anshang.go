package anshang

import (
	"bytes"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"time"
)

var (
	token string
)

func Start() {
	log.Println("安商平台控制测试 start ------")
	updateToken()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 * * * * *", relayOpen)
	_, _ = job.AddFunc("30 * * * * *", relayClose)
	job.Start()
}

func updateToken() {
	token = xphapi.RNGetToken("jiangsushengnywlw", "88888888")
}

func relayOpen() {
	control(72970413, 0, 1)
}

func relayClose() {
	control(72970413, 0, 0)
}

func control(deviceID int, index int, state int) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	loginParam := map[string]int{"deviceId": deviceID, "relayNum": index, "relayState": state}
	jsonStr, _ := json.Marshal(loginParam)
	log.Println(string(jsonStr))
	request, err := http.NewRequest("POST", "http://121.40.59.50:8005/relay", bytes.NewBuffer(jsonStr))
	request.Header.Add("token", token)
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}
	result, _ := io.ReadAll(resp.Body)
	log.Println(string(result))
	parseBool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false
	}
	return parseBool
}
