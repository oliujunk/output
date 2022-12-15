package jiangsushengnywlw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
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
	logrus.Info("江苏省农业物联网平台 start ------")
	updateToken()
	updateDevices()
	c := cron.New()
	_ = c.AddFunc("0 0 0/12 * * *", updateToken)
	_ = c.AddFunc("0 0 0/1 * * *", updateDevices)
	_ = c.AddFunc("0 0 */1 * * *", sendData)
	//_ = c.AddFunc("0 */1 * * * *", sendData)
	c.Start()
}

func updateToken() {
	token = xphapi.NewGetToken("4691615", "123456")
}

func updateDevices() {
	devices = xphapi.NewGetDevices("4691615", token)
	devices1 := xphapi.NewGetDevices("5270415", token)
	devices = append(devices, devices1...)
}

func sendData() {
	for _, device := range devices {
		if len(device.DeviceRemark) <= 0 {
			continue
		}
		resp, err := http.Get("http://47.105.215.208:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
		if err != nil {
			logrus.Error("获取数据异常")
			continue
		}
		result, _ := ioutil.ReadAll(resp.Body)
		var dataEntity xphapi.DataEntity
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Minute * 6)) {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"deviceId":"%s",`, device.DeviceRemark))
				build.WriteString(fmt.Sprintf(`"sessionKey":"%s",`, device.DeviceRemark))
				build.WriteString(fmt.Sprintf(`"data":{`))
				for _, entity := range dataEntity.Entity {
					build.WriteString(fmt.Sprintf(`"%s":%s,`, entity.EName, entity.EValue))
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `}}`
				logrus.Info(message)
				req, err := http.NewRequest("POST", "http://210.12.220.220:8012/dataswitch-sq/sendApi/sendSingleInfo", bytes.NewBuffer([]byte(message)))
				if err != nil {
					logrus.Error(err)
				}
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					logrus.Error(err)
				}

				result, _ := ioutil.ReadAll(resp.Body)
				logrus.Info(string(result))
			}
			time.Sleep(1 * time.Second)
		}
	}
}
