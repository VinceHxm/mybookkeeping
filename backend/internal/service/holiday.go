package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mybookkeeping/internal/model"
)

const (
	holidayAPIBase = "https://timor.tech/api/holiday/year"
	// 热缓存仅加速读库；权威数据在 MySQL，不自动打远程
	holidayHotCacheTTL = 24 * time.Hour
	holidayAutoMetaKey = "holiday_auto_sync_year" // value = 已执行自动同步的公历年（12 月那一年）
)

// HolidayDay 单日节假日信息（法定放假 / 调休上班）
type HolidayDay struct {
	Date    string `json:"date"`
	Name    string `json:"name"`
	Holiday bool   `json:"holiday"`
	Wage    int    `json:"wage,omitempty"`
}

type HolidayYear struct {
	Year      int          `json:"year"`
	Days      []HolidayDay `json:"days"`
	Source    string       `json:"source,omitempty"`
	FetchedAt *time.Time   `json:"fetchedAt,omitempty"`
	FromDB    bool         `json:"fromDb"`
}

type HolidayService struct {
	db     *gorm.DB
	rdb    *redis.Client
	client *http.Client
	mu     sync.Mutex
	mem    map[int]*HolidayYear
}

func NewHolidayService(db *gorm.DB, rdb *redis.Client) *HolidayService {
	return &HolidayService{
		db:  db,
		rdb: rdb,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
		mem: map[int]*HolidayYear{},
	}
}

type timorYearResp struct {
	Code    int                      `json:"code"`
	Holiday map[string]timorDayEntry `json:"holiday"`
}

type timorDayEntry struct {
	Holiday bool   `json:"holiday"`
	Name    string `json:"name"`
	Wage    int    `json:"wage"`
	Date    string `json:"date"`
}

// Year 只读本地（内存→Redis→MySQL），不自动请求远程。
func (s *HolidayService) Year(ctx context.Context, year int) (*HolidayYear, error) {
	if year < 2007 || year > 2100 {
		return nil, fmt.Errorf("年份无效")
	}
	if y := s.fromMem(year); y != nil {
		return y, nil
	}
	if y := s.fromRedis(ctx, year); y != nil {
		s.toMem(y)
		return y, nil
	}
	y, err := s.fromDB(year)
	if err != nil {
		return nil, err
	}
	if y != nil {
		s.toMem(y)
		s.toRedis(ctx, y)
		return y, nil
	}
	// 未入库：返回空数据，调用方自行提示
	empty := &HolidayYear{Year: year, Days: []HolidayDay{}, FromDB: false}
	return empty, nil
}

// RefreshYear 从远程拉取并入库（管理员手动 / 定时任务）
func (s *HolidayService) RefreshYear(ctx context.Context, year int, source string) (*HolidayYear, error) {
	if year < 2007 || year > 2100 {
		return nil, fmt.Errorf("年份无效")
	}
	if source == "" {
		source = "manual"
	}
	y, err := s.fetchRemote(ctx, year)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	y.Source = source
	y.FetchedAt = &now
	y.FromDB = true
	if err := s.saveDB(y); err != nil {
		return nil, err
	}
	s.invalidateHot(ctx, year)
	s.toMem(y)
	s.toRedis(ctx, y)
	return y, nil
}

// TryAutoSync 每年 12 月第三周（15–21 日）自动拉取当年 + 下一年，全年只跑一次。
func (s *HolidayService) TryAutoSync(ctx context.Context, now time.Time) error {
	if !IsDecemberThirdWeek(now) {
		return nil
	}
	calYear := now.Year()
	if s.alreadyAutoSynced(calYear) {
		return nil
	}
	if _, err := s.RefreshYear(ctx, calYear, "auto"); err != nil {
		return fmt.Errorf("同步 %d: %w", calYear, err)
	}
	// 下一年可能尚未公布：尽力拉取，失败只记入返回信息供日志，仍写入「本年度已自动同步」标记
	nextErr := error(nil)
	if _, err := s.RefreshYear(ctx, calYear+1, "auto"); err != nil {
		nextErr = fmt.Errorf("同步 %d: %w（当年已入库，可稍后由管理员手动刷新下一年）", calYear+1, err)
	}
	if err := s.markAutoSynced(calYear); err != nil {
		return err
	}
	return nextErr
}

// IsDecemberThirdWeek 12 月 15–21 日视为第三周
func IsDecemberThirdWeek(t time.Time) bool {
	if t.Month() != time.December {
		return false
	}
	d := t.Day()
	return d >= 15 && d <= 21
}

// IsRestDay 是否法定放假休息日；无本地数据时视为非假期（不报错）
func (s *HolidayService) IsRestDay(ctx context.Context, at time.Time) (bool, string, error) {
	y, err := s.Year(ctx, at.Year())
	if err != nil {
		return false, "", err
	}
	if y == nil || len(y.Days) == 0 {
		return false, "", nil
	}
	key := at.Format("2006-01-02")
	for _, d := range y.Days {
		if d.Date == key {
			return d.Holiday, d.Name, nil
		}
	}
	return false, "", nil
}

func (s *HolidayService) fromDB(year int) (*HolidayYear, error) {
	if s.db == nil {
		return nil, nil
	}
	var rec model.HolidayYearRecord
	err := s.db.Where("year = ?", year).First(&rec).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var days []HolidayDay
	if rec.DaysJSON != "" {
		_ = json.Unmarshal([]byte(rec.DaysJSON), &days)
	}
	if days == nil {
		days = []HolidayDay{}
	}
	ft := rec.FetchedAt
	return &HolidayYear{
		Year:      rec.Year,
		Days:      days,
		Source:    rec.Source,
		FetchedAt: &ft,
		FromDB:    true,
	}, nil
}

func (s *HolidayService) saveDB(y *HolidayYear) error {
	if s.db == nil || y == nil {
		return fmt.Errorf("数据库未就绪")
	}
	b, err := json.Marshal(y.Days)
	if err != nil {
		return err
	}
	now := time.Now()
	if y.FetchedAt != nil {
		now = *y.FetchedAt
	}
	rec := model.HolidayYearRecord{
		Year:      y.Year,
		DaysJSON:  string(b),
		Source:    y.Source,
		FetchedAt: now,
		UpdatedAt: time.Now(),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "year"}},
		DoUpdates: clause.AssignmentColumns([]string{"days_json", "source", "fetched_at", "updated_at"}),
	}).Create(&rec).Error
}

func (s *HolidayService) alreadyAutoSynced(calYear int) bool {
	if s.db == nil {
		return false
	}
	var m model.AppMeta
	if err := s.db.Where("`key` = ?", holidayAutoMetaKey).First(&m).Error; err != nil {
		return false
	}
	return m.Value == fmt.Sprintf("%d", calYear)
}

func (s *HolidayService) markAutoSynced(calYear int) error {
	if s.db == nil {
		return nil
	}
	m := model.AppMeta{
		Key:       holidayAutoMetaKey,
		Value:     fmt.Sprintf("%d", calYear),
		UpdatedAt: time.Now(),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&m).Error
}

func (s *HolidayService) invalidateHot(ctx context.Context, year int) {
	s.mu.Lock()
	delete(s.mem, year)
	s.mu.Unlock()
	if s.rdb != nil {
		_ = s.rdb.Del(ctx, s.redisKey(year)).Err()
	}
}

func (s *HolidayService) fromMem(year int) *HolidayYear {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mem[year]
}

func (s *HolidayService) toMem(y *HolidayYear) {
	if y == nil {
		return
	}
	s.mu.Lock()
	s.mem[y.Year] = y
	s.mu.Unlock()
}

func (s *HolidayService) redisKey(year int) string {
	return fmt.Sprintf("mbk:holiday:year:%d", year)
}

func (s *HolidayService) fromRedis(ctx context.Context, year int) *HolidayYear {
	if s.rdb == nil {
		return nil
	}
	b, err := s.rdb.Get(ctx, s.redisKey(year)).Bytes()
	if err != nil || len(b) == 0 {
		return nil
	}
	var y HolidayYear
	if json.Unmarshal(b, &y) != nil || y.Year != year {
		return nil
	}
	return &y
}

func (s *HolidayService) toRedis(ctx context.Context, y *HolidayYear) {
	if s.rdb == nil || y == nil || !y.FromDB {
		return
	}
	b, err := json.Marshal(y)
	if err != nil {
		return
	}
	_ = s.rdb.Set(ctx, s.redisKey(y.Year), b, holidayHotCacheTTL).Err()
}

func (s *HolidayService) fetchRemote(ctx context.Context, year int) (*HolidayYear, error) {
	url := fmt.Sprintf("%s/%d/", holidayAPIBase, year)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "myBookkeeping/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("节假日接口请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("节假日接口 HTTP %d", resp.StatusCode)
	}
	var raw timorYearResp
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("节假日数据解析失败: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("节假日接口返回 code=%d", raw.Code)
	}
	out := &HolidayYear{Year: year, Days: make([]HolidayDay, 0, len(raw.Holiday))}
	for _, e := range raw.Holiday {
		date := e.Date
		if date == "" {
			continue
		}
		out.Days = append(out.Days, HolidayDay{
			Date:    date,
			Name:    e.Name,
			Holiday: e.Holiday,
			Wage:    e.Wage,
		})
	}
	if len(out.Days) == 0 {
		return nil, fmt.Errorf("%d 年暂无节假日数据（可能尚未公布）", year)
	}
	return out, nil
}
