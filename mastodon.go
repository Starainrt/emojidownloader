package main

import (
	"errors"
	"io/ioutil"

	"b612.me/starnet"
)

func GetEmojiNormal(url, proxy string) ([]byte, error) {
	url = url + "/api/v1/custom_emojis"
	req := starnet.NewRequests(url, nil, "GET")
	if proxy != "" {
		req.Proxy = proxy
	}
	data, err := starnet.Curl(req)
	if err != nil {
		return nil, err
	}
	if data.RespHttpCode == 401 {
		return nil, errors.New("Cannot Get Emojis!Server Returned 401,Authorized_Fetch May Opended by the Server's Administrator")
	}
	return data.RecvData, nil
}

func GetEmojiFromFile(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
}
