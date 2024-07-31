package service

import (
	"context"
	"errors"
	ccnuv1 "github.com/asynccnu/be-api/gen/proto/ccnu/v1"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func (c *ccnuService) Login(ctx context.Context, studentId string, password string) (bool, error) {
	client, err := c.loginClient(ctx, studentId, password)
	return client != nil, err
}

func (c *ccnuService) client() *http.Client {
	j, _ := cookiejar.New(&cookiejar.Options{})
	return &http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Jar:     j,
		Timeout: c.timeout,
	}
}

func (c *ccnuService) loginClient(ctx context.Context, studentId string, password string) (*http.Client, error) {
	params, err := c.makeAccountPreflightRequest()
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("username", studentId)
	v.Set("password", password)
	v.Set("lt", params.lt)
	v.Set("execution", params.execution)
	v.Set("_eventId", params._eventId)
	v.Set("submit", params.submit)

	request, err := http.NewRequest("POST", "https://account.ccnu.edu.cn/cas/login;jsessionid="+params.JSESSIONID, strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.109 Safari/537.36")
	request.WithContext(ctx)

	client := c.client()
	resp, err := client.Do(request)
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return nil, ccnuv1.ErrorNetworkToXkError("网络异常")
		}
		return nil, err
	}
	if len(resp.Header.Get("Set-Cookie")) == 0 {
		return nil, ccnuv1.ErrorInvalidSidOrPwd("学号或密码错误")
	}
	return client, nil
}

type ClientKey struct{} // 用于 context 的键

// 将 http.Client 添加到 context 中
func (c *ccnuService) addClientToContext(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, ClientKey{}, client)
}

// 从 context 中获取 http.Client
func (c *ccnuService) getClientFromContext(ctx context.Context) *http.Client {
	client, ok := ctx.Value(ClientKey{}).(*http.Client)
	if !ok {
		return nil // 这里可以处理默认逻辑或错误
	}
	return client
}
