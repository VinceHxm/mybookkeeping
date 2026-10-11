package handler

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

func (a *API) ListFareRules(c *gin.Context) {
	list, err := a.fareRules.List(middleware.GetUserID(c))
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) CreateFareRule(c *gin.Context) {
	var req service.FareRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	r, err := a.fareRules.Create(middleware.GetUserID(c), req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, r)
}

func (a *API) UpdateFareRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.FareRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	r, err := a.fareRules.Update(middleware.GetUserID(c), id, req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, r)
}

func (a *API) DeleteFareRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.fareRules.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (a *API) PreviewFareRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Date          string `json:"date"`
		MonthSpentFen *int64 `json:"monthSpentFen"`
		BaseFen       *int64 `json:"baseFen"`
		RideCount     *int   `json:"rideCount"`
	}
	_ = c.ShouldBindJSON(&req)
	at := time.Now()
	if req.Date != "" {
		if t, err := time.ParseInLocation("2006-01-02", req.Date, time.Local); err == nil {
			at = t
		}
	}
	res, err := a.fareRules.Preview(middleware.GetUserID(c), id, at, req.MonthSpentFen, req.BaseFen, req.RideCount)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	localizeFarePreview(c, res)
	OK(c, res)
}

// 时段标签由「时段名 + 票卡/全价 + X折」拼成，时段名是用户数据，无法整句登记
var fareTimeWindowLabel = regexp.MustCompile(`^(.+?)((?:票卡|全价)\d+折)$`)

// localizeFarePreview 计价说明是「 · 」连接的多段中文，逐段按请求语言翻译；节假日名、时段名等数据原样保留
func localizeFarePreview(c *gin.Context, res *service.FarePreviewResult) {
	if res == nil || res.Reason == "" {
		return
	}
	loc := i18n.FromRequest(c)
	if loc == i18n.ZhCN {
		return
	}
	parts := strings.Split(res.Reason, " · ")
	for i, p := range parts {
		tr := i18n.T(loc, p)
		if tr == p {
			if m := fareTimeWindowLabel.FindStringSubmatch(p); m != nil {
				tr = m[1] + " " + i18n.T(loc, m[2])
			}
		}
		parts[i] = tr
	}
	res.Reason = strings.Join(parts, " · ")
	if res.HolidayName == "法定节假日" {
		res.HolidayName = i18n.T(loc, res.HolidayName)
	}
}
