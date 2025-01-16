package bjhlc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"
)

// 北京华隆辰

var (
	token   string
	devices []xphapi.Device
	//prodUrl = "https://sco.pipechina.com.cn:8443"
	prodUrl = "https://xcsco.pipechina.com.cn"
)

func updateXphToken() {
	token = xphapi.RNGetToken("Y810166", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("Y810166", token)
}

func Start() {
	// 全国土壤墒情平台
	log.Println("北京华隆辰数据推送 start ------")
	updateXphToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateXphToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	//_, _ = job.AddFunc("0 */30 * * * *", sendData)
	_, _ = job.AddFunc("0 */1 * * * *", sendData)

	job.Start()
}

func sendData() {
	for _, device := range devices {
		resp, err := http.Get("http://101.34.116.221:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
		if err != nil {
			log.Println("获取数据异常")
			return
		}
		result, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		dataEntity := xphapi.DataEntity{}
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)

			var build strings.Builder
			build.WriteString(`{`)
			build.WriteString(fmt.Sprintf(`"data":{`))
			for _, entity := range dataEntity.Entity {
				build.WriteString(fmt.Sprintf(`"%s":%s,`, entity.EName, entity.EValue))
			}
			message := build.String()
			message = strings.TrimRight(message, ",")
			message = message + `}`
			message = message + fmt.Sprintf(`,"method":"dataPost"`)
			message = message + fmt.Sprintf(`,"deviceId":"%d"`, device.DeviceID)
			message = message + fmt.Sprintf(`,"deviceName":"%s"`, device.DeviceName)
			message = message + fmt.Sprintf(`,"time":"%s"`, dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {
				message = message + fmt.Sprintf(`,"online":%t`, true)
			} else {
				message = message + fmt.Sprintf(`,"online":%t`, false)
			}
			message = message + `}`

			log.Println(message)
			req, err := http.NewRequest("POST", prodUrl+"/pims/prod-api/business/badWeather/badWeatherPoint/point/listen", bytes.NewBuffer([]byte(message)))
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

			req, err = http.NewRequest("POST", prodUrl+"/vueiot/roma/environmentStation/syncEnvirStationDevice", bytes.NewBuffer([]byte(message)))
			if err != nil {
				log.Println(err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)
			if err != nil {
				log.Println(err)
				continue
			}

			result, _ = io.ReadAll(resp.Body)
			log.Println(string(result))

			time.Sleep(1 * time.Second)
		}
	}
}
