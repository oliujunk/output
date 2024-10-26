package zhiling

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
	token   string
	devices []xphapi.Device
)

func Start() {
	log.Println("智菱 start ------")
	updateToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	_, _ = job.AddFunc("0 */2 * * * *", sendData)
	job.Start()
}

func updateToken() {
	token = xphapi.RNGetToken("zhiling", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("zhiling", token)
}

func sendData() {
	for _, device := range devices {
		resp, err := http.Get("http://101.34.116.221:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
		if err != nil {
			log.Println("获取数据异常")
			continue
		}
		result, _ := io.ReadAll(resp.Body)
		var dataEntity xphapi.DataEntity
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {
				log.Println(dataEntity)
				req, err := http.NewRequest("POST", "http://szxc.zillion-ioe.com/iot-web/api/platform/realtime-data/"+strconv.Itoa(device.DeviceID), bytes.NewBuffer(result))
				if err != nil {
					log.Println(err)
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					log.Println(err)
					continue
				}

				result, _ := io.ReadAll(resp.Body)
				log.Println(string(result))
			}
			time.Sleep(1 * time.Second)
		}
	}
}
