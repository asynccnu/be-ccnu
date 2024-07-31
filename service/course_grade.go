package service

import (
	"context"
	"encoding/json"
	"github.com/asynccnu/be-ccnu/domain"
	"github.com/asynccnu/be-ccnu/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GradeList struct {
	Items []GradeItem `json:"items"`
}

type GradeItem struct {
	//JgID               string `json:"jg_id"`
	Jsxm   string `json:"jsxm"`   // 教工名称
	Kch    string `json:"kch"`    // 课程号
	Kcmc   string `json:"kcmc"`   // 课程名称
	Kcxzmc string `json:"kcxzmc"` // 课程性质名称
	Kkbmmc string `json:"kkbmmc"` // 开课学院
	Xf     string `json:"xf"`     // 学分
	Cj     string `json:"cj"`     // 成绩
	JxbId  string `json:"jxb_id"`
	Jxbmc  string `json:"jxbmc"`
	Xnm    string `json:"xnm" binding:"required"`   // 学年名，如 2023
	Xqmmc  string `json:"xqmmc" binding:"required"` // 学期名称，如 1/2/3
}

func (c *ccnuService) GetSelfGradeList(ctx context.Context, studentId, password, year, term string) ([]domain.Grade, error) {
	client := c.getClientFromContext(ctx)
	if client == nil {
		var er error
		client, er = c.xkLoginClient(ctx, studentId, password) // 登录，直接
		if er != nil {
			return nil, er
		}
	}
	var termMap = map[string]string{"1": "3", "2": "12", "3": "16"} // 学期参数
	if year == "0" {
		year = ""
	}
	formData := url.Values{}
	formData.Set("xnm", year)          // 学年名
	formData.Set("xqm", termMap[term]) // 学期名
	formData.Set("kcbj", "")           //
	formData.Set("_search", "false")
	formData.Set("nd", string(time.Now().UnixNano()))
	formData.Set("queryModel.showCount", "1000")
	formData.Set("queryModel.currentPage", "1")
	formData.Set("queryModel.sortName", "")
	formData.Set("queryModel.sortOrder", "asc")
	formData.Set("time", "5")

	requestUrl := "https://xk.ccnu.edu.cn/jwglxt/cjcx/cjcx_cxXsgrcj.html?doType=query&gnmkdm=N305005"
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0")
	req.Header.Set("Referer", "http://xk.ccnu.edu.cn/jwglxt/cjcx/cjcx_cxDgXscj.html?gnmkdm=N305005&layout=default")
	req.Header.Set("Origin", "http://xk.ccnu.edu.cn")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var gl GradeList
	err = json.Unmarshal(body, &gl)
	if err != nil {
		return nil, err
	}

	res := slice.Map(gl.Items, func(idx int, src GradeItem) domain.Grade {
		credit, _ := strconv.ParseFloat(src.Xf, 10)
		total, _ := strconv.ParseFloat(src.Cj, 10)
		return domain.Grade{
			Course: domain.Course{
				CourseId: src.Kch,
				Name:     src.Kcmc,
				Teacher:  src.Jsxm,
				Class:    src.Jxbmc,
				School:   src.Kkbmmc,
				Property: src.Kcxzmc,
				Credit:   credit,
			},
			Total: total,
			Year:  src.Xnm,
			Term:  src.Xqmmc,
			JxbId: src.JxbId,
		}
	})
	return res, nil
}

func (c *ccnuService) GetDetailOfGradeList(ctx context.Context, studentId string, password string, year string, term string) ([]domain.Grade, error) {
	// 优化为一次登录，因为这里有很多请求，用一个client就行
	client, err := c.xkLoginClient(ctx, studentId, password)
	if err != nil {
		return nil, err
	}
	ctx = c.addClientToContext(ctx, client)
	gradeList, err := c.GetSelfGradeList(ctx, studentId, password, year, term) // A
	if err != nil {
		return nil, err
	}
	// 聚成绩的详细内容
	for i, grade := range gradeList {
		detail, er := c.getGradeDetail(ctx, studentId, password, grade.Year, grade.Term, grade.JxbId) // B
		if er != nil {
			return nil, er
		}
		// 0.平时 1.期末 2.总评
		if len(detail.Items) == 3 {
			gradeList[i].Regular, _ = strconv.ParseFloat(detail.Items[0].Xmcj, 64)
			gradeList[i].Final, _ = strconv.ParseFloat(detail.Items[1].Xmcj, 64)
			gradeList[i].Total, _ = strconv.ParseFloat(detail.Items[2].Xmcj, 64)
		} else {
			c.l.Info("条目长度有点怪", logger.Int("长度", len(detail.Items)))
			// 有这种课，军事理论就只有一个总评，没有平时成绩，期末成绩...，这里是查不到的，那就给一个只给一个总评分，平时分和期末都设置为总评分，
			gradeList[i].Regular = grade.Total
			gradeList[i].Final = grade.Total
		}
	}
	return gradeList, nil
}

type xkGradeListItem struct {
	Xmblmc string `json:"xmblmc"`
	Xmcj   string `json:"xmcj"`
}

type xkGradeListRespBody struct {
	Items []xkGradeListItem `json:"items"`
}

func (c *ccnuService) getGradeDetail(ctx context.Context, studentId, password string, year string, term string, jxbId string) (xkGradeListRespBody, error) {
	client := c.getClientFromContext(ctx)
	if client == nil {
		var er error
		client, er = c.xkLoginClient(ctx, studentId, password) // 登录，直接
		if er != nil {
			return xkGradeListRespBody{}, er
		}
	}
	var termMap = map[string]string{"1": "3", "2": "12", "3": "16"} // 学期参数
	// 准备请求参数
	formData := url.Values{}
	formData.Set("xnm", year)          // 学年, 留空
	formData.Set("xqm", termMap[term]) // 学期, 留空
	formData.Set("jxb_id", jxbId)
	formData.Set("_search", "false")
	formData.Set("nd", strconv.FormatInt(time.Now().Unix(), 10))
	formData.Set("queryModel.showCount", "1000")
	formData.Set("queryModel.currentPage", "1")
	formData.Set("queryModel.sortName", "xmblmc")
	formData.Set("queryModel.sortOrder", "asc")
	formData.Set("time", "3")

	// 请求URL
	requestUrl := "https://xk.ccnu.edu.cn/jwglxt/cjcx/cjcx_cxXsXmcjList.html?gnmkdm=N305007"
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return xkGradeListRespBody{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://xk.ccnu.edu.cn")
	req.Header.Set("Referer", "https://xk.ccnu.edu.cn/jwglxt/cjcx/cjcx_cxDgXsxmcj.html?gnmkdm=N305007&layout=default")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return xkGradeListRespBody{}, err
	}
	defer resp.Body.Close()

	// 读取并解析响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return xkGradeListRespBody{}, err
	}
	var gradeList xkGradeListRespBody // 此处定义合适的数据结构来解析JSON响应
	err = json.Unmarshal(body, &gradeList)
	if err != nil {
		return xkGradeListRespBody{}, err
	}

	return gradeList, err
}
