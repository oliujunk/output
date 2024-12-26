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
	_, _ = job.AddFunc("0/20 * * * * *", relayOpen)
	_, _ = job.AddFunc("10/20 * * * * *", relayClose)
	job.Start()
}

func updateToken() {
	token = xphapi.GetTokenAnshang("test2", "123456")
}

func relayOpen() {
	control(72970413, 0, 1)
}

func relayClose() {
	control(72970413, 0, 0)
}

func control(deviceID int, index int, state int) bool {
	resp, err := http.Get("http://121.40.59.50:8005/intfa/queryData/72970413")
	if err != nil {
		return false
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	dataEntity := xphapi.DataEntity{}
	_ = json.Unmarshal(result, &dataEntity)

	if dataEntity.Online {
		client := &http.Client{Timeout: 10 * time.Second}
		loginParam := map[string]int{"deviceId": deviceID, "relayNum": index, "relayState": state}
		jsonStr, _ := json.Marshal(loginParam)
		log.Println(string(jsonStr))
		request, err := http.NewRequest("POST", "http://121.40.59.50:8005/relay", bytes.NewBuffer(jsonStr))
		request.Header.Add("token", token)
		request.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(request)
		if err != nil {
			log.Println(err)
		}
		result, _ = io.ReadAll(resp.Body)
		log.Println(string(result))
		parseBool, err := strconv.ParseBool(string(result))
		if err != nil {
			return false
		}
		return parseBool
	}
	return false
}
