package PixivLocalReverseProxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type dnsQueryRequest struct {
	HostName string
}

type dnsQueryResponse struct {
	Status   int            `json:"Status"`
	TC       bool           `json:"TC"`
	RD       bool           `json:"RD"`
	RA       bool           `json:"RA"`
	AD       bool           `json:"AD"`
	CD       bool           `json:"CD"`
	Question []questionItem `json:"Question"`
	Answer   []answerItem   `json:"Answer"`
}

type questionItem struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

type answerItem struct {
	Name    string `json:"name"`
	Type    int    `json:"type"`
	TTL     int    `json:"TTL"`
	Expires string `json:"Expires"`
	Data    string `json:"data"`
}

var (
	//被屏蔽的那几个host
	mainHost = map[string]bool{
		"app-api.pixiv.net":      false,
		"oauth.secure.pixiv.net": false,
		"accounts.pixiv.net":     false,
	}
	//通用host
	universalMainHost = "pixivision.net"
	pixivMainIp       *string
)

func (dnsQueryRequest *dnsQueryRequest) fetch() (*dnsQueryResponse, error) {
	answer := &dnsQueryResponse{}
	_, isMainHost := mainHost[dnsQueryRequest.HostName]
	//如果是被屏蔽的那几个host
	if isMainHost {
		if pixivMainIp != nil {
			answer.Answer = []answerItem{{Data: *pixivMainIp, TTL: 50, Type: 1}}
			return answer, nil
		} else {
			dnsQueryRequest.HostName = universalMainHost
		}
	}

	dnsQueryUrl := fmt.Sprintf(
		"https://doh.dns.sb/dns-query?ct=application/dns-json&name=%s&type=A&do=false&cd=false",
		dnsQueryRequest.HostName,
	)

	var body []byte
	var err error
	//最多重试3次
	for i := 0; i < 3; i++ {
		body, err = request(dnsQueryUrl)
		if err == nil {
			break
		}
	}

	err = json.Unmarshal(body, answer)
	if err != nil {
		return nil, err
	}

	if isMainHost {
		pixivMainIp = &answer.Answer[0].Data
	}

	return answer, err
}

func request(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
