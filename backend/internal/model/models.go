package model

import "time"

type User struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	Username         string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash     string    `gorm:"size:255;not null" json:"-"`
	Email            string    `gorm:"size:128" json:"email"`
	EmailVerified    bool      `gorm:"not null;default:false" json:"emailVerified"`
	Role             string    `gorm:"size:16;not null;default:user;index" json:"role"` // user | admin
	Disabled         bool      `gorm:"not null;default:false;index" json:"disabled"`   // 停用后不可登录
	DefaultAccountID *uint64   `json:"defaultAccountId"`
	WeekStart        int       `gorm:"not null;default:1" json:"weekStart"` // 0=周日 1=周一
	ExpenseColor     string    `gorm:"size:16;default:#c45c3e" json:"expenseColor"`
	IncomeColor      string    `gorm:"size:16;default:#1b7f5a" json:"incomeColor"`
	Theme            string    `gorm:"size:16;default:light" json:"theme"` // light/dark/system
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

func (u *User) IsAdmin() bool {
	return u != nil && u.Role == RoleAdmin
}

// HolidayYearRecord 法定节假日按年入库（来源 timor.tech，定时或管理员手动刷新）
type HolidayYearRecord struct {
	Year      int       `gorm:"primaryKey" json:"year"`
	DaysJSON  string    `gorm:"type:longtext;not null" json:"-"`
	Source    string    `gorm:"size:32" json:"source"` // auto | manual
	FetchedAt time.Time `json:"fetchedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AppMeta 系统级键值（如节假日自动同步标记）
type AppMeta struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Account struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	UserID           uint64     `gorm:"index;not null" json:"userId"`
	Name             string     `gorm:"size:64;not null" json:"name"`
	Type             string     `gorm:"size:32;not null;default:cash" json:"type"` // cash/bank/credit/other
	Balance          int64      `gorm:"not null;default:0" json:"balance"`         // 分
	Sort             int        `gorm:"not null;default:0" json:"sort"`
	Archived         bool       `gorm:"not null;default:false" json:"archived"`
	CreditLimit      int64      `gorm:"not null;default:0" json:"creditLimit"` // 分，信用卡额度
	// CreditBilledFen 已出账金额（分）；未出账 = max(0, 已用额度 - CreditBilledFen)。仅信用账户有意义。
	CreditBilledFen  int64      `gorm:"not null;default:0" json:"creditBilledFen"`
	BillingDay       int        `gorm:"not null;default:0" json:"billingDay"`      // 1-28，0=未设
	PaymentDueDay    int        `gorm:"not null;default:0" json:"paymentDueDay"` // 1-28，还款日，0=未设
	Institution      string     `gorm:"size:64" json:"institution"`                // 开户行 / 发卡机构
	CardNo           string     `gorm:"size:64" json:"cardNo"`                     // 卡号（完整存，前端脱敏）
	HolderName       string     `gorm:"size:64" json:"holderName"`                 // 户名 / 持卡人
	StorageNote      string     `gorm:"size:512" json:"storageNote"`               // 存放位置文字
	Remark           string     `gorm:"size:512" json:"remark"`                    // 备注
	LastReconciledAt *time.Time `json:"lastReconciledAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`

	Attachments []Attachment `gorm:"foreignKey:AccountID" json:"attachments,omitempty"`
}

type Category struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	UserID    uint64     `gorm:"index;not null" json:"userId"`
	ParentID  *uint64    `gorm:"index" json:"parentId"` // nil=一级分类
	Name      string     `gorm:"size:64;not null" json:"name"`
	Kind      string     `gorm:"size:16;not null" json:"kind"` // expense/income
	Icon      string     `gorm:"size:64" json:"icon"`
	Sort      int        `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Children  []Category `gorm:"-" json:"children,omitempty"`
}

type Tag struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"userId"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Color     string    `gorm:"size:16;default:#1b7f5a" json:"color"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Transaction struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	UserID      uint64    `gorm:"index;not null" json:"userId"`
	Type        string    `gorm:"size:16;not null" json:"type"` // expense/income/transfer
	Amount      int64     `gorm:"not null" json:"amount"`       // 分，始终为正
	AccountID   uint64    `gorm:"index;not null" json:"accountId"`
	ToAccountID *uint64   `gorm:"index" json:"toAccountId"`
	CategoryID  *uint64   `gorm:"index" json:"categoryId"`
	Remark      string    `gorm:"size:512" json:"remark"`
	GeoMode     string    `gorm:"size:16;not null;default:point" json:"geoMode"` // point | route
	GeoLng      *float64  `json:"geoLng"`
	GeoLat      *float64  `json:"geoLat"`
	GeoName     string    `gorm:"size:255" json:"geoName"`
	GeoEndLng   *float64  `json:"geoEndLng"`
	GeoEndLat   *float64  `json:"geoEndLat"`
	GeoEndName  string    `gorm:"size:255" json:"geoEndName"`
	FareRuleID         *uint64   `gorm:"index" json:"fareRuleId"`
	InstallmentPeriods int       `gorm:"not null;default:1" json:"installmentPeriods"` // 1=不分期；>1 等额展示分期
	InterestFen        int64     `gorm:"not null;default:0" json:"interestFen"`         // 利息/手续费（分），仅展示拆分
	HappenedAt         time.Time `gorm:"index;not null" json:"happenedAt"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	Account     *Account     `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	ToAccount   *Account     `gorm:"foreignKey:ToAccountID" json:"toAccount,omitempty"`
	Category    *Category    `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:TransactionID" json:"attachments,omitempty"`
	Tags        []Tag        `gorm:"many2many:transaction_tags;" json:"tags,omitempty"`
}

type TransactionTag struct {
	TransactionID uint64 `gorm:"primaryKey"`
	TagID         uint64 `gorm:"primaryKey"`
}

type Attachment struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	UserID        uint64    `gorm:"index;not null" json:"userId"`
	TransactionID *uint64   `gorm:"index" json:"transactionId"`
	AccountID     *uint64   `gorm:"index" json:"accountId"` // 账户存放位置等照片
	ObjectKey     string    `gorm:"size:512;not null" json:"objectKey"`
	URL           string    `gorm:"size:1024" json:"url"`
	ContentType   string    `gorm:"size:128" json:"contentType"`
	Size          int64     `json:"size"`
	CreatedAt     time.Time `json:"createdAt"`
}

// Template 快捷记账模板
type Template struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	UserID       uint64    `gorm:"index;not null" json:"userId"`
	Name         string    `gorm:"size:64;not null" json:"name"`
	Type         string    `gorm:"size:16;not null" json:"type"`
	Amount       int64     `gorm:"not null;default:0" json:"amount"`
	AccountID    *uint64   `json:"accountId"`
	ToAccountID  *uint64   `json:"toAccountId"`
	CategoryID   *uint64   `json:"categoryId"`
	Remark       string    `gorm:"size:512" json:"remark"`
	GeoMode      string    `gorm:"size:16;not null;default:''" json:"geoMode"` // "" | point | route
	GeoLng       *float64  `json:"geoLng"`
	GeoLat       *float64  `json:"geoLat"`
	GeoName      string    `gorm:"size:255" json:"geoName"`
	GeoEndLng    *float64  `json:"geoEndLng"`
	GeoEndLat    *float64  `json:"geoEndLat"`
	GeoEndName   string    `gorm:"size:255" json:"geoEndName"`
	FareRuleID   *uint64   `gorm:"index" json:"fareRuleId"`
	FareRuleJSON string    `gorm:"type:text" json:"-"` // 旧数据迁移用，新逻辑不再写入
	TagIDsJSON   string    `gorm:"size:512" json:"-"`  // [1,2,3]
	Icon         string    `gorm:"size:64" json:"icon"`
	Sort         int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`

	TagIDs   []uint64  `gorm:"-" json:"tagIds"`
	FareRule *FareRule `gorm:"foreignKey:FareRuleID" json:"fareRule,omitempty"`
}

// FareRule 全局计费规则（公交卡等）
type FareRule struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	UserID          uint64    `gorm:"index;not null" json:"userId"`
	Name            string    `gorm:"size:64;not null" json:"name"`
	City            string    `gorm:"size:64" json:"city"` // 空=不限地市
	Note            string    `gorm:"size:255" json:"note"`
	ValidFrom       string    `gorm:"size:16" json:"validFrom"` // YYYY-MM-DD，空=不限
	ValidTo         string    `gorm:"size:16" json:"validTo"`
	BaseFen         int64     `gorm:"not null;default:0" json:"baseFen"`
	CardRate        float64   `gorm:"not null;default:1" json:"cardRate"` // 1=不打折, 0.9=九折
	AmountOffFen    int64     `gorm:"not null;default:0" json:"amountOffFen"` // 全局立减（分）
	StackMode       string    `gorm:"size:16;not null;default:prefer" json:"stackMode"` // prefer | lowest | stack
	FreeOnHoliday   bool      `gorm:"not null;default:false" json:"freeOnHoliday"`     // 法定节假日休息日免费
	FreePeriodsJSON string    `gorm:"type:text" json:"-"`
	TiersJSON       string    `gorm:"type:text" json:"-"`
	TimeWindowsJSON string    `gorm:"type:text" json:"-"`
	CountTiersJSON  string    `gorm:"type:text" json:"-"`
	CycleType       string    `gorm:"size:32;not null;default:calendar_month" json:"cycleType"` // calendar_month | from_day
	CycleStartDay   int       `gorm:"not null;default:1" json:"cycleStartDay"`                  // from_day: 1-28
	// CycleSeedFen 本周期手工补录累计（分）；与流水合计后用于阶梯。仅当 CycleSeedKey 等于当前周期起点时生效。
	CycleSeedFen   int64  `gorm:"not null;default:0" json:"cycleSeedFen"`
	CycleSeedKey   string `gorm:"size:16" json:"cycleSeedKey"` // YYYY-MM-DD，对应周期起点
	CycleCountSeed int    `gorm:"not null;default:0" json:"cycleCountSeed"` // 本周期补录乘次
	Enabled        bool   `gorm:"not null;default:true" json:"enabled"`
	Icon           string `gorm:"size:64" json:"icon"`
	Sort           int    `gorm:"not null;default:0" json:"sort"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`

	FreePeriods []FareFreePeriod `gorm:"-" json:"freePeriods"`
	Tiers       []FareTier       `gorm:"-" json:"tiers"`
	TimeWindows []FareTimeWindow `gorm:"-" json:"timeWindows"`
	CountTiers  []FareCountTier  `gorm:"-" json:"countTiers"`
}

type FareFreePeriod struct {
	StartDate string `json:"startDate"` // YYYY-MM-DD
	EndDate   string `json:"endDate"`
	Recur     string `json:"recur"` // none | yearly | weekly
	Weekdays  []int  `json:"weekdays,omitempty"` // weekly: 0=Sun … 6=Sat
}

type FareTier struct {
	// MinFen 累计超过该值（不含）后生效；兼容旧字段 thresholdFen
	MinFen int64 `json:"minFen"`
	// MaxFen 累计不超过该值（含）；0 表示无上限
	MaxFen int64 `json:"maxFen"`
	Rate   float64 `json:"rate"`
	// Base: full=全价×rate；card=票卡价×rate（rate 通常为 1 表示恢复票卡折）
	Base string `json:"base"`
	// ThresholdFen 旧版字段，仅反序列化迁移用
	ThresholdFen int64 `json:"thresholdFen,omitempty"`
}

// FareTimeWindow 日内时段优惠（如工作日 7 点前）
type FareTimeWindow struct {
	Name     string  `json:"name,omitempty"`
	StartHM  string  `json:"startHm"` // HH:MM，含
	EndHM    string  `json:"endHm"`   // HH:MM，不含；可跨午夜
	Weekdays []int   `json:"weekdays,omitempty"` // 空=每天；0=Sun…6=Sat
	Rate     float64 `json:"rate"`               // 折扣；Free=true 时忽略
	Free     bool    `json:"free"`
	Base     string  `json:"base"` // full | card，默认 card
}

// FareCountTier 按乘次阶梯（如第 40 次免费）
type FareCountTier struct {
	// MinCount 本周期已乘次数超过该值（不含）后，下一次命中；第 40 次 → MinCount=39
	MinCount     int     `json:"minCount"`
	MaxCount     int     `json:"maxCount"` // 本趟序号至（含）；0=无上限
	Rate         float64 `json:"rate"`
	Base         string  `json:"base"` // full | card
	Free         bool    `json:"free"`
	AmountOffFen int64   `json:"amountOffFen"` // 本档立减
}

// Schedule 周期记账
// Frequency: daily / weekly / monthly / yearly / every_n_days
type Schedule struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	UserID      uint64     `gorm:"index;not null" json:"userId"`
	Name        string     `gorm:"size:64;not null" json:"name"`
	Type        string     `gorm:"size:16;not null" json:"type"`
	Amount      int64      `gorm:"not null" json:"amount"`
	AccountID   uint64     `gorm:"not null" json:"accountId"`
	ToAccountID *uint64    `json:"toAccountId"`
	CategoryID  *uint64    `json:"categoryId"`
	Remark      string     `gorm:"size:512" json:"remark"`
	TagIDsJSON  string     `gorm:"size:512" json:"-"`
	Frequency   string     `gorm:"size:32;not null" json:"frequency"`
	IntervalN   int        `gorm:"not null;default:1" json:"intervalN"` // every_n_days 的 N；weekly 可表示每 N 周
	Enabled     bool       `gorm:"not null;default:true" json:"enabled"`
	NextRunAt   time.Time  `gorm:"index;not null" json:"nextRunAt"`
	EndAt       *time.Time `json:"endAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	TagIDs []uint64 `gorm:"-" json:"tagIds"`
}
