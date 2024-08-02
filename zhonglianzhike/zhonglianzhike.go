package zhonglianzhike

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
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

const broker = "tcp://112.35.122.205:1883"

var (
	rnToken       string
	devices       []xphapi.Device
	clients       []mqtt.Client
	clientsStatus []bool
)

type ControlGroupParam struct {
	DeviceID int   `json:"deviceId"`
	NumList  []int `json:"numList"`
	State    int   `json:"state"`
}

func updateXphToken() {
	rnToken = xphapi.RNGetToken("zhonglianzhike", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("zhonglianzhike", rnToken)
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
		clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername(fmt.Sprintf("%d", device.DeviceID)).SetClientID(fmt.Sprintf("%d", device.DeviceID))
		clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
		client := mqtt.NewClient(clientOptions)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			clientsStatus = append(clientsStatus, false)
		} else {
			clientsStatus = append(clientsStatus, true)
			client.Subscribe("v1/devices/me/rpc/request/+", 0, controlHandler)
		}
		clients = append(clients, client)
	}
}

func Start() {
	log.Println("中联智科 start ------")
	updateXphToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateXphToken)
	_, _ = job.AddFunc("30 0 0 */1 * *", updateDevices)
	//_, _ = job.AddFunc("0 0 */1 * * *", sendData)
	_, _ = job.AddFunc("0 */5 * * * *", sendData)

	job.Start()
}

func sendData() {
	for index, device := range devices {
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
			dataMap := make(map[string]string)
			for _, entity := range dataEntity.Entity {
				dataMap[entity.EName] = entity.EValue
			}

			if len(dataEntity.RelayEntity) > 0 {
				for _, entity1 := range dataEntity.RelayEntity {
					dataMap[entity1.RName] = fmt.Sprintf("%d", entity1.RState)
				}
			}

			data, err := json.Marshal(dataMap)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(device.DeviceID, string(data))
			if clientsStatus[index] {
				clients[index].Publish("v1/devices/me/telemetry", 0, false, string(data))
				clients[index].Publish("v1/devices/me/attributes", 0, false, string(data))
			}
		}
	}
}

func controlHandler(client mqtt.Client, message mqtt.Message) {
	reader := client.OptionsReader()
	deviceIDS := reader.ClientID()
	deviceID, err := strconv.Atoi(deviceIDS)
	if err != nil {
		return
	}

	payload, err := simplejson.NewJson(message.Payload())
	if err != nil {
		return
	}

	topic := message.Topic()

	topicList := strings.Split(topic, "/")
	requestId := topicList[len(topicList)-1]

	method := payload.Get("method").MustString()
	params := payload.Get("params").MustString()

	if method == "setPower" {
		openNum := make([]int, 0)
		closeNum := make([]int, 0)
		for i := 0; i < len(params); i++ {
			if params[i] == '1' {
				openNum = append(openNum, i)
			} else if params[i] == '0' {
				closeNum = append(closeNum, i)
			}
		}

		if len(openNum) > 0 {
			controlGroup(deviceID, openNum, 1)
			time.Sleep(2 * time.Second)
		}

		if len(closeNum) > 0 {
			controlGroup(deviceID, closeNum, 0)
		}

		var build strings.Builder
		build.WriteString(`{`)
		build.WriteString(`"method":"setPower",`)
		build.WriteString(`"state":0`)
		build.WriteString(`}`)
		log.Println(build.String())
		client.Publish(fmt.Sprintf("v1/devices/me/rpc/response/%s", requestId), 1, false, build.String())

	} else if method == "setFrequency1" {

		value, err := strconv.Atoi(params)
		if err != nil {
			return
		}

		setProperty(deviceID, 0, value)

		var build strings.Builder
		build.WriteString(`{`)
		build.WriteString(`"method":"setPower",`)
		build.WriteString(`"state":0`)
		build.WriteString(`}`)
		log.Println(build.String())
		client.Publish(fmt.Sprintf("v1/devices/me/rpc/response/%s", requestId), 1, false, build.String())

	} else if method == "setFrequency2" {

		value, err := strconv.Atoi(params)
		if err != nil {
			return
		}

		setProperty(deviceID, 1, value)

		var build strings.Builder
		build.WriteString(`{`)
		build.WriteString(`"method":"setPower",`)
		build.WriteString(`"state":0`)
		build.WriteString(`}`)
		log.Println(build.String())
		client.Publish(fmt.Sprintf("v1/devices/me/rpc/response/%s", requestId), 1, false, build.String())
	}
}

func controlGroup(deviceID int, numList []int, state int) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	controlGroupParam := ControlGroupParam{DeviceID: deviceID, NumList: numList, State: state}
	jsonStr, _ := json.Marshal(controlGroupParam)
	log.Println(string(jsonStr))
	request, err := http.NewRequest("POST", "http://101.34.116.221:8005/relay/group", bytes.NewBuffer(jsonStr))
	request.Header.Add("token", rnToken)
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}
	result, _ := io.ReadAll(resp.Body)
	log.Println(string(result))
	parseBool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false
	}
	return parseBool
}

func setProperty(deviceID int, num int, value int) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	setPropertyParam := map[string]int{"deviceId": deviceID, "num": num, "value": value}
	jsonStr, _ := json.Marshal(setPropertyParam)
	log.Println(string(jsonStr))
	request, err := http.NewRequest("POST", "http://101.34.116.221:8005/property", bytes.NewBuffer(jsonStr))
	request.Header.Add("token", rnToken)
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}
	result, _ := io.ReadAll(resp.Body)
	log.Println(string(result))
	parseBool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false
	}
	return parseBool
}
