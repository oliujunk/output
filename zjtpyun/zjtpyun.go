package zjtpyun

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
	rnToken string
	devices []xphapi.Device

	oldXphToken string
	oldDevices  []xphapi.Device
)

func Start() {
	log.Println("浙江托普云农 start ------")
	updateXphToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateXphToken)
	_, _ = job.AddFunc("0 0 0/1 * * *", updateDevices)
	_, _ = job.AddFunc("0 0 */1 * * *", sendData)
	//_, _ = job.AddFunc("0 */1 * * * *", sendData)
	job.Start()
}

func updateXphToken() {
	rnToken = xphapi.RNGetToken("zjtpyun", "88888888")
	oldXphToken = xphapi.GetToken("zjtpyun", "123456")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("zjtpyun", rnToken)
	oldDevices = xphapi.GetDevices("zjtpyun", oldXphToken)
}

func sendData() {
	for _, device := range devices {

		deviceNum := strings.Split(device.DeviceRemark, ",")

		if len(deviceNum) != 2 {
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
			if datatime.After(now.Add(-time.Minute * 6)) {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"latitude":%f,`, device.Latitude))
				build.WriteString(fmt.Sprintf(`"longitude":%f,`, device.Longitude))
				build.WriteString(`"sensors":[`)
				for _, entity := range dataEntity.Entity {
					name := strings.Split(entity.EName, ",")
					if len(name[0]) <= 0 {
						continue
					}

					if name[0] == "129" {
						build.WriteString(`{`)
						build.WriteString(fmt.Sprintf(`"data_type":%d,`, 0))
						build.WriteString(fmt.Sprintf(`"collect_time":"%s",`, entity.Datetime))
						build.WriteString(fmt.Sprintf(`"sensor_tag":%d,`, 1))
						build.WriteString(fmt.Sprintf(`"sensor_type_id":%s,`, name[0]))
						build.WriteString(fmt.Sprintf(`"sensor_val":%s`, entity.EValue))
						build.WriteString(`},`)

						build.WriteString(`{`)
						build.WriteString(fmt.Sprintf(`"data_type":%d,`, 0))
						build.WriteString(fmt.Sprintf(`"collect_time":"%s",`, entity.Datetime))
						build.WriteString(fmt.Sprintf(`"sensor_tag":%d,`, 1))
						build.WriteString(fmt.Sprintf(`"sensor_type_id":%s,`, "161"))
						build.WriteString(fmt.Sprintf(`"sensor_val":%s`, entity.EValue))
						build.WriteString(`},`)
					} else if name[0] == "145" {
						build.WriteString(`{`)
						build.WriteString(fmt.Sprintf(`"data_type":%d,`, 0))
						build.WriteString(fmt.Sprintf(`"collect_time":"%s",`, entity.Datetime))
						build.WriteString(fmt.Sprintf(`"sensor_tag":%d,`, 1))
						build.WriteString(fmt.Sprintf(`"sensor_type_id":%s,`, name[0]))
						build.WriteString(fmt.Sprintf(`"sensor_val":%s`, entity.EValue))
						build.WriteString(`},`)

						build.WriteString(`{`)
						build.WriteString(fmt.Sprintf(`"data_type":%d,`, 0))
						build.WriteString(fmt.Sprintf(`"collect_time":"%s",`, entity.Datetime))
						build.WriteString(fmt.Sprintf(`"sensor_tag":%d,`, 1))
						build.WriteString(fmt.Sprintf(`"sensor_type_id":%s,`, "162"))
						build.WriteString(fmt.Sprintf(`"sensor_val":%s`, entity.EValue))
						build.WriteString(`},`)
					} else {
						build.WriteString(`{`)
						build.WriteString(fmt.Sprintf(`"data_type":%d,`, 0))
						build.WriteString(fmt.Sprintf(`"collect_time":"%s",`, entity.Datetime))
						build.WriteString(fmt.Sprintf(`"sensor_tag":%d,`, 1))
						build.WriteString(fmt.Sprintf(`"sensor_type_id":%s,`, name[0]))
						build.WriteString(fmt.Sprintf(`"sensor_val":%s`, entity.EValue))
						build.WriteString(`},`)
					}
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `]}`

				log.Println(message)

				timestamp := time.Now().Format("20060102150405")
				signContent := "c2658fe999ee4c6a9ec66c05c51bf95c" +
					"app_keye42ff6f583354801a73d35f35b137b37" +
					"data" + message +
					"formatjson" +
					"methodtop.custom.dataupload" +
					"serial_num" + deviceNum[1] +
					"timestamp" + timestamp +
					"versionv1.0" +
					"c2658fe999ee4c6a9ec66c05c51bf95c"
				signCode := md5.Sum([]byte(signContent))
				sign := hex.EncodeToString(signCode[:])

				params := url.Values{}
				Url, err := url.Parse("https://api.topyn.cn/rest")
				if err != nil {
					return
				}
				params.Set("app_key", "e42ff6f583354801a73d35f35b137b37")
				params.Set("method", "top.custom.dataupload")
				params.Set("format", "json")
				params.Set("version", "v1.0")
				params.Set("timestamp", timestamp)
				params.Set("serial_num", deviceNum[1])
				params.Set("sign", sign)
				params.Set("data", message)
				Url.RawQuery = params.Encode()
				urlPath := Url.String()
				resp, err := http.Post(urlPath, "application/json", nil)
				if err != nil {
					log.Println(err)
					continue
				}

				result, _ := ioutil.ReadAll(resp.Body)
				log.Println(string(result))

			}
		}
		time.Sleep(1 * time.Second)
	}

	for _, device := range oldDevices {

		deviceNum := strings.Split(device.DeviceRemark, ",")

		if len(deviceNum) != 2 {
			continue
		}

		resp, err := http.Get("http://115.28.187.9:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
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
			if datatime.After(now.Add(-time.Minute * 6)) {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"latitude":%f,`, device.Latitude))
				build.WriteString(fmt.Sprintf(`"longitude":%f,`, device.Longitude))
				build.WriteString(`"sensors":[`)
				for _, entity := range dataEntity.Entity {
					name := strings.Split(entity.EName, ",")
					if len(name[0]) <= 0 {
						continue
					}
					build.WriteString(`{`)
					build.WriteString(fmt.Sprintf(`"data_type":%d,`, 0))
					build.WriteString(fmt.Sprintf(`"collect_time":"%s",`, entity.Datetime))
					build.WriteString(fmt.Sprintf(`"sensor_tag":%d,`, 1))
					build.WriteString(fmt.Sprintf(`"sensor_type_id":%s,`, name[0]))
					build.WriteString(fmt.Sprintf(`"sensor_val":%s`, entity.EValue))
					build.WriteString(`},`)
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `]}`

				log.Println(message)

				timestamp := time.Now().Format("20060102150405")
				signContent := "c2658fe999ee4c6a9ec66c05c51bf95c" +
					"app_keye42ff6f583354801a73d35f35b137b37" +
					"data" + message +
					"formatjson" +
					"methodtop.custom.dataupload" +
					"serial_num" + deviceNum[1] +
					"timestamp" + timestamp +
					"versionv1.0" +
					"c2658fe999ee4c6a9ec66c05c51bf95c"
				signCode := md5.Sum([]byte(signContent))
				sign := hex.EncodeToString(signCode[:])

				params := url.Values{}
				Url, err := url.Parse("https://api.topyn.cn/rest")
				if err != nil {
					return
				}
				params.Set("app_key", "e42ff6f583354801a73d35f35b137b37")
				params.Set("method", "top.custom.dataupload")
				params.Set("format", "json")
				params.Set("version", "v1.0")
				params.Set("timestamp", timestamp)
				params.Set("serial_num", deviceNum[1])
				params.Set("sign", sign)
				params.Set("data", message)
				Url.RawQuery = params.Encode()
				urlPath := Url.String()
				resp, err := http.Post(urlPath, "application/json", nil)
				if err != nil {
					log.Println(err)
					continue
				}

				result, _ := ioutil.ReadAll(resp.Body)
				log.Println(string(result))

			}
		}
		time.Sleep(1 * time.Second)
	}
}
