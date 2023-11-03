package taicangnywlw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
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
	thirdToken string
	devices    []xphapi.Device
	pests      []xphapi.Pest
	pestName   = [...]string{"airHumidity", "airTemp", "lightIntensity", "lan", "lat", "BugNumOne"}
)

func Start() {
	log.Println("太仓物联网平台 start ------")
	updateToken()
	updateThirdToken()
	updateDevices()
	updatePests()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0/12 * * *", updateThirdToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	_, _ = job.AddFunc("0 0 0/1 * * *", updatePests)
	//_, _ = job.AddFunc("0 */5 * * * *", sendData)
	_, _ = job.AddFunc("0 */30 * * * *", sendData)
	job.Start()
}

func updateToken() {
	token = xphapi.RNGetToken("jiangsushengnywlw", "88888888")
}

func updateThirdToken() {
	urlValue := url.Values{}
	urlValue.Add("username", "TCnync")
	urlValue.Add("password", "TCnync@123")
	payload := strings.NewReader(urlValue.Encode())
	req, err := http.NewRequest("POST", "http://180.108.205.73:8001/iotManager/app/iotuser/login", payload)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	buf := bytes.NewBuffer(result)
	message, err := simplejson.NewFromReader(buf)
	if err != nil {
		log.Println(err)
		return
	}
	token, _ := message.Get("token").String()
	thirdToken = token
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
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"data":{`))
				build.WriteString(fmt.Sprintf(`"deviceid":"%d",`, device.DeviceID))
				for index, entity := range dataEntity.Entity {
					build.WriteString(fmt.Sprintf(`"%s":%s,`, names[index], entity.EValue))
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `},`
				message = message + fmt.Sprintf(`"deviceid":"%d"`, device.DeviceID)
				message = message + `}`

				log.Println(message)
				req, err := http.NewRequest("POST", "http://180.108.205.73:8001/iotManager/receiveNormal/statusReceive", bytes.NewBuffer([]byte(message)))
				if err != nil {
					log.Println(err)
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization-iot", thirdToken)
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
		result, _ := ioutil.ReadAll(resp.Body)
		var dataEntity xphapi.DataEntity
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {

			resp1, err1 := http.Get("http://101.34.116.221:8005/intfa/queryData/68273394")
			if err1 != nil {
				log.Println("获取数据异常")
				continue
			}
			result1, _ := ioutil.ReadAll(resp1.Body)
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

				result, _ := ioutil.ReadAll(resp.Body)
				log.Println(string(result))
			}
			time.Sleep(1 * time.Second)
		}
	}

}
