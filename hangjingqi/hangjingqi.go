package hangjingqi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"oliujunk/output/utils"
	"oliujunk/output/xphapi"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	token   string
	devices []xphapi.Device
)

func updateToken() {
	token = xphapi.RNGetToken("H810031", "88888888")
}

func updateDevices() {
	devices = xphapi.RNGetDevices("H810031", token)
}

func Start() {
	// 佳木斯项目 卡奥斯平台推送
	log.Println("杭锦旗项目平台推送 start ------")
	updateToken()
	updateDevices()
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
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest("GET", "http://101.34.116.221:8005/data/"+strconv.Itoa(device.DeviceID), nil)
		if err != nil {
			log.Println("获取数据异常")
			continue
		}
		req.Header.Set("token", token)
		resp, err := client.Do(req)
		if err != nil {
			log.Println("获取数据异常")
			continue
		}
		result, _ := ioutil.ReadAll(resp.Body)
		var currentData xphapi.CurrentData
		_ = json.Unmarshal(result, &currentData)
		if currentData.Datatime != "" {
			now := time.Now()
			datatime, _ := time.Parse("2006-01-02 15:04:05", currentData.Datatime)
			if datatime.After(now.Add(-time.Minute*6)) || true {
				sendBuf := make([]byte, 72)
				sendBuf[0] = byte(device.DeviceID / 1000000)
				sendBuf[1] = byte(device.DeviceID / 10000 % 100)
				sendBuf[2] = byte(device.DeviceID / 100 % 100)
				sendBuf[3] = byte(device.DeviceID % 100)
				sendBuf[4] = 0xA2
				sendBuf[5] = 0x40
				for i := 0; i < 16; i++ {
					sendBuf[6+i*2] = byte(int16(GetFieldName(fmt.Sprintf("E%d", i+1), currentData)) >> 8)
					sendBuf[7+i*2] = byte(int16(GetFieldName(fmt.Sprintf("E%d", i+1), currentData)) & 0xFF)
				}
				for i := 0; i < 32; i++ {
					sendBuf[38+i] = 0x00
				}
				crc := utils.Crc16(sendBuf, 70)
				sendBuf[70] = byte(crc & 0xFF)
				sendBuf[71] = byte(crc >> 8)
				log.Println(strings.ToUpper(hex.EncodeToString(sendBuf)))
				//conn, err := net.Dial("tcp", "119.3.182.168:14088")
				conn, err := net.Dial("tcp", "119.3.182.168:14088")
				if err != nil {
					log.Printf("connect failed, err : %v\n\n", err.Error())
				}
				if conn != nil {
					_, _ = conn.Write(sendBuf)
					_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
					receive := make([]byte, 20)
					receiveBytes, err := conn.Read(receive)
					if err != nil {
						log.Printf(err.Error())
					}
					if receiveBytes > 0 {
						log.Printf(strings.ToUpper(hex.EncodeToString(receive[:receiveBytes])))
					}
					_ = conn.Close()
				}
			}
		}
	}
}

func GetFieldName(columnName string, currentData xphapi.CurrentData) int64 {
	var val int64
	t := reflect.TypeOf(currentData)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		fmt.Println("Check type error not Struct")
		return 0
	}
	fieldNum := t.NumField()
	for i := 0; i < fieldNum; i++ {
		if strings.ToUpper(t.Field(i).Name) == strings.ToUpper(columnName) {
			v := reflect.ValueOf(currentData)
			val := v.FieldByName(t.Field(i).Name).Int()
			return val
		}
	}
	return val
}
