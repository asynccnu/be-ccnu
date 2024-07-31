package service

import (
	"errors"
	"io"
	"net/http"
	"regexp"
)

type accountRequestParams struct {
	lt         string
	execution  string
	_eventId   string
	submit     string
	JSESSIONID string
}

func (c *ccnuService) makeAccountPreflightRequest() (*accountRequestParams, error) {
	var JSESSIONID string
	var lt string
	var execution string
	var _eventId string

	params := &accountRequestParams{}

	// 初始化 http request
	request, err := http.NewRequest("GET", "https://account.ccnu.edu.cn/cas/login", nil)
	if err != nil {
		return params, err
	}

	// 发起请求
	resp, err := c.client().Do(request)
	if err != nil {
		return params, err
	}

	// 读取 Body
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return params, err
	}

	// 获取 Cookie 中的 JSESSIONID
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			JSESSIONID = cookie.Value
		}
	}

	if JSESSIONID == "" {
		return params, errors.New("Can not get JSESSIONID")
	}

	// 正则匹配 HTML 返回的表单字段
	ltReg := regexp.MustCompile("name=\"lt\".+value=\"(.+)\"")
	executionReg := regexp.MustCompile("name=\"execution\".+value=\"(.+)\"")
	_eventIdReg := regexp.MustCompile("name=\"_eventId\".+value=\"(.+)\"")

	bodyStr := string(body)

	ltArr := ltReg.FindStringSubmatch(bodyStr)
	if len(ltArr) != 2 {
		return params, errors.New("Can not get form paramater: lt")
	}
	lt = ltArr[1]

	execArr := executionReg.FindStringSubmatch(bodyStr)
	if len(execArr) != 2 {
		return params, errors.New("Can not get form paramater: execution")
	}
	execution = execArr[1]

	_eventIdArr := _eventIdReg.FindStringSubmatch(bodyStr)
	if len(_eventIdArr) != 2 {
		return params, errors.New("Can not get form paramater: _eventId")
	}
	_eventId = _eventIdArr[1]

	params.lt = lt
	params.execution = execution
	params._eventId = _eventId
	params.submit = "LOGIN"
	params.JSESSIONID = JSESSIONID

	return params, nil
}
