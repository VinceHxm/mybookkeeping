package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybookkeeping/internal/model"
)

// 双用户越权测试：用户 A 拿用户 B 的各类 ID 调用写/读接口，必须全部失败且不影响 B 的数据。
// 新增接收 ID 的接口时，在这里补一条用例。

type tenantFixture struct {
	db                     *gorm.DB
	a, b                   uint64
	aAcc, bAcc, aCat, bCat uint64
	bTag, bFare, bTx, bAtt uint64
	bSchedule, bTemplate   uint64
}

func newTenantFixture(t *testing.T) *tenantFixture {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Account{}, &model.CreditInstallmentPlan{}, &model.Category{},
		&model.Tag{}, &model.Transaction{}, &model.TransactionTag{}, &model.Attachment{},
		&model.FareRule{}, &model.Template{}, &model.Schedule{},
	); err != nil {
		t.Fatal(err)
	}
	f := &tenantFixture{db: db}
	mk := func(v any) {
		t.Helper()
		if err := db.Create(v).Error; err != nil {
			t.Fatal(err)
		}
	}
	ua := &model.User{Username: "alice", PasswordHash: "x"}
	ub := &model.User{Username: "bob", PasswordHash: "x"}
	mk(ua)
	mk(ub)
	f.a, f.b = ua.ID, ub.ID

	aAcc := &model.Account{UserID: f.a, Name: "A 现金", Type: "cash"}
	bAcc := &model.Account{UserID: f.b, Name: "B 银行卡", Type: "bank", CardNo: "6222000011112222", HolderName: "鲍勃"}
	aCat := &model.Category{UserID: f.a, Name: "A 餐饮", Kind: "expense"}
	bCat := &model.Category{UserID: f.b, Name: "B 私密分类", Kind: "expense"}
	bTag := &model.Tag{UserID: f.b, Name: "B 标签"}
	bFare := &model.FareRule{UserID: f.b, Name: "B 公交", Enabled: true, CardRate: 1}
	for _, v := range []any{aAcc, bAcc, aCat, bCat, bTag, bFare} {
		mk(v)
	}
	f.aAcc, f.bAcc, f.aCat, f.bCat, f.bTag, f.bFare = aAcc.ID, bAcc.ID, aCat.ID, bCat.ID, bTag.ID, bFare.ID

	bTx, err := NewTransactionService(db).Create(f.b, TransactionInput{
		Type: "expense", Amount: 1000, AccountID: f.bAcc, CategoryID: &f.bCat,
		TagIDs: []uint64{f.bTag}, HappenedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	f.bTx = bTx.ID

	bAtt := &model.Attachment{UserID: f.b, ObjectKey: "u2/x.jpg", ContentType: "image/jpeg"}
	mk(bAtt)
	f.bAtt = bAtt.ID

	sch := &model.Schedule{UserID: f.b, Name: "B 房租", Type: "expense", Amount: 1, AccountID: f.bAcc, Frequency: "monthly", NextRunAt: time.Now().Add(time.Hour)}
	tpl := &model.Template{UserID: f.b, Name: "B 模板", Type: "expense"}
	mk(sch)
	mk(tpl)
	f.bSchedule, f.bTemplate = sch.ID, tpl.ID
	return f
}

func (f *tenantFixture) count(t *testing.T, m any, where string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := f.db.Model(m).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestTenantIsolation_Transactions(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTransactionService(f.db)

	if _, err := svc.Get(f.a, f.bTx); err == nil {
		t.Error("A 不应读到 B 的流水")
	}
	upd := TransactionInput{Type: "expense", Amount: 1, AccountID: f.aAcc, CategoryID: &f.aCat, HappenedAt: time.Now()}
	if _, err := svc.Update(f.a, f.bTx, upd); err == nil {
		t.Error("A 不应能修改 B 的流水")
	}
	if err := svc.Delete(f.a, f.bTx); err == nil {
		t.Error("A 不应能删除 B 的流水")
	}
	if f.count(t, &model.Transaction{}, "id = ?", f.bTx) != 1 {
		t.Fatal("B 的流水被破坏")
	}
	if res, err := svc.List(f.a, TransactionQuery{}); err != nil || res.Total != 0 {
		t.Errorf("A 的流水列表不应包含 B 的数据: res=%+v err=%v", res, err)
	}

	cases := []struct {
		name string
		in   TransactionInput
	}{
		{"他人账户", TransactionInput{Type: "expense", Amount: 1, AccountID: f.bAcc, CategoryID: &f.aCat}},
		{"他人转入账户", TransactionInput{Type: "transfer", Amount: 1, AccountID: f.aAcc, ToAccountID: &f.bAcc}},
		{"他人分类", TransactionInput{Type: "expense", Amount: 1, AccountID: f.aAcc, CategoryID: &f.bCat}},
		{"他人计费规则", TransactionInput{Type: "expense", Amount: 1, AccountID: f.aAcc, CategoryID: &f.aCat, FareRuleID: &f.bFare}},
	}
	for _, c := range cases {
		c.in.HappenedAt = time.Now()
		if _, err := svc.Create(f.a, c.in); err == nil {
			t.Errorf("创建流水引用%s应失败", c.name)
		}
	}
	var bAcc model.Account
	f.db.First(&bAcc, f.bAcc)
	if bAcc.Balance != -1000 {
		t.Errorf("B 的账户余额被改动: %d", bAcc.Balance)
	}

	// 引用他人附件 / 标签：流水可以建成，但不得挂接
	tx, err := svc.Create(f.a, TransactionInput{
		Type: "expense", Amount: 1, AccountID: f.aAcc, CategoryID: &f.aCat, HappenedAt: time.Now(),
		AttachmentIDs: []uint64{f.bAtt}, TagIDs: []uint64{f.bTag},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tx.Attachments) != 0 || len(tx.Tags) != 0 {
		t.Error("A 的流水不应挂上 B 的附件或标签")
	}
	if f.count(t, &model.Attachment{}, "id = ? AND transaction_id IS NULL", f.bAtt) != 1 {
		t.Error("B 的附件被挂到 A 的流水上")
	}
}

func TestTenantIsolation_Tags(t *testing.T) {
	f := newTenantFixture(t)
	svc := NewTagService(f.db)
	if err := svc.Delete(f.a, f.bTag); err == nil {
		t.Error("A 不应能删除 B 的标签")
	}
	if f.count(t, &model.TransactionTag{}, "tag_id = ?", f.bTag) != 1 {
		t.Error("B 流水上的标签关联被清除")
	}
	if _, err := svc.Update(f.a, f.bTag, "x", "", 0); err == nil {
		t.Error("A 不应能修改 B 的标签")
	}
}

func TestTenantIsolation_AccountsAndCategories(t *testing.T) {
	f := newTenantFixture(t)
	acc := NewAccountService(f.db)
	if _, _, err := acc.GetSensitive(f.a, f.bAcc); err == nil {
		t.Error("A 不应读到 B 的卡号明文")
	}
	_ = acc.Delete(f.a, f.bAcc)
	if f.count(t, &model.Account{}, "id = ?", f.bAcc) != 1 {
		t.Error("A 删除了 B 的账户")
	}
	if list, _ := acc.List(f.a, true); len(list) != 1 || list[0].ID != f.aAcc {
		t.Errorf("A 的账户列表应只有自己的账户: %+v", list)
	}

	cat := NewCategoryService(f.db)
	_ = cat.Delete(f.a, f.bCat)
	if f.count(t, &model.Category{}, "id = ?", f.bCat) != 1 {
		t.Error("A 删除了 B 的分类")
	}
	if _, err := cat.Create(f.a, CategoryInput{Name: "子类", Kind: "expense", ParentID: &f.bCat}); err == nil {
		t.Error("A 不应能把分类挂到 B 的父分类下")
	}
}

func TestTenantIsolation_SchedulesAndTemplates(t *testing.T) {
	f := newTenantFixture(t)
	sch := NewScheduleService(f.db, NewTransactionService(f.db))
	if _, err := sch.Create(f.a, ScheduleInput{Name: "x", Type: "expense", Amount: 1, AccountID: f.bAcc, Frequency: "daily"}); err == nil {
		t.Error("周期记账引用他人账户应失败")
	}
	if _, err := sch.Create(f.a, ScheduleInput{Name: "x", Type: "expense", Amount: 1, AccountID: f.aAcc, CategoryID: &f.bCat, Frequency: "daily"}); err == nil {
		t.Error("周期记账引用他人分类应失败")
	}
	if _, err := sch.Update(f.a, f.bSchedule, ScheduleInput{Name: "改名"}); err == nil {
		t.Error("A 不应能修改 B 的周期记账")
	}

	tpl := NewTemplateService(f.db, NewFareRuleService(f.db, nil))
	if _, err := tpl.Create(f.a, TemplateInput{Name: "x", CategoryID: &f.bCat}); err == nil {
		t.Error("模板引用他人分类应失败")
	}
	if _, err := tpl.Create(f.a, TemplateInput{Name: "x", AccountID: &f.bAcc}); err == nil {
		t.Error("模板引用他人账户应失败")
	}
	if _, err := tpl.Create(f.a, TemplateInput{Name: "x", FareRuleID: &f.bFare}); err == nil {
		t.Error("模板引用他人计费规则应失败")
	}
	if _, err := tpl.Update(f.a, f.bTemplate, TemplateInput{Name: "改名"}); err == nil {
		t.Error("A 不应能修改 B 的模板")
	}
}

func TestAttachmentSignedURL(t *testing.T) {
	f := newTenantFixture(t)
	ctx := context.Background()
	svc := NewAttachmentService(f.db, nil, []byte("test-secret"))
	att := &model.Attachment{ID: f.bAtt, UserID: f.b}
	u, err := url.Parse(svc.SignedURL(att))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(u.RawQuery, "token") {
		t.Error("签名链接不应包含会话令牌")
	}
	exp, sig := u.Query().Get("exp"), u.Query().Get("sig")
	expN, _ := strconv.ParseInt(exp, 10, 64)
	if sig != svc.sign(f.bAtt, f.b, expN) {
		t.Fatal("签名不一致")
	}

	past := time.Now().Add(-time.Minute).Unix()
	bad := []struct{ name, exp, sig string }{
		{"篡改签名", exp, strings.Repeat("0", len(sig))},
		{"篡改过期时间", strconv.FormatInt(expN+3600, 10), sig},
		{"已过期", strconv.FormatInt(past, 10), svc.sign(f.bAtt, f.b, past)},
		{"空签名", exp, ""},
	}
	for _, c := range bad {
		if _, _, _, err := svc.OpenSigned(ctx, f.bAtt, c.exp, c.sig); err != ErrAttachmentNotFound {
			t.Errorf("%s: 期望 ErrAttachmentNotFound，得到 %v", c.name, err)
		}
	}
	if _, _, _, err := svc.OpenSigned(ctx, f.bAtt+1, exp, sig); err != ErrAttachmentNotFound {
		t.Errorf("跨附件复用签名应失败: %v", err)
	}
}
