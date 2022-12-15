package jiamusi

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"time"
)

const broker = "tcp://cosmoiot.cosmoplat.com:1883"

var (
	xphToken      string
	devices       []xphapi.Device
	clients       []mqtt.Client
	clientsStatus []bool
)

func updateXphToken() {
	xphToken = xphapi.GetToken("527110152", "123456")
}

func updateDevices() {
	devices = xphapi.GetDevices("527110152", xphToken)
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
		clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername(device.DeviceRemark)
		clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
		client := mqtt.NewClient(clientOptions)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			clientsStatus = append(clientsStatus, false)
		} else {
			clientsStatus = append(clientsStatus, true)
		}
		clients = append(clients, client)
	}

}

func Start() {
	// 佳木斯项目 卡奥斯平台推送
	log.Println("卡奥斯平台推送 start ------")
	updateXphToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateXphToken)
	_, _ = job.AddFunc("30 0 0 */1 * *", updateDevices)
	_, _ = job.AddFunc("0 0 */1 * * *", sendData)
	//_, _ = job.AddFunc("0 */1 * * * *", sendData)

	job.Start()
}

func sendData() {
	for index, device := range devices {
		resp, err := http.Get("http://115.28.187.9:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
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
			dataMap := make(map[string]string)
			for _, entity := range dataEntity.Entity {
				dataMap[entity.EName] = entity.EValue
			}
			data, err := json.Marshal(dataMap)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(index, string(data))
			if clientsStatus[index] {
				clients[index].Publish("v1/devices/me/telemetry", 0, false, string(data))
			}
		}
	}
}
