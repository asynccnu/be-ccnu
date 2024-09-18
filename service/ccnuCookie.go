package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (s *ccnuService) GetCCNUCookie(ctx context.Context, studentId string, password string) (string, error) {
	var cookie string
	var err error
	if CheckIsUndergraduate(studentId) {
		cookie, err = BKSloginCCNU(studentId, password)
		if err != nil {
			return "", fmt.Errorf("failed to login: %v", err)
		}
	}
	return cookie, nil
}

// CheckIsUndergraduate 检查该学号是否是本科生
func CheckIsUndergraduate(stuId string) bool {
	return stuId[4] == '2'
	//区分是学号第五位，本科是2，硕士是1，博士是0，工号是6或9
}

// BKSloginCCNU 模拟本科生登录CCNU并返回Cookie
func BKSloginCCNU(username, password string) (string, error) {
	loginURL := "https://account.ccnu.edu.cn/cas/login" // 真实的登录URL
	data := fmt.Sprintf("username=%s&password=%s", username, password)
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var value string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			value = cookie.Value
		}
	}
	return fmt.Sprintf("JSESSIONID=%s", value), nil
}

//研究生登陆,暂时放弃
//func YJSloginCCNU(id, mm string) (ccnuCookie string, err error) {
//	client := &http.Client{}
//	mp := make(map[string]string)
//	str := fmt.Sprintf("csrftoken=&yhm=%s&mm=%s", id, mm)
//	var data = strings.NewReader(str)
//	timestamp := time.Now().Unix()
//	url := fmt.Sprintf("https://grd.ccnu.edu.cn/yjsxt/xtgl/login_slogin.html?time=%d", timestamp)
//	req, err := http.NewRequest("POST", url, data)
//	if err != nil {
//		return "", err
//	}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//	defer resp.Body.Close()
//	for _, v := range resp.Cookies() {
//		mp[v.Name] = v.Value
//	}
//	ccnuCookie = fmt.Sprintf("JSESSIONID=%s; route=%s", mp["JSESSIONID"], mp["route"])
//	return ccnuCookie, nil
//}
