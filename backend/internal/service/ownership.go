package service

import (
	"errors"

	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

// refIDs 一条记录引用的其他实体；nil 或 0 表示未引用
type refIDs struct {
	Account   *uint64
	ToAccount *uint64
	Category  *uint64
	FareRule  *uint64
}

// assertRefsOwned 校验引用的账户 / 分类 / 计费规则都属于 userID。
// 所有接收客户端 ID 并落库的写接口都必须调用，防止挂接他人数据后经 Preload 读出。
func assertRefsOwned(db *gorm.DB, userID uint64, r refIDs) error {
	checks := []struct {
		model any
		id    *uint64
		msg   string
	}{
		{&model.Account{}, r.Account, "账户不存在"},
		{&model.Account{}, r.ToAccount, "转入账户不存在"},
		{&model.Category{}, r.Category, "分类不存在"},
		{&model.FareRule{}, r.FareRule, "计费规则不存在"},
	}
	for _, c := range checks {
		if c.id == nil || *c.id == 0 {
			continue
		}
		var n int64
		if err := db.Model(c.model).Where("id = ? AND user_id = ?", *c.id, userID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return errors.New(c.msg)
		}
	}
	return nil
}

func txRefs(in TransactionInput) refIDs {
	acc := in.AccountID
	return refIDs{Account: &acc, ToAccount: in.ToAccountID, Category: in.CategoryID, FareRule: in.FareRuleID}
}
