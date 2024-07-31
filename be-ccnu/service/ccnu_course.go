package service

import (
	"context"
	"encoding/json"
	"github.com/asynccnu/be-ccnu/domain"
	"github.com/ecodeclub/ekit/slice"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type OriginalCourses struct {
	Items []OriginalCourseItem `json:"items" binding:"required"`
}

// 课程数据信息
type OriginalCourseItem struct {
	Kch    string `json:"kch" binding:"required"`  // 课程号
	Kcmc   string `json:"kcmc" binding:"required"` // 课程名称
	Jsxx   string `json:"jsxx" binding:"required"` // 教师信息，格式如：2008980036/宋冰玉/讲师
	Xnm    string `json:"xnm" binding:"required"`  // 学年名，如 2019
	Xqmc   string `json:"xqmc" binding:"required"` // 学期名称，如 1/2/3
	Kkxymc string `json:"kkxymc"`                  // 开课学院
	// Kclbmc string `json:"kclbmc"`                  // 课程类别名称，如公共课/专业课
	Jxbmc  string `json:"jxbmc"`
	Kcxzmc string `json:"kcxzmc"` // 课程性质，如专业主干课程/通识必修课
	Xf     string `json:"xf"`
}

// GetSelfCourseList 个人课程列表
func (c *ccnuService) GetSelfCourseList(ctx context.Context, studentId, password, year, term string) ([]domain.Course, error) {
	// var termMap = map[string]string{"3": "1", "12": "2", "16": "3"} // 学期参数（逆向）
	originalCourses, err := c.getSelfCoursesFromXK(ctx, studentId, password, year, term)
	if err != nil {
		return nil, err
	}
	return slice.Map(originalCourses.Items, func(idx int, src OriginalCourseItem) domain.Course {
		credit, _ := strconv.ParseFloat(src.Xf, 10)
		return domain.Course{
			CourseId: src.Kch,
			Name:     src.Kcmc,
			Teacher:  c.getTeachersSqStrBySplitting(src.Jsxx),
			Class:    src.Jxbmc,
			School:   src.Kkxymc,
			Property: src.Kcxzmc,
			Credit:   credit,
			Year:     src.Xnm,
			Term:     src.Xqmc,
		}
	}), nil
}

func (c *ccnuService) getTeachersSqStrBySplitting(s string) string {
	sqs := strings.Split(s, ",")
	var teachers []string
	for _, s := range sqs {
		teachers = append(teachers, strings.Split(s, "/")[1])
	}
	return strings.Join(teachers, ",")
}

// getSelfCoursesFromXK 获取个人已上过的课程（教务系统原生结果）
func (c *ccnuService) getSelfCoursesFromXK(ctx context.Context, studentId, password string, year, term string) (OriginalCourses, error) {
	return c.makeCoursesGetRequest(ctx, studentId, password, year, term)
}

// makeCoursesGetRequest 请求获取课程列表
func (c *ccnuService) makeCoursesGetRequest(ctx context.Context, studentId, password, year, term string) (OriginalCourses, error) {
	var termMap = map[string]string{"1": "3", "2": "12", "3": "16"} // 学期参数
	if year == "0" {
		year = ""
	}

	formData := url.Values{}
	formData.Set("xnm", year)          // 学年名
	formData.Set("xqm", termMap[term]) // 学期名
	formData.Set("_search", "false")
	formData.Set("nd", string(time.Now().UnixNano()))
	formData.Set("queryModel.showCount", "1000")
	formData.Set("queryModel.currentPage", "1")
	formData.Set("queryModel.sortName", "")
	formData.Set("queryModel.sortOrder", "asc")
	formData.Set("time", "5")

	requestUrl := "http://xk.ccnu.edu.cn/jwglxt/xkcx/xkmdcx_cxXkmdcxIndex.html?doType=query&gnmkdm=N255010&su=" + studentId
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return OriginalCourses{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Origin", "http://xk.ccnu.edu.cn")
	req.Header.Set("Host", "xk.ccnu.edu.cn")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")

	client, err := c.xkLoginClient(ctx, studentId, password)
	resp, err := client.Do(req)
	if err != nil {
		return OriginalCourses{}, err
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return OriginalCourses{}, err
	}

	var data OriginalCourses
	if err := json.Unmarshal(body, &data); err != nil {
		return OriginalCourses{}, err
	}

	return data, nil
}

// xkLoginClient 教务系统模拟登录
func (c *ccnuService) xkLoginClient(ctx context.Context, studentId string, password string) (*http.Client, error) {
	client, err := c.loginClient(ctx, studentId, password)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", "https://account.ccnu.edu.cn/cas/login?service=http%3A%2F%2Fxk.ccnu.edu.cn%2Fsso%2Fpziotlogin", nil)
	if err != nil {
		return nil, err
	}

	_, err = client.Do(request)
	if err != nil {
		return nil, err
	}

	return client, nil
}
