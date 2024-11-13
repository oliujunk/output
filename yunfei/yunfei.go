package yunfei

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"
)

var (
	token         string
	devices       []xphapi.Device
	clients       []mqtt.Client
	clientsStatus []bool
)

func Start() {
	log.Println("云飞四情平台推送 start ------")
	updateToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateToken)
	_, _ = job.AddFunc("0 0 0 */1 * *", updateDevices)
	_, _ = job.AddFunc("0 */5 * * * *", sendData)
	job.Start()
}

func updateToken() {
	token = xphapi.GetToken121("四情", "123456")
}

func updateDevices() {
	devices = xphapi.GetDevices121("四情", token)

	if len(clients) > 0 {
		for _, client := range clients {
			client.Disconnect(500)
		}
		clients = []mqtt.Client{}
		clientsStatus = []bool{}
	}

	initMqttClient()
}

func initMqttClient() {
	for _, device := range devices {
		clientOptions := mqtt.NewClientOptions().AddBroker("120.27.222.26:1883").SetUsername("admin").SetPassword("password").SetClientID(fmt.Sprintf("%d", device.DeviceID))
		clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
		client := mqtt.NewClient(clientOptions)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		clients = append(clients, client)
		clientsStatus = append(clientsStatus, true)
	}
}

func sendData() {
	for index, device := range devices {
		resp, err := http.Get("http://121.43.37.41:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
		if err != nil {
			log.Println("获取数据异常")
			continue
		}
		result, _ := io.ReadAll(resp.Body)
		var dataEntity xphapi.DataEntity
		_ = json.Unmarshal(result, &dataEntity)
		if len(dataEntity.Entity) > 0 {
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Hour * 2)) {

				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"cmd":"terminalData",`))
				build.WriteString(fmt.Sprintf(`"ext":{`))
				build.WriteString(fmt.Sprintf(`"StationID":"%d",`, device.DeviceID))
				build.WriteString(fmt.Sprintf(`"MonitorTime":"%s",`, now.Format("2006-01-02 15:04:05")))
				build.WriteString(fmt.Sprintf(`"data":[`))
				for _, entity := range dataEntity.Entity {
					jsonData, err := json.Marshal(entity)
					if err != nil {
						fmt.Println("序列化失败:", err)
						continue
					}
					build.WriteString(string(jsonData))
					build.WriteString(",")
				}
				message := build.String()
				message = strings.TrimRight(message, ",")
				message = message + `]`
				message = message + `}`
				message = message + `}`

				log.Println(message)

				clients[index].Publish(fmt.Sprintf(`/yfkj/qxz/pub/%d`, device.DeviceID), 1, false, message)
			}
			time.Sleep(1 * time.Second)
		}
	}

}
