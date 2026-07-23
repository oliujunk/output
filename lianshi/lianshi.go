package lianshi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"oliujunk/output/xphapi"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

var (
	token101   string
	devices101 []xphapi.Device
)

func updateToken() {
	token101 = xphapi.RNGetToken("Y810360", "88888888")
}

func updateDevices() {
	devices101 = xphapi.RNGetDevices("Y810360", token101)
}

func Start() {
	log.Println("仪征大屏数据 start ------")

	updateToken()

	updateDevices()

	sendData()

	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	_, _ = job.AddFunc("0 */30 * * * *", sendData)

	job.Start()
}

func sendData() {
	now := time.Now()

	for _, device := range devices101 {
		if len(device.DeviceRemark) <= 0 {
			continue
		}
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
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {
				for _, entity := range dataEntity.Entity {
					params := url.Values{}
					params.Add("area", device.DeviceRemark)
					params.Add("sensor", entity.EName)
					params.Add("value", entity.EValue)
					req, err := http.NewRequest("POST", "http://shk.sinoagridb.com/MonitorServer/sensor/areaSoil/upload?"+params.Encode(), nil)
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
					time.Sleep(1 * time.Second)
				}
			}
		}
	}
}
