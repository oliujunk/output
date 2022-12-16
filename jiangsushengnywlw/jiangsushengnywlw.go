package jiangsushengnywlw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"
)

var (
	token   string
	devices []xphapi.Device
)

func Start() {
	log.Println("江苏省农业物联网平台 start ------")
	updateToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	//_, _ = job.AddFunc("0 0 */1 * * *", sendData)
	_, _ = job.AddFunc("0 */1 * * * *", sendData)
	job.Start()
}

func updateToken() {
	token = xphapi.RNGetToken("Y810066", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("Y810066", token)
}

func sendData() {
	for _, device := range devices {
		if len(device.DeviceRemark) <= 0 {
			continue
		}
		resp, err := http.Get("http://101.34.116.221:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
		if err != nil {
			log.Println("获取数据异常")
			continue
		}
		result, _ := ioutil.ReadAll(resp.Body)
		var dataEntity xphapi.DataEntity
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"deviceId":"%d",`, device.DeviceID))
				build.WriteString(fmt.Sprintf(`"sessionKey":"%s",`, "121345"))
				build.WriteString(fmt.Sprintf(`"data":{`))
				for _, entity := range dataEntity.Entity {
					build.WriteString(fmt.Sprintf(`"%s":%s,`, entity.EName, entity.EValue))
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `}}`
				log.Println(message)
				req, err := http.NewRequest("POST", "http://210.12.220.220:8012/dataswitch-sq/sendApi/sendSingleInfo", bytes.NewBuffer([]byte(message)))
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

				result, _ := ioutil.ReadAll(resp.Body)
				log.Println(string(result))
			}
			time.Sleep(1 * time.Second)
		}
	}
}
