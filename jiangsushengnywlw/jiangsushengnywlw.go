package jiangsushengnywlw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"
)

var (
	token    string
	devices  []xphapi.Device
	pests    []xphapi.Pest
	pestName = [...]string{"airHumidity", "airTemp", "lightIntensity", "lan", "lat", "BugNumOne"}
)

func Start() {
	log.Println("江苏省农业物联网平台 start ------")
	updateToken()
	updateDevices()
	updatePests()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	_, _ = job.AddFunc("0 0 0/1 * * *", updatePests)
	//_, _ = job.AddFunc("0 */1 * * * *", sendData)
	_, _ = job.AddFunc("0 */30 * * * *", sendData)
	job.Start()
}

func updateToken() {
	token = xphapi.RNGetToken("jiangsushengnywlw", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("jiangsushengnywlw", token)
}

func updatePests() {
	pests = xphapi.RNGetPests("jiangsushengnywlw", token)
}

func sendData() {
	for _, device := range devices {
		if len(device.DeviceRemark) <= 0 {
			continue
		}

		if device.DeviceID == 59141938 {
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
				names := strings.Split(device.ElementExtendName, "/")
				if len(names) < 2 {
					continue
				}
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"deviceId":"%d",`, device.DeviceID))
				build.WriteString(fmt.Sprintf(`"sessionKey":"%s",`, "121345"))
				build.WriteString(fmt.Sprintf(`"data":{`))
				for index, entity := range dataEntity.Entity {
					build.WriteString(fmt.Sprintf(`"%s":%s,`, names[index], entity.EValue))
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

	for _, pest := range pests {
		if len(pest.DeviceRemark) <= 0 {
			continue
		}
		resp, err := http.Get("http://101.34.116.221:8005/pest/intfa/queryData/" + pest.DeviceID)
		if err != nil {
			log.Println("获取数据异常")
			continue
		}
		result, _ := io.ReadAll(resp.Body)
		var dataEntity xphapi.DataEntity
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {

			resp1, err1 := http.Get("http://101.34.116.221:8005/intfa/queryData/68268113")
			if err1 != nil {
				log.Println("获取数据异常")
				continue
			}
			result1, _ := io.ReadAll(resp1.Body)
			var dataEntity1 xphapi.DataEntity
			_ = json.Unmarshal(result1, &dataEntity1)

			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"deviceId":"%s",`, pest.DeviceID))
				build.WriteString(fmt.Sprintf(`"sessionKey":"%s",`, "121345"))
				build.WriteString(fmt.Sprintf(`"data":{`))
				for index, entity := range dataEntity.Entity {
					if index == 0 {
						build.WriteString(fmt.Sprintf(`"%s":%s,`, pestName[index], dataEntity1.Entity[3].EValue))
					} else if index == 1 {
						build.WriteString(fmt.Sprintf(`"%s":%s,`, pestName[index], dataEntity1.Entity[2].EValue))
					} else {
						build.WriteString(fmt.Sprintf(`"%s":%s,`, pestName[index], entity.EValue))
					}

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

				result, _ := io.ReadAll(resp.Body)
				log.Println(string(result))
			}
			time.Sleep(1 * time.Second)
		}
	}

}
