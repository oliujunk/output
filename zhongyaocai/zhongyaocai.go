package zhongyaocai

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"
)

var (
	token      string
	devices    []xphapi.Device
	app_id     = "d70b66ea3e63ff35585b5c9d801sadfsdf7" // 正式
	app_secret = "a52b2fb2bf6264e0b7920c4cce4fssdf102"
	//app_id     = "55141b3bd61ca0970a7fc9e6b3f3c291"			// 测试
	//app_secret = "a35f93feb4d4fa94236c6884653e1752"
)

func updateToken() {
	token = xphapi.RNGetToken("Y810108", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("Y810108", token)
}

func Start() {
	log.Println("中草药平台推送 start ------")
	updateToken()
	updateDevices()
	//sendData()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("30 0 0 */1 * *", updateDevices)
	_, _ = job.AddFunc("0 0 */1 * * *", sendData)
	//_, _ = job.AddFunc("0 */1 * * * *", sendData)

	job.Start()
}

func sendData() {
	for _, device := range devices {
		resp, err := http.Get("http://101.34.116.221:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
		if err != nil {
			log.Println("获取数据异常")
			return
		}
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		dataEntity := xphapi.DataEntity{}
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {
			now := time.Now()
			timestamp := time.Now().Format("20060102150405")
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {
				params := url.Values{}
				params.Set("atmos", dataEntity.Entity[4].EValue)
				//params.Set("base", "825")	// 测试
				params.Set("base", "2550") // 正式
				params.Set("data_date", dataEntity.Entity[0].Datetime)
				params.Set("device_code", "64470577")
				params.Set("ec_val", "0")
				params.Set("gateway_node", "1")
				params.Set("humidity", dataEntity.Entity[3].EValue)
				params.Set("illumination", "0")
				params.Set("rainfall", dataEntity.Entity[6].EValue)
				params.Set("soil_humidity", dataEntity.Entity[8].EValue)
				params.Set("soil_ph", "0")
				params.Set("soil_temperature", dataEntity.Entity[7].EValue)
				params.Set("sunshine_hour", "0")
				params.Set("temperature", dataEntity.Entity[2].EValue)
				params.Set("wind_speed", dataEntity.Entity[0].EValue)
				params.Set("wind_direction", dataEntity.Entity[1].EValue)
				payload := strings.NewReader(params.Encode())

				urlValue := url.Values{}
				Url, err := url.Parse("https://sy.nrc.ac.cn/open-api/v1/basic/env/")
				//Url, err := url.Parse("http://sy.nrc.ac.cn:8100/open-api/v1/basic/env/")
				if err != nil {
					return
				}
				urlValue.Set("app_id", app_id)
				urlValue.Set("timestamp", timestamp)

				var build strings.Builder
				build.WriteString(fmt.Sprintf(`app_id=%s&`, app_id))
				build.WriteString(fmt.Sprintf(`atmos=%s&`, dataEntity.Entity[4].EValue))
				build.WriteString(fmt.Sprintf(`base=%s&`, "2550"))
				build.WriteString(fmt.Sprintf(`data_date=%s&`, dataEntity.Entity[0].Datetime))
				build.WriteString(fmt.Sprintf(`device_code=%s&`, "64470577"))
				build.WriteString(fmt.Sprintf(`ec_val=%s&`, "0"))
				build.WriteString(fmt.Sprintf(`gateway_node=%s&`, "1"))
				build.WriteString(fmt.Sprintf(`humidity=%s&`, dataEntity.Entity[3].EValue))
				build.WriteString(fmt.Sprintf(`illumination=%s&`, "0"))
				build.WriteString(fmt.Sprintf(`rainfall=%s&`, dataEntity.Entity[6].EValue))
				build.WriteString(fmt.Sprintf(`soil_humidity=%s&`, dataEntity.Entity[8].EValue))
				build.WriteString(fmt.Sprintf(`soil_ph=%s&`, "0"))
				build.WriteString(fmt.Sprintf(`soil_temperature=%s&`, dataEntity.Entity[7].EValue))
				build.WriteString(fmt.Sprintf(`sunshine_hour=%s&`, "0"))
				build.WriteString(fmt.Sprintf(`temperature=%s&`, dataEntity.Entity[2].EValue))
				build.WriteString(fmt.Sprintf(`timestamp=%s&`, timestamp))
				build.WriteString(fmt.Sprintf(`wind_direction=%s&`, dataEntity.Entity[1].EValue))
				build.WriteString(fmt.Sprintf(`wind_speed=%s`, dataEntity.Entity[0].EValue))

				signContent := build.String() + app_secret
				log.Println(signContent)
				signCode := md5.Sum([]byte(signContent))
				sign := hex.EncodeToString(signCode[:])
				urlValue.Set("sign", strings.ToUpper(sign))

				Url.RawQuery = urlValue.Encode()
				urlPath := Url.String()
				req, err := http.NewRequest("POST", urlPath, payload)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(urlValue.Encode())
				log.Println(params.Encode())
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				client := &http.Client{}
				resp, err := client.Do(req)
				res, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(string(res))
			}
			time.Sleep(1 * time.Second)
		}
	}
}
