package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mybookkeeping/internal/i18n"
	"mybookkeeping/internal/model"
)

func newCategoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Category{}, &model.Transaction{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func categoryByKey(t *testing.T, db *gorm.DB, userID uint64, key string) model.Category {
	t.Helper()
	var c model.Category
	if err := db.Where("user_id = ? AND preset_key = ? AND parent_id IS NULL", userID, key).First(&c).Error; err != nil {
		t.Fatalf("找不到预设 %q: %v", key, err)
	}
	return c
}

func countCategories(db *gorm.DB, userID uint64) int64 {
	var n int64
	db.Model(&model.Category{}).Where("user_id = ?", userID).Count(&n)
	return n
}

func TestCategoryPresetFollowsLocale(t *testing.T) {
	db := newCategoryTestDB(t)
	svc := NewCategoryService(db)
	const uid = 1

	if err := svc.EnsureDefaults(uid, i18n.En); err != nil {
		t.Fatal(err)
	}
	total := countCategories(db, uid)
	food := categoryByKey(t, db, uid, "餐饮")
	if food.Name != "Food & Dining" || food.NameZh != "餐饮" || food.NameJa != "食費" || food.NameKo != "식비" {
		t.Fatalf("四语名未入库: %+v", food)
	}

	if err := svc.EnsureDefaults(uid, i18n.Ja); err != nil {
		t.Fatal(err)
	}
	if got := categoryByKey(t, db, uid, "餐饮").Name; got != "食費" {
		t.Fatalf("切日文后应为 食費，实际 %q", got)
	}
	if n := countCategories(db, uid); n != total {
		t.Fatalf("切换语言不应新增分类: %d -> %d", total, n)
	}

	// 改名后固定用自定义名，且不会被补齐重新加回
	traffic := categoryByKey(t, db, uid, "交通")
	if _, err := svc.Update(uid, traffic.ID, CategoryInput{Name: "通勤", Icon: traffic.Icon, Sort: traffic.Sort}); err != nil {
		t.Fatal(err)
	}
	if err := svc.EnsureDefaults(uid, i18n.Ko); err != nil {
		t.Fatal(err)
	}
	if got := categoryByKey(t, db, uid, "交通").Name; got != "通勤" {
		t.Fatalf("改过名的分类不应被翻译覆盖，实际 %q", got)
	}
	if got := categoryByKey(t, db, uid, "餐饮").Name; got != "식비" {
		t.Fatalf("切韩文后应为 식비，实际 %q", got)
	}
	if n := countCategories(db, uid); n != total {
		t.Fatalf("改名后补齐不应重复添加: %d -> %d", total, n)
	}
}

func TestCategoryPresetBackfillLegacy(t *testing.T) {
	db := newCategoryTestDB(t)
	svc := NewCategoryService(db)
	const uid = 2

	// 模拟老用户：中文名、无四语字段；其中一个二级改过名
	root := model.Category{UserID: uid, Name: "餐饮", Kind: "expense", Sort: 1}
	db.Create(&root)
	pid := root.ID
	db.Create(&model.Category{UserID: uid, Name: "外卖", Kind: "expense", ParentID: &pid, Sort: 1})
	db.Create(&model.Category{UserID: uid, Name: "我的早餐", Kind: "expense", ParentID: &pid, Sort: 2})

	if err := svc.EnsureDefaults(uid, i18n.En); err != nil {
		t.Fatal(err)
	}
	var takeout, custom model.Category
	db.Where("user_id = ? AND parent_id = ? AND preset_key = ?", uid, pid, "外卖").First(&takeout)
	db.Where("user_id = ? AND name = ?", uid, "我的早餐").First(&custom)
	if takeout.Name != "Takeout" || takeout.NameZh != "外卖" {
		t.Fatalf("老数据应回填并切到英文: %+v", takeout)
	}
	if custom.PresetKey != "" || custom.NameEn != "" {
		t.Fatalf("自建分类不应被识别为预设: %+v", custom)
	}
	if got := categoryByKey(t, db, uid, "餐饮").Name; got != "Food & Dining" {
		t.Fatalf("老一级分类应切到英文，实际 %q", got)
	}
}
