package wssw

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
	log.Println("wssw平台推送 start ------")
	updateToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	//_, _ = job.AddFunc("0 */1 * * * *", sendData)
	_, _ = job.AddFunc("0 */30 * * * *", sendData)
	job.Start()
}

func updateToken() {
	token = xphapi.GetToken121("wssw", "123456")
}

func updateDevices() {
	devices = xphapi.GetDevices121("wssw", token)
}

func sendData() {
	for _, device := range devices {
		resp, err := http.Get("http://121.43.37.41:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
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
				build.WriteString(fmt.Sprintf(`"requestBody":{`))
				build.WriteString(fmt.Sprintf(`"ID":"%d",`, device.DeviceID))
				for _, entity := range dataEntity.Entity {
					build.WriteString(fmt.Sprintf(`"%s":"%s",`, entity.EName, entity.EValue))
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `}`
				message = message + `}`

				log.Println(message)
				req, err := http.NewRequest("POST", "http://47.92.144.241:81/deviceData/putData", bytes.NewBuffer([]byte(message)))
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
