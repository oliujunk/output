package soilmoisture

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"oliujunk/output/xphapi"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// 土壤湿度1	0
// 土壤湿度2	1
// 土壤湿度3	2
// 土壤湿度4	3
// 土壤温度1	4
// 土壤温度2	5
// 土壤温度3	6
// 土壤温度4	7
// 空气温度	8
// 空气湿度	9
// 雨量累计	10
// 最大风速	11
// 最小风速	12
// 平均风速	13
// 风向		14
// 辐射		15
// 紫外线	16
// 小时ET	17
// 日累计ET	18
// 有效降雨	19
// 累计有效降雨	20
// 电池电压	21
// NC		22

// 全国土壤墒情平台，新平台老平台均可上报，分配到 soil 用户下，该用户的后台为 whruinong，
// 修改要素名称即可自动上报

var (
	xphToken    string
	devices     []xphapi.Device
	oldXphToken string
	oldDevices  []xphapi.Device
	token101    string
	devices101  []xphapi.Device
)

func updateXphToken() {
	//xphToken = xphapi.NewGetToken("soil", "123456")
	oldXphToken = xphapi.GetToken("soil", "123456")
	token101 = xphapi.RNGetToken("soil", "88888888")
}

func updateDevices() {
	//devices = xphapi.NewGetDevices("soil", xphToken)
	oldDevices = xphapi.GetDevices("soil", oldXphToken)
	devices101 = xphapi.RNGetDevices("soil", token101)
}

func Start() {
	// 全国土壤墒情平台
	log.Println("全国土壤墒情平台 start ------")
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

func sendData() {
	now := time.Now()

	// 新平台设备
	for _, device := range devices {
		if len(device.DeviceRemark) <= 0 {
			continue
		}
		resp, err := http.Get("http://47.105.215.208:8005/intfa/queryData/" + strconv.Itoa(device.DeviceID))
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
			var sendIndex [23]int
			for i := 0; i < 23; i++ {
				sendIndex[i] = 32
			}
			for index, entity := range dataEntity.Entity {
				name := strings.Split(entity.EName, "_")
				if len(name) == 2 {
					indexValue, _ := strconv.Atoi(name[1])
					if indexValue <= 23 {
						sendIndex[indexValue] = index
					}
				}
			}

			content := "005," + device.DeviceRemark + "," +
				fmt.Sprintf("%4d-%2d-%2d %2d:%2d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()) + ","

			for i := 0; i < 23; i++ {
				if sendIndex[i] < 32 {
					content += dataEntity.Entity[sendIndex[i]].EValue + ","
				} else {
					content += "0,"
				}
			}

			content += "0,0"

			log.Printf("[%d]: %s\n", device.DeviceID, content)
			conn, err := net.Dial("tcp", "123.127.160.49:10001")
			if err != nil {
				log.Println(err.Error())
			}
			if conn != nil {
				_, _ = conn.Write([]byte(content))
				recv := make([]byte, 20)
				recvBytes, err := conn.Read(recv)
				if err != nil {
					log.Println(err.Error())
				}
				log.Println(string(recv[:recvBytes]))
				conn.Close()
			}
			time.Sleep(2 * time.Second)
		}
	}

	// 老平台设备
	for _, device := range oldDevices {
		if len(device.DeviceRemark) <= 0 {
			continue
		}
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
			var sendIndex [23]int
			for i := 0; i < 23; i++ {
				sendIndex[i] = 32
			}
			for index, entity := range dataEntity.Entity {
				name := strings.Split(entity.EName, "_")
				if len(name) == 2 {
					indexValue, _ := strconv.Atoi(name[1])
					if indexValue <= 23 {
						sendIndex[indexValue] = index
					}
				}
			}

			content := "005," + device.DeviceRemark + "," +
				fmt.Sprintf("%4d-%2d-%2d %2d:%2d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()) + ","

			for i := 0; i < 23; i++ {
				if sendIndex[i] < 32 {
					content += dataEntity.Entity[sendIndex[i]].EValue + ","
				} else {
					content += "0,"
				}
			}

			content += "0,0"

			log.Printf("[%d]: %s\n", device.DeviceID, content)
			conn, err := net.Dial("tcp", "123.127.160.49:10001")
			if err != nil {
				log.Println(err.Error())
			}
			if conn != nil {
				_, _ = conn.Write([]byte(content))
				recv := make([]byte, 20)
				recvBytes, err := conn.Read(recv)
				if err != nil {
					log.Println(err.Error())
				}
				log.Println(string(recv[:recvBytes]))
				conn.Close()
			}
			time.Sleep(2 * time.Second)
		}
	}

	// 101平台设备
	for _, device := range devices101 {
		if len(device.DeviceRemark) <= 0 {
			continue
		}
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
			var sendIndex [23]int
			for i := 0; i < 23; i++ {
				sendIndex[i] = 32
			}
			for index, entity := range dataEntity.Entity {
				name := strings.Split(entity.EName, "_")
				if len(name) == 2 {
					indexValue, _ := strconv.Atoi(name[1])
					if indexValue <= 23 {
						sendIndex[indexValue] = index
					}
				}
			}

			content := "005," + device.DeviceRemark + "," +
				fmt.Sprintf("%4d-%2d-%2d %2d:%2d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()) + ","

			for i := 0; i < 23; i++ {
				if sendIndex[i] < 32 {
					content += dataEntity.Entity[sendIndex[i]].EValue + ","
				} else {
					content += "0,"
				}
			}

			content += "0,0"

			log.Printf("[%d]: %s\n", device.DeviceID, content)
			conn, err := net.Dial("tcp", "123.127.160.49:10001")
			if err != nil {
				log.Println(err.Error())
			}
			if conn != nil {
				_, _ = conn.Write([]byte(content))
				recv := make([]byte, 20)
				recvBytes, err := conn.Read(recv)
				if err != nil {
					log.Println(err.Error())
				}
				log.Println(string(recv[:recvBytes]))
				conn.Close()
			}
			time.Sleep(2 * time.Second)
		}
	}

}
