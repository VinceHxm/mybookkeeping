package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mybookkeeping/internal/middleware"
	"mybookkeeping/internal/service"
)

func (a *API) ListTransactions(c *gin.Context) {
	q := service.TransactionQuery{Limit: 50}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Offset = n
		}
	}
	if v := c.Query("accountId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.AccountID = &n
		}
	}
	if v := c.Query("categoryId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.CategoryID = &n
		}
	}
	if v := c.Query("tagId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			q.TagID = &n
		}
	}
	q.Pending = c.Query("pending") == "1" || c.Query("pending") == "true"
	q.Voucher = c.Query("voucher")
	q.Type = c.Query("type")
	q.Keyword = c.Query("keyword")
	q.From = parseTimePtr(c.Query("from"))
	q.To = parseTimePtr(c.Query("to"))
	res, err := a.transactions.List(middleware.GetUserID(c), q)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxListAttachments(res.Items)
	}
	OK(c, res)
}

// RecentPlaces 最近使用过的地点（只返回地理字段）
func (a *API) RecentPlaces(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, err := a.transactions.RecentPlaces(middleware.GetUserID(c), limit)
	if err != nil {
		ServerError(c, err.Error())
		return
	}
	OK(c, list)
}

func (a *API) GetTransaction(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := a.transactions.Get(middleware.GetUserID(c), id)
	if err != nil {
		NotFound(c, "流水不存在")
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxAttachments(t)
	}
	OK(c, t)
}

type txBody struct {
	Type               string   `json:"type"`
	Amount             int64    `json:"amount"`
	AccountID          uint64   `json:"accountId"`
	ToAccountID        *uint64  `json:"toAccountId"`
	CategoryID         *uint64  `json:"categoryId"`
	Remark             string   `json:"remark"`
	GeoMode            string   `json:"geoMode"`
	GeoLng             *float64 `json:"geoLng"`
	GeoLat             *float64 `json:"geoLat"`
	GeoName            string   `json:"geoName"`
	GeoEndLng          *float64 `json:"geoEndLng"`
	GeoEndLat          *float64 `json:"geoEndLat"`
	GeoEndName         string   `json:"geoEndName"`
	FareRuleID         *uint64  `json:"fareRuleId"`
	InstallmentPeriods int      `json:"installmentPeriods"`
	InterestFen        int64    `json:"interestFen"`
	InstallmentDeferCycles int  `json:"installmentDeferCycles"`
	AmountPending      bool     `json:"amountPending"`
	ForeignCurrency    string   `json:"foreignCurrency"`
	ForeignAmount      int64    `json:"foreignAmount"`
	VoucherStatus      string   `json:"voucherStatus"`
	VoucherExpireOn    string   `json:"voucherExpireOn"`
	HappenedAt         string   `json:"happenedAt"`
	AttachmentIDs      []uint64 `json:"attachmentIds"`
	TagIDs             []uint64 `json:"tagIds"`
}

func (b txBody) toInput() (service.TransactionInput, error) {
	in := service.TransactionInput{
		Type: b.Type, Amount: b.Amount, AccountID: b.AccountID, ToAccountID: b.ToAccountID,
		CategoryID: b.CategoryID, Remark: b.Remark, GeoMode: b.GeoMode,
		GeoLng: b.GeoLng, GeoLat: b.GeoLat, GeoName: b.GeoName,
		GeoEndLng: b.GeoEndLng, GeoEndLat: b.GeoEndLat, GeoEndName: b.GeoEndName,
		FareRuleID:         b.FareRuleID,
		InstallmentPeriods: b.InstallmentPeriods, InterestFen: b.InterestFen,
		InstallmentDeferCycles: b.InstallmentDeferCycles,
		AmountPending: b.AmountPending, ForeignCurrency: b.ForeignCurrency, ForeignAmount: b.ForeignAmount,
		VoucherStatus: b.VoucherStatus, VoucherExpireOn: b.VoucherExpireOn,
		AttachmentIDs: b.AttachmentIDs, TagIDs: b.TagIDs,
	}
	if b.HappenedAt == "" {
		in.HappenedAt = time.Now()
	} else if t := parseTimePtr(b.HappenedAt); t != nil {
		in.HappenedAt = *t
	} else {
		return in, errors.New("时间格式无效")
	}
	return in, nil
}

func (a *API) CreateTransaction(c *gin.Context) {
	var body txBody
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in, err := body.toInput()
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	t, err := a.transactions.Create(middleware.GetUserID(c), in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxAttachments(t)
	}
	OK(c, t)
}

func (a *API) UpdateTransaction(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body txBody
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in, err := body.toInput()
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	t, err := a.transactions.Update(middleware.GetUserID(c), id, in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	if a.attachments != nil {
		a.attachments.RefreshTxAttachments(t)
	}
	OK(c, t)
}

// SetVoucherStatus 一键核销 / 撤销核销
func (a *API) SetVoucherStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	t, err := a.transactions.SetVoucherStatus(middleware.GetUserID(c), id, body.Status)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, t)
}

// RefundVoucher 团购券退款：生成退款收入并把券标记为已退款
func (a *API) RefundVoucher(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Amount     int64   `json:"amount"`
		AccountID  uint64  `json:"accountId"`
		CategoryID *uint64 `json:"categoryId"`
		HappenedAt string  `json:"happenedAt"`
		Remark     string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		BadRequest(c, "参数错误")
		return
	}
	in := service.VoucherRefundInput{
		Amount: body.Amount, AccountID: body.AccountID, CategoryID: body.CategoryID, Remark: body.Remark,
	}
	if body.HappenedAt != "" {
		t := parseTimePtr(body.HappenedAt)
		if t == nil {
			BadRequest(c, "时间格式无效")
			return
		}
		in.HappenedAt = *t
	}
	t, err := a.transactions.RefundVoucher(middleware.GetUserID(c), id, in)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, t)
}

func (a *API) DeleteTransaction(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := a.transactions.Delete(middleware.GetUserID(c), id); err != nil {
		BadRequest(c, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}
