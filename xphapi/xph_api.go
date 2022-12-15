package xphapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// NewGetToken 获取token
func NewGetToken(username, password string) string {
	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	loginParam := map[string]string{"username": username, "password": password}
	jsonStr, _ := json.Marshal(loginParam)
	resp, err := client.Post("http://47.105.215.208:8005/login", "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	var token Token
	_ = json.Unmarshal(result, &token)
	return token.Token
}

// NewGetDevices 获取设备ID
func NewGetDevices(username, token string) []Device {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "http://47.105.215.208:8005/user/"+username, nil)
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
