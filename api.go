package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"git.vegner.org/vsvegner/gomtc/internal/app/model"
	"github.com/sirupsen/logrus"
)

func BlockOverAPI(a *BlockDevices, p model.ParseType) {
	type Req struct {
		Ip    string
		Mac   string
		Delay string
	}
	arr := []Req{}
	for _, item := range *a {
		arr = append(arr, Req{
			Ip:    item.IP,
			Mac:   item.Mac,
			Delay: item.Delay,
		})
	}

	jsonStr, err := json.Marshal(arr)
	if err != nil {
		logrus.Error(err)
		return
	}
	url := p.GomtcSshHost + "/api/v1/block/" + p.BlockAddressList
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		logrus.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer resp.Body.Close()

}

func GetDataOverApi(p model.ParseType) []model.DeviceType {
	arrDevices, err := getDevicesFromJSON(p.GomtcSshHost, "/api/v1/devices")
	if err != nil {
		logrus.Error(err)
		return []model.DeviceType{}
	}
	return arrDevices
}

func getDevicesFromJSON(server, uri string) ([]model.DeviceType, error) {
	url := server + uri

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Set("User-Agent", "spacecount-tutorial")

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		return nil, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}
	v := []model.DeviceType{}
	jsonErr := json.Unmarshal(body, &v)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return v, nil
}
