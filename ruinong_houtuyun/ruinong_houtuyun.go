package ruinong_houtuyun

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"oliujunk/output/xphapi"
	"os"
	"strconv"
	"strings"
	"time"
)

const broker = "ssl://lbs-hivemqtt-private-hty.lunz.cn:8883"

var (
	rnToken       string
	devices       []xphapi.Device
	clients       []mqtt.Client
	clientsStatus []bool

	sporeClient mqtt.Client
	pestClient  mqtt.Client
	pestImei    = "K09577011"
	sporeImei   = "867814046371775"
)

func updateXphToken() {
	rnToken = xphapi.RNGetToken("houtuyun", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("houtuyun", rnToken)
	if len(clients) > 0 {
		for _, client := range clients {
			client.Disconnect(500)
		}
		clients = []mqtt.Client{}
		clientsStatus = []bool{}
	}

	initMqttClient()

	initSporeMqttClient()

	initPestMqttClient()
}

func NewTLSConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile("./ca-zs1.pem")
	if err != nil {
		log.Fatalln(err.Error())
	}
	certpool.AppendCertsFromPEM(ca)

	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
	}
}

func initMqttClient() {
	for _, device := range devices {
		tlsConfig := NewTLSConfig()
		clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername("zrlbsmqttgateway1").SetPassword("zrlbsmqttgateway1pw").SetTLSConfig(tlsConfig).SetClientID(fmt.Sprintf("%d", device.DeviceID))
		clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
		client := mqtt.NewClient(clientOptions)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		clients = append(clients, client)
		clientsStatus = append(clientsStatus, true)
		var build strings.Builder
		build.WriteString(`{`)
		build.WriteString(fmt.Sprintf(`"did":"%d",`, device.DeviceID))
		build.WriteString(fmt.Sprintf(`"gid":"%d",`, device.DeviceID))
		build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
		build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
		build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
		build.WriteString(fmt.Sprintf(`"func":%d,`, 98))
		build.WriteString(fmt.Sprintf(`"info":"%s"`, "Online"))
		build.WriteString(`}`)
		log.Println(build.String())
		client.Publish("/Iot/Status", 1, true, build.String())

		client.Subscribe("/Iot/Sub", 1, controlHandler)
	}
}

func initSporeMqttClient() {
	tlsConfig := NewTLSConfig()
	clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername("zrlbsmqttgateway1").SetPassword("zrlbsmqttgateway1pw").SetTLSConfig(tlsConfig)
	clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
	client := mqtt.NewClient(clientOptions)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	sporeClient = client
	var build strings.Builder
	build.WriteString(`{`)
	build.WriteString(fmt.Sprintf(`"did":"%s",`, sporeImei))
	build.WriteString(fmt.Sprintf(`"gid":"%s",`, sporeImei))
	build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
	build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
	build.WriteString(fmt.Sprintf(`"func":%d,`, 98))
	build.WriteString(fmt.Sprintf(`"info":"%s"`, "Online"))
	build.WriteString(`}`)
	log.Println(build.String())
	sporeClient.Publish("/Iot/Status", 1, true, build.String())
}

func initPestMqttClient() {
	tlsConfig := NewTLSConfig()
	clientOptions := mqtt.NewClientOptions().AddBroker(broker).SetUsername("zrlbsmqttgateway1").SetPassword("zrlbsmqttgateway1pw").SetTLSConfig(tlsConfig)
	clientOptions.SetConnectTimeout(time.Duration(60) * time.Second)
	client := mqtt.NewClient(clientOptions)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	pestClient = client
	var build strings.Builder
	build.WriteString(`{`)
	build.WriteString(fmt.Sprintf(`"did":"%s",`, pestImei))
	build.WriteString(fmt.Sprintf(`"gid":"%s",`, pestImei))
	build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
	build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
	build.WriteString(fmt.Sprintf(`"func":%d,`, 98))
	build.WriteString(fmt.Sprintf(`"info":"%s"`, "Online"))
	build.WriteString(`}`)
	log.Println(build.String())
	pestClient.Publish("/Iot/Status", 1, true, build.String())
}

func Start() {
	log.Println("睿农-后土云平台推送 start ------")
	updateXphToken()
	updateDevices()
	job := cron.New(
		cron.WithSeconds(),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	_, _ = job.AddFunc("0 0 0/12 * * *", updateXphToken)
	_, _ = job.AddFunc("30 0 0 */1 * *", updateDevices)
	_, _ = job.AddFunc("0 */10 * * * *", sendData)
	_, _ = job.AddFunc("0 */2 * * * *", sendHeartBeat)

	job.Start()
}

func sendHeartBeat() {
	var build strings.Builder
	for index, device := range devices {
		if clientsStatus[index] {
			build.Reset()
			build.WriteString(`{`)
			build.WriteString(fmt.Sprintf(`"did":"%d",`, device.DeviceID))
			build.WriteString(fmt.Sprintf(`"gid":"%d",`, device.DeviceID))
			build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
			build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
			build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
			build.WriteString(fmt.Sprintf(`"func":%d`, 1))
			build.WriteString(`}`)
			log.Println(index, build.String())
			clients[index].Publish("/Iot/Pub", 1, false, build.String())
		}
	}

	build.Reset()
	build.WriteString(`{`)
	build.WriteString(fmt.Sprintf(`"did":"%s",`, pestImei))
	build.WriteString(fmt.Sprintf(`"gid":"%s",`, pestImei))
	build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
	build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
	build.WriteString(fmt.Sprintf(`"func":%d`, 1))
	build.WriteString(`}`)
	log.Println(build.String())
	pestClient.Publish("/Iot/Pub", 1, false, build.String())

	build.Reset()
	build.WriteString(`{`)
	build.WriteString(fmt.Sprintf(`"did":"%s",`, sporeImei))
	build.WriteString(fmt.Sprintf(`"gid":"%s",`, sporeImei))
	build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
	build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
	build.WriteString(fmt.Sprintf(`"func":%d`, 1))
	build.WriteString(`}`)
	log.Println(build.String())
	sporeClient.Publish("/Iot/Pub", 1, false, build.String())
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
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", dataEntity.Entity[0].Datetime)
			if datatime.After(now.Add(-time.Minute * 60)) {
				var dataBuilder strings.Builder
				dataBuilder.WriteString(`{`)
				for index, entity := range dataEntity.Entity {
					if index+1 == len(dataEntity.Entity) {
						dataBuilder.WriteString(fmt.Sprintf(`"%s":%s`, entity.ENum, entity.EValue))
					} else {
						dataBuilder.WriteString(fmt.Sprintf(`"%s":%s,`, entity.ENum, entity.EValue))
					}
				}

				if len(dataEntity.RelayEntity) > 0 {
					dataBuilder.WriteString(`,`)
					for index1, entity1 := range dataEntity.RelayEntity {
						if index1+1 == len(dataEntity.RelayEntity) {
							dataBuilder.WriteString(fmt.Sprintf(`"%d":%d`, entity1.RIndex, entity1.RState))
						} else {
							dataBuilder.WriteString(fmt.Sprintf(`"%d":%d,`, entity1.RIndex, entity1.RState))
						}
					}
				}

				dataBuilder.WriteString(`}`)

				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"did":"%d",`, device.DeviceID))
				build.WriteString(fmt.Sprintf(`"gid":"%d",`, device.DeviceID))
				build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
				build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
				build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
				build.WriteString(fmt.Sprintf(`"func":%d,`, 2))
				build.WriteString(fmt.Sprintf(`"level":%d,`, 100))
				build.WriteString(fmt.Sprintf(`"consume":%d,`, 1000))
				build.WriteString(fmt.Sprintf(`"err":%d,`, 0))
				build.WriteString(fmt.Sprintf(`"points":%s`, dataBuilder.String()))
				build.WriteString(`}`)
				log.Println(index, build.String())
				clients[index].Publish("/Iot/Pub", 1, false, build.String())
			} else {
				if clientsStatus[index] {
					var build strings.Builder
					build.WriteString(`{`)
					build.WriteString(fmt.Sprintf(`"did":"%d",`, device.DeviceID))
					build.WriteString(fmt.Sprintf(`"gid":"%d",`, device.DeviceID))
					build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
					build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
					build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
					build.WriteString(fmt.Sprintf(`"func":%d,`, 99))
					build.WriteString(fmt.Sprintf(`"info":"%s"`, "GateWayOffline"))
					build.WriteString(`}`)
					log.Println(index, build.String())
					clients[index].Publish("/Iot/Pub", 1, false, build.String())
					clientsStatus[index] = false
				}
			}
			time.Sleep(1 * time.Second)
		}

	}

	sendPestData()

	sendSporeData()
}

func sendPestData() {
	pestData := getPestData(pestImei, rnToken)
	var dataBuilder strings.Builder
	dataBuilder.WriteString(`{`)
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%.1f,`, "102", float32(pestData.E1)/10.0))      // 湿度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%.1f,`, "101", float32(pestData.E2)/10.0))      // 温度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "112", pestData.E3))                      // 照度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%.6f,`, "262", float32(pestData.E4)/1000000.0)) // 经度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%.6f,`, "263", float32(pestData.E5)/1000000.0)) // 纬度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d`, "277", pestData.E6))                       // 诱虫数
	dataBuilder.WriteString(`}`)

	var build strings.Builder
	build.WriteString(`{`)
	build.WriteString(fmt.Sprintf(`"did":"%s",`, pestImei))
	build.WriteString(fmt.Sprintf(`"gid":"%s",`, pestImei))
	build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
	build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
	build.WriteString(fmt.Sprintf(`"func":%d,`, 2))
	build.WriteString(fmt.Sprintf(`"level":%d,`, 100))
	build.WriteString(fmt.Sprintf(`"consume":%d,`, 1000))
	build.WriteString(fmt.Sprintf(`"err":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"points":%s`, dataBuilder.String()))
	build.WriteString(`}`)
	log.Println(build.String())
	pestClient.Publish("/Iot/Pub", 1, false, build.String())
}

func getPestData(imei, token string) xphapi.PestData {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", "http://101.34.116.221:8005/pest/dataextend/"+imei, nil)

	req.Header.Set("token", token)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pestData xphapi.PestData
	_ = json.Unmarshal(body, &pestData)
	return pestData
}

func sendSporeData() {
	sporeData := getPestData(sporeImei, rnToken)
	var dataBuilder strings.Builder
	dataBuilder.WriteString(`{`)
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "264", sporeData.E1))                       // 工作状态
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "265", sporeData.E2))                       // 制冷机开关
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "266", sporeData.E3))                       // 信号强度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "267", sporeData.E4))                       // 电池状态
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "268", sporeData.E5))                       // 海拔高度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "102", sporeData.E6))                       // 湿度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "269", sporeData.E7))                       // 设备开关
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "101", sporeData.E8))                       // 温度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "270", sporeData.E9))                       // 培养时间
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "271", sporeData.E10))                      // 保温仓设定温度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "272", sporeData.E11))                      // 雨控状态
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "273", sporeData.E12))                      // 机箱温度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "124", sporeData.E13))                      // 电压
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "274", sporeData.E14))                      // 保温仓当前温度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%.6f,`, "262", float32(sporeData.E15)/1000000.0)) // 经度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%.6f,`, "263", float32(sporeData.E16)/1000000.0)) // 纬度
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d,`, "275", sporeData.E17))                      // 摄像头状态
	dataBuilder.WriteString(fmt.Sprintf(`"%s":%d`, "276", sporeData.E18))                       // 风机开关
	dataBuilder.WriteString(`}`)

	var build strings.Builder
	build.WriteString(`{`)
	build.WriteString(fmt.Sprintf(`"did":"%s",`, sporeImei))
	build.WriteString(fmt.Sprintf(`"gid":"%s",`, sporeImei))
	build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
	build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
	build.WriteString(fmt.Sprintf(`"func":%d,`, 2))
	build.WriteString(fmt.Sprintf(`"level":%d,`, 100))
	build.WriteString(fmt.Sprintf(`"consume":%d,`, 1000))
	build.WriteString(fmt.Sprintf(`"err":%d,`, 0))
	build.WriteString(fmt.Sprintf(`"points":%s`, dataBuilder.String()))
	build.WriteString(`}`)
	log.Println(build.String())
	sporeClient.Publish("/Iot/Pub", 1, false, build.String())
}

func controlHandler(client mqtt.Client, message mqtt.Message) {
	reader := client.OptionsReader()
	deviceID := reader.ClientID()

	payload, err := simplejson.NewJson(message.Payload())
	if err != nil {
		return
	}

	if deviceID == payload.Get("did").MustString() {

		log.Println(string(message.Payload()))

		numS := payload.Get("cmd").Get("K").MustString()
		stateS := payload.Get("cmd").Get("V").MustString()
		func111 := payload.Get("func").MustInt()
		num, err := strconv.Atoi(numS)
		if err != nil {
			return
		}
		state, err := strconv.Atoi(stateS)
		if err != nil {
			return
		}
		if func111 == 83 {
			log.Println(deviceID, num, state)
			id, _ := strconv.Atoi(deviceID)
			result := control(id, num, state)
			if result {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"did":"%s",`, deviceID))
				build.WriteString(fmt.Sprintf(`"gid":"%s",`, deviceID))
				build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
				build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
				build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
				build.WriteString(fmt.Sprintf(`"func":%d,`, 4))
				build.WriteString(fmt.Sprintf(`"code":"%s",`, payload.Get("code").MustString()))
				build.WriteString(fmt.Sprintf(`"err":%d,`, 0))
				build.WriteString(fmt.Sprintf(`"info":"%t"`, result))
				build.WriteString(`}`)
				log.Println(build.String())
				client.Publish("/Iot/Pub", 1, false, build.String())
			} else {
				var build strings.Builder
				build.WriteString(`{`)
				build.WriteString(fmt.Sprintf(`"did":"%s",`, deviceID))
				build.WriteString(fmt.Sprintf(`"gid":"%s",`, deviceID))
				build.WriteString(fmt.Sprintf(`"ptid":%d,`, 0))
				build.WriteString(fmt.Sprintf(`"cid":%d,`, 1))
				build.WriteString(fmt.Sprintf(`"time":"%s",`, time.Now().Format("2006/01/02 15:04:05")))
				build.WriteString(fmt.Sprintf(`"func":%d,`, 4))
				build.WriteString(fmt.Sprintf(`"code":"%s",`, payload.Get("code").MustString()))
				build.WriteString(fmt.Sprintf(`"err":%d,`, 5))
				build.WriteString(fmt.Sprintf(`"info":"%t"`, result))
				build.WriteString(`}`)
				log.Println(build.String())
				client.Publish("/Iot/Pub", 1, false, build.String())
			}
		}
	}
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
	log.Println(string(result))
	parseBool, err := strconv.ParseBool(string(result))
	if err != nil {
		return false
	}
	return parseBool
}
