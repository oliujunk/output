package xphapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Token token
type Token struct {
	Token      string `json:"token"`
	Expiration int    `json:"expiration"`
	Message    string `json:"message"`
	UserID     int    `json:"userID"`
}

// User 用户信息
type User struct {
	Username int      `json:"username"`
	UserType string   `json:"userType"`
	Devices  []Device `json:"devices"`
	Pests    []Pest   `json:"pests"`
}

// Device 设备信息
type Device struct {
	DeviceID          int    `json:"facId"`
	DeviceName        string `json:"facName"`
	DeviceRemark      string `json:"remark"`
	SIM               string `json:"sim"`
	ElementExtendName string `json:"elementExtendName"`
}

// Pest 虫情设备信息
type Pest struct {
	DeviceID     string `json:"facId"`
	DeviceName   string `json:"facName"`
	DeviceRemark string `json:"remark"`
}

// DataEntity 数据
type DataEntity struct {
	DeviceID   int      `json:"deviceId"`
	DeviceName string   `json:"deviceName"`
	Entity     []Entity `json:"entity"`
}

// Entity 实体
type Entity struct {
	Datetime string `json:"datetime"`
	EUnit    string `json:"eUnit"`
	EValue   string `json:"eValue"`
	EKey     string `json:"eKey"`
	EName    string `json:"eName"`
	ENum     string `json:"eNum"`
}

type CurrentData struct {
	Datatime string `json:"dataTime"`
	E1       int    `json:"e1"`
	E2       int    `json:"e2"`
	E3       int    `json:"e3"`
	E4       int    `json:"e4"`
	E5       int    `json:"e5"`
	E6       int    `json:"e6"`
	E7       int    `json:"e7"`
	E8       int    `json:"e8"`
	E9       int    `json:"e9"`
	E10      int    `json:"e10"`
	E11      int    `json:"e11"`
	E12      int    `json:"e12"`
	E13      int    `json:"e13"`
	E14      int    `json:"e14"`
	E15      int    `json:"e15"`
	E16      int    `json:"e16"`
	J1       byte   `json:"j1"`
	J2       byte   `json:"j2"`
	J3       byte   `json:"j3"`
	J4       byte   `json:"j4"`
	J5       byte   `json:"j5"`
	J6       byte   `json:"j6"`
	J7       byte   `json:"j7"`
	J8       byte   `json:"j8"`
	J9       byte   `json:"j9"`
	J10      byte   `json:"j10"`
	J11      byte   `json:"j11"`
	J12      byte   `json:"j12"`
	J13      byte   `json:"j13"`
	J14      byte   `json:"j14"`
	J15      byte   `json:"j15"`
	J16      byte   `json:"j16"`
	J17      byte   `json:"j17"`
	J18      byte   `json:"j18"`
	J19      byte   `json:"j19"`
	J20      byte   `json:"j20"`
	J21      byte   `json:"j21"`
	J22      byte   `json:"j22"`
	J23      byte   `json:"j23"`
	J24      byte   `json:"j24"`
	J25      byte   `json:"j25"`
	J26      byte   `json:"j26"`
	J27      byte   `json:"j27"`
	J28      byte   `json:"j28"`
	J29      byte   `json:"j29"`
	J30      byte   `json:"j30"`
	J31      byte   `json:"j31"`
	J32      byte   `json:"j32"`
}

type PestData struct {
	Datatime string `json:"dataTime"`
	E1       int    `json:"e1"`
	E2       int    `json:"e2"`
	E3       int    `json:"e3"`
	E4       int    `json:"e4"`
	E5       int    `json:"e5"`
	E6       int    `json:"e6"`
	E7       int    `json:"e7"`
	E8       int    `json:"e8"`
	E9       int    `json:"e9"`
	E10      int    `json:"e10"`
	E11      int    `json:"e11"`
	E12      int    `json:"e12"`
	E13      int    `json:"e13"`
	E14      int    `json:"e14"`
	E15      int    `json:"e15"`
	E16      int    `json:"e16"`
	E17      int    `json:"e17"`
	E18      int    `json:"e18"`
	E19      int    `json:"e19"`
	E20      int    `json:"e20"`
	E21      int    `json:"e21"`
	E22      int    `json:"e22"`
	E23      int    `json:"e23"`
	E24      int    `json:"e24"`
	E25      int    `json:"e25"`
	E26      int    `json:"e26"`
	E27      int    `json:"e27"`
	E28      int    `json:"e28"`
	E29      int    `json:"e29"`
	E30      int    `json:"e30"`
	E31      int    `json:"e31"`
	E32      int    `json:"e32"`
	Imei     string `json:"imei"`
}

// GetToken 获取token
func GetToken(username, password string) string {
	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	loginParam := map[string]string{"username": username, "password": password}
	jsonStr, _ := json.Marshal(loginParam)
	resp, err := client.Post("http://115.28.187.9:8005/login", "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	var token Token
	_ = json.Unmarshal(result, &token)
	return token.Token
}

func GetDevices(username, token string) []Device {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "http://115.28.187.9:8005/user/"+username, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("token", token)
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var user User
	_ = json.Unmarshal(body, &user)
	return user.Devices
}

func GetPests(username, token string) []Pest {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "http://115.28.187.9:8005/user/"+username, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("token", token)
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var user User
	_ = json.Unmarshal(body, &user)
	return user.Pests
}
