package jiaguwen

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"net/url"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"
)

var (
	//devUrl     = "https://xinchang.kf315.net"
	devUrl     = "https://ccy.zjxc.gov.cn"
	username   = "17769856943"
	password   = "Cjm@856943"
	token      string
	thirdToken string
	devices    []xphapi.Device
	pests      []xphapi.Pest
)

func Start() {
	log.Println("甲骨文物联网平台 start ------")
	updateToken()
	updateThirdToken()
	updateDevices()
	updatePests()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 56 0/1 * * *", updateThirdToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	_, _ = job.AddFunc("0 0 0/1 * * *", updatePests)
	_, _ = job.AddFunc("0 */30 * * * *", sendData)
	_, _ = job.AddFunc("0 35 23 */1 * *", sendPestData)
	job.Start()
}

func updateToken() {
	token = xphapi.RNGetToken("jiaguwen", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("jiaguwen", token)
}

func updatePests() {
	pests = xphapi.RNGetPests("jiaguwen", token)
}

func updateThirdToken() {

	signCode := md5.Sum([]byte(password))
	sign := hex.EncodeToString(signCode[:])

	client := &http.Client{Timeout: 5 * time.Second}
	loginParam := map[string]string{"account": username, "password": sign}
	jsonStr, _ := json.Marshal(loginParam)

	resp, err := client.Post(devUrl+"/apiInterface/interface/iot/openApi/getToken", "application/json", bytes.NewBuffer(jsonStr))

	if err != nil {
		log.Println(err)
		return
	}
	result, err := io.ReadAll(resp.Body)
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
	results, _ := message.Get("results").String()
	thirdToken = results
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
			if dataEntity.Online || true {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"deviceId":"%d",`, device.DeviceID))
				build.WriteString(fmt.Sprintf(`"type":1,`))
				build.WriteString(fmt.Sprintf(`"data":{`))

				for _, entity := range dataEntity.Entity {
					build.WriteString(fmt.Sprintf(`"%s":%s,`, entity.EName, entity.EValue))
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `}}`

				log.Println(message)
				req, err := http.NewRequest("POST", devUrl+"/apiInterface/interface/iot/openApi/device/upload", bytes.NewBuffer([]byte(message)))
				if err != nil {
					log.Println(err)
					continue
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("iot-super-token", thirdToken)
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

func sendPestData() {
	for _, pest := range pests {

		params := url.Values{}
		Url, err := url.Parse("http://101.34.116.221:8005/pest/images/" + pest.DeviceID)
		if err != nil {
			return
		}
		now := time.Now()
		params.Set("pageNum", "1")
		params.Set("pageSize", "10")
		params.Set("startTime", now.Add(-24*time.Hour).Format("2006-01-02 15:04:05"))
		params.Set("endTime", now.Format("2006-01-02 15:04:05"))
		Url.RawQuery = params.Encode()
		urlPath := Url.String()
		resp, err := http.Get(urlPath)
		if err != nil {
			log.Println(err)
			continue
		}
		result, _ := io.ReadAll(resp.Body)
		log.Println(string(result))

		buf := bytes.NewBuffer(result)
		message, err := simplejson.NewFromReader(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		imageList, err := message.Array()
		if err != nil {
			log.Println(err)
			continue
		}
		for index, _ := range imageList {
			imageJson := message.GetIndex(index)
			imageUrl, err := imageJson.Get("image").String()
			if err != nil {
				continue
			}
			result, err := imageJson.Get("result").String()
			js, err := simplejson.NewJson([]byte(result))
			if err != nil {
				return
			}
			kind, err := js.Get("kind").Int()
			if err != nil {
				return
			}
			total, err := js.Get("total").Int()
			if err != nil {
				return
			}

			var build strings.Builder
			build.WriteString(`{`)
			build.WriteString(fmt.Sprintf(`"deviceId":"%s",`, pest.DeviceID))
			build.WriteString(fmt.Sprintf(`"type":1,`))
			build.WriteString(fmt.Sprintf(`"data":{`))

			build.WriteString(fmt.Sprintf(`"%s":"%s",`, "url", imageUrl))
			build.WriteString(fmt.Sprintf(`"%s":%d,`, "kind", kind))
			build.WriteString(fmt.Sprintf(`"%s":%d,`, "number", total))
			build.WriteString(fmt.Sprintf(`"%s":%s,`, "longitude", "120.890552"))
			build.WriteString(fmt.Sprintf(`"%s":%s`, "latitude", "29.457824"))

			message1 := build.String()
			message1 = strings.TrimRight(message1, ",")
			message1 = message1 + `}}`

			log.Println(message1)

			req, err := http.NewRequest("POST", devUrl+"/apiInterface/interface/iot/openApi/device/upload", bytes.NewBuffer([]byte(message1)))
			if err != nil {
				log.Println(err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("iot-super-token", thirdToken)
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				continue
			}

			res1, _ := io.ReadAll(resp.Body)
			log.Println(string(res1))

			time.Sleep(15 * time.Second)
		}
	}
}
