package mmkj

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/robfig/cron/v3"
)

const broker = "tcp://47.97.73.216:1883"

//const broker = "tcp://101.34.116.221:1883"

const mqttUsername = "mqtt"

//const mqttUsername = "admin"

const mqttPassword = "mqtt"

// const mqttPassword = "public"
const gatewayDeviceName1 = "tldiot00120010000055"
const gatewayDeviceName2 = "tldiot00120010000056"
const pubDeviceNamePrefix = "tldiot009100100000"

var (
	rnToken       string
	gateway1Nodes []DeviceNet
	gateway2Nodes []DeviceNet
)

type DeviceNet struct {
	device      xphapi.Device
	client      mqtt.Client
	clientState bool
	clientId    string
}

func Start() {
	log.Println("麦麦科技4G阀控器推送 start ------")
	updateXphToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateXphToken)
	_, _ = job.AddFunc("30 0 0 */1 * *", updateDevices)

	job.Start()
}

func updateXphToken() {
	rnToken = xphapi.RNGetToken("H810421", "88888888")
}

func updateDevices() {
	for _, node := range gateway1Nodes {
		if node.clientState {
			node.client.Disconnect(250)
		}
	}
	for _, node := range gateway2Nodes {
		if node.clientState {
			node.client.Disconnect(250)
		}
	}
	gateway1Nodes = gateway1Nodes[:0]
	gateway2Nodes = gateway2Nodes[:0]

	devices := xphapi.RNGetDevices("H810421-1", rnToken)
	for index, device := range devices {
		var deviceNet DeviceNet
		deviceNet.device = device

		clientId := fmt.Sprintf("%s%d", pubDeviceNamePrefix, 11+index)
		clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername(mqttUsername).SetPassword(mqttPassword).SetClientID(clientId)
		clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
		clientOptions.SetOnConnectHandler(onConnectHandler1)
		clientOptions.SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Println("断线重连")
		})
		client := mqtt.NewClient(clientOptions)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			deviceNet.clientState = false
		} else {
			deviceNet.clientState = true
		}
		deviceNet.client = client
		deviceNet.clientId = clientId
		gateway1Nodes = append(gateway1Nodes, deviceNet)
		log.Println(device.DeviceName, clientId)
	}

	devices = xphapi.RNGetDevices("H810421-2", rnToken)
	for index, device := range devices {
		var deviceNet DeviceNet
		deviceNet.device = device
		clientId := fmt.Sprintf("%s%d", pubDeviceNamePrefix, 11+18+index)
		clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername(mqttUsername).SetPassword(mqttPassword).SetClientID(clientId)
		clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
		clientOptions.SetOnConnectHandler(onConnectHandler2)
		client := mqtt.NewClient(clientOptions)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			deviceNet.clientState = false
		} else {
			deviceNet.clientState = true
		}
		deviceNet.client = client
		deviceNet.clientId = clientId

		gateway2Nodes = append(gateway2Nodes, deviceNet)
		log.Println(device.DeviceName, clientId)
	}
}

func onConnectHandler1(client mqtt.Client) {
	client.Subscribe(fmt.Sprintf("/tldiot/%s/sub", gatewayDeviceName1), 0, controlHandler)
}

func onConnectHandler2(client mqtt.Client) {
	client.Subscribe(fmt.Sprintf("/tldiot/%s/sub", gatewayDeviceName2), 0, controlHandler)
}

func controlHandler(client mqtt.Client, message mqtt.Message) {
	reader := client.OptionsReader()
	clientId := reader.ClientID()
	node, err := findDeviceNetForClientId(clientId)
	if err != nil {
		return
	}
	deviceName := node.device.DeviceName
	deviceNameList := strings.Split(deviceName, "#")
	if len(deviceNameList) < 2 {
		return
	}

	deviceNum, err := strconv.Atoi(deviceNameList[0])
	if err != nil {
		return
	}
	payload, err := simplejson.NewJson(message.Payload())
	if err != nil {
		return
	}
	log.Println(clientId, payload)

	slaveNumber := payload.Get("SlaveNumber").MustInt()

	if slaveNumber > 0 && slaveNumber == deviceNum {
		valveSwitch := payload.Get("ValveSwitch").MustInt()
		ret := false
		State := ""
		if valveSwitch == 0 {
			ret = control(node.device.DeviceID, 0, 0)
			State = "Close"
		} else if valveSwitch == 1 {
			ret = control(node.device.DeviceID, 0, 1)
			State = "Open"
		}
		if ret {
			var build strings.Builder
			build.WriteString(`{`)
			build.WriteString(fmt.Sprintf(`"State":"%s",`, State))
			build.WriteString(fmt.Sprintf(`"Operate":"%s",`, payload.Get("Operate").MustString()))
			build.WriteString(fmt.Sprintf(`"ValveSwitch":%d,`, valveSwitch))
			build.WriteString(fmt.Sprintf(`"Update":"%s",`, time.Now().Format("2006-01-02 15:04:05")))
			build.WriteString(fmt.Sprintf(`"ValveNumber":"%d"`, 1))
			build.WriteString(`}`)
			log.Println(build.String())
			client.Publish(fmt.Sprintf("/tldiot/%s/pub", clientId), 0, false, build.String())
		}
	}

}

func findDeviceNetForClientId(clientId string) (*DeviceNet, error) {
	for _, node := range gateway1Nodes {
		if node.clientId == clientId {
			return &node, nil
		}
	}
	for _, node := range gateway2Nodes {
		if node.clientId == clientId {
			return &node, nil
		}
	}
	var err error
	return nil, err
}

func control(deviceID, num, state int) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	loginParam := map[string]int{"deviceId": deviceID, "relayNum": num, "relayState": state}
	jsonStr, _ := json.Marshal(loginParam)
	request, err := http.NewRequest("POST", "http://101.34.116.221:8005/relay", bytes.NewBuffer(jsonStr))
	request.Header.Add("token", rnToken)
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		log.Println(err)
	}
	result, _ := io.ReadAll(resp.Body)
	log.Println(deviceID, string(result))
	parseBool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false
	}
	return parseBool
}
