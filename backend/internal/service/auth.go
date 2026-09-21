package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"mybookkeeping/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserExists         = errors.New("用户名已存在")
	ErrWeakPassword       = errors.New("密码至少 6 位")
	ErrBadResetToken      = errors.New("重置令牌无效或已过期")
	ErrUserDisabled       = errors.New("账户已停用，请联系管理员")
	ErrCannotModifySelf   = errors.New("不能对自己执行此操作")
	ErrLastAdmin          = errors.New("不能停用或降级最后一个管理员")
	ErrInvalidRole        = errors.New("角色无效，仅支持 user / admin")
)

type AuthService struct {
	db                 *gorm.DB
	rdb                *redis.Client
	ttl                time.Duration
	resetTokenTTL      time.Duration
	resetTokenInResp   bool
}

func NewAuthService(db *gorm.DB, rdb *redis.Client, ttl time.Duration, resetTTLHours int, resetInResp bool) *AuthService {
	if resetTTLHours <= 0 {
		resetTTLHours = 2
	}
	return &AuthService{
		db: db, rdb: rdb, ttl: ttl,
		resetTokenTTL:    time.Duration(resetTTLHours) * time.Hour,
		resetTokenInResp: resetInResp,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (token string, user *model.User, err error) {
	var u model.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}
	if u.Disabled {
		return "", nil, ErrUserDisabled
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}
	token, err = randomToken(32)
	if err != nil {
		return "", nil, err
	}
	uidStr := strconv.FormatUint(u.ID, 10)
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, "sess:"+token, uidStr, s.ttl)
	pipe.SAdd(ctx, "usess:"+uidStr, token)
	pipe.Expire(ctx, "usess:"+uidStr, s.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", nil, fmt.Errorf("redis set session: %w", err)
	}
	return token, &u, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	uidStr, err := s.rdb.Get(ctx, "sess:"+token).Result()
	_ = s.rdb.Del(ctx, "sess:"+token).Err()
	if err == nil && uidStr != "" {
		_ = s.rdb.SRem(ctx, "usess:"+uidStr, token).Err()
	}
	return nil
}

// RevokeUserSessions 踢掉该用户全部会话（停用/删除时）
func (s *AuthService) RevokeUserSessions(ctx context.Context, userID uint64) error {
	key := "usess:" + strconv.FormatUint(userID, 10)
	tokens, err := s.rdb.SMembers(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if len(tokens) > 0 {
		pipe := s.rdb.Pipeline()
		for _, t := range tokens {
			pipe.Del(ctx, "sess:"+t)
		}
		pipe.Del(ctx, key)
		_, _ = pipe.Exec(ctx)
	}
	return nil
}

func (s *AuthService) GetUser(id uint64) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *AuthService) CreateUser(username, password string) (*model.User, error) {
	return s.Register(username, password, "")
}

func (s *AuthService) Register(username, password, email string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(password) < 6 {
		return nil, ErrWeakPassword
	}
	var n int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&n)
	if n > 0 {
		return nil, ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		Role:         model.RoleUser,
		WeekStart:    1,
		ExpenseColor: "#c45c3e",
		IncomeColor:  "#1b7f5a",
		Theme:        "light",
	}
	if err := s.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

// UpsertUserForCLI 供 init-user：存在则重置密码；CLI 创建的用户默认 admin
func (s *AuthService) UpsertUserForCLI(username, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var existing model.User
	err = s.db.Where("username = ?", username).First(&existing).Error
	if err == nil {
		updates := map[string]any{
			"password_hash": string(hash),
			"role":          model.RoleAdmin,
			"disabled":      false,
		}
		if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
			return nil, err
		}
		existing.Role = model.RoleAdmin
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	u := &model.User{
		Username: username, PasswordHash: string(hash), Role: model.RoleAdmin,
		WeekStart: 1, ExpenseColor: "#c45c3e", IncomeColor: "#1b7f5a", Theme: "light",
	}
	if err := s.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

// EnsureAdminRoles 迁移：空角色补 user；若系统尚无 admin，则提升 username=admin 或最早用户
func (s *AuthService) EnsureAdminRoles() error {
	_ = s.db.Model(&model.User{}).
		Where("role = '' OR role IS NULL").
		Update("role", model.RoleUser).Error
	var n int64
	if err := s.db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	var u model.User
	err := s.db.Where("username = ?", "admin").First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = s.db.Order("id asc").First(&u).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 尚无用户
		}
		return err
	}
	return s.db.Model(&u).Update("role", model.RoleAdmin).Error
}

func (s *AuthService) ChangePassword(userID uint64, oldPwd, newPwd string) error {
	if len(newPwd) < 6 {
		return ErrWeakPassword
	}
	var u model.User
	if err := s.db.First(&u, userID).Error; err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPwd)) != nil {
		return errors.New("原密码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&u).Update("password_hash", string(hash)).Error
}

type ForgotResult struct {
	Message    string `json:"message"`
	ResetToken string `json:"resetToken,omitempty"`
}

func (s *AuthService) ForgotPassword(ctx context.Context, username string) (*ForgotResult, error) {
	username = strings.TrimSpace(username)
	var u model.User
	if err := s.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 防枚举：仍返回成功文案
			return &ForgotResult{Message: "若用户存在，已生成重置令牌"}, nil
		}
		return nil, err
	}
	token, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	if err := s.rdb.Set(ctx, "reset:"+token, strconv.FormatUint(u.ID, 10), s.resetTokenTTL).Err(); err != nil {
		return nil, err
	}
	out := &ForgotResult{Message: "已生成重置令牌（个人部署默认直接返回，请尽快使用）"}
	if s.resetTokenInResp {
		out.ResetToken = token
	}
	return out, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPwd string) error {
	if len(newPwd) < 6 {
		return ErrWeakPassword
	}
	uidStr, err := s.rdb.Get(ctx, "reset:"+token).Result()
	if err != nil {
		return ErrBadResetToken
	}
	uid, _ := strconv.ParseUint(uidStr, 10, 64)
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.db.Model(&model.User{}).Where("id = ?", uid).Update("password_hash", string(hash)).Error; err != nil {
		return err
	}
	_ = s.rdb.Del(ctx, "reset:"+token).Err()
	return nil
}

type UserSettingsInput struct {
	Email            *string
	DefaultAccountID **uint64 // 指针的指针：nil=不改；非 nil 且 *nil=清空
	WeekStart        *int
	ExpenseColor     *string
	IncomeColor      *string
	Theme            *string
}

func (s *AuthService) UpdateSettings(userID uint64, in UserSettingsInput) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, userID).Error; err != nil {
		return nil, err
	}
	if in.Email != nil {
		u.Email = strings.TrimSpace(*in.Email)
		u.EmailVerified = false
	}
	if in.DefaultAccountID != nil {
		u.DefaultAccountID = *in.DefaultAccountID
	}
	if in.WeekStart != nil {
		u.WeekStart = *in.WeekStart
	}
	if in.ExpenseColor != nil && *in.ExpenseColor != "" {
		u.ExpenseColor = *in.ExpenseColor
	}
	if in.IncomeColor != nil && *in.IncomeColor != "" {
		u.IncomeColor = *in.IncomeColor
	}
	if in.Theme != nil && *in.Theme != "" {
		u.Theme = *in.Theme
	}
	if err := s.db.Save(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// AdminUserView 管理员列表项
type AdminUserView struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Disabled  bool      `json:"disabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *AuthService) ListUsers() ([]AdminUserView, error) {
	var list []model.User
	if err := s.db.Order("id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]AdminUserView, 0, len(list))
	for _, u := range list {
		out = append(out, AdminUserView{
			ID: u.ID, Username: u.Username, Email: u.Email,
			Role: u.Role, Disabled: u.Disabled,
			CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		})
	}
	return out, nil
}

type AdminUserUpdateInput struct {
	Role     *string
	Disabled *bool
}

func (s *AuthService) AdminUpdateUser(ctx context.Context, actorID, targetID uint64, in AdminUserUpdateInput) (*model.User, error) {
	if actorID == targetID && (in.Role != nil || in.Disabled != nil) {
		return nil, ErrCannotModifySelf
	}
	var u model.User
	if err := s.db.First(&u, targetID).Error; err != nil {
		return nil, err
	}
	if in.Role != nil {
		role := strings.TrimSpace(*in.Role)
		if role != model.RoleUser && role != model.RoleAdmin {
			return nil, ErrInvalidRole
		}
		if u.Role == model.RoleAdmin && role != model.RoleAdmin {
			if err := s.ensureNotLastAdmin(u.ID); err != nil {
				return nil, err
			}
		}
		u.Role = role
	}
	if in.Disabled != nil {
		if *in.Disabled && u.IsAdmin() {
			if err := s.ensureNotLastAdmin(u.ID); err != nil {
				return nil, err
			}
		}
		u.Disabled = *in.Disabled
	}
	if err := s.db.Save(&u).Error; err != nil {
		return nil, err
	}
	if u.Disabled {
		_ = s.RevokeUserSessions(ctx, u.ID)
	}
	return &u, nil
}

func (s *AuthService) AdminDeleteUser(ctx context.Context, actorID, targetID uint64) error {
	if actorID == targetID {
		return ErrCannotModifySelf
	}
	var u model.User
	if err := s.db.First(&u, targetID).Error; err != nil {
		return err
	}
	if u.IsAdmin() {
		if err := s.ensureNotLastAdmin(u.ID); err != nil {
			return err
		}
	}
	_ = s.RevokeUserSessions(ctx, targetID)
	return s.db.Transaction(func(tx *gorm.DB) error {
		uid := targetID
		var txIDs []uint64
		_ = tx.Model(&model.Transaction{}).Where("user_id = ?", uid).Pluck("id", &txIDs).Error
		if len(txIDs) > 0 {
			_ = tx.Where("transaction_id IN ?", txIDs).Delete(&model.TransactionTag{}).Error
			_ = tx.Where("transaction_id IN ?", txIDs).Delete(&model.Attachment{}).Error
		}
		_ = tx.Where("user_id = ?", uid).Delete(&model.Attachment{}).Error
		if err := tx.Where("user_id = ?", uid).Delete(&model.Transaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.Schedule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.Template{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.FareRule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.Tag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.Category{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.Account{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, uid).Error
	})
}

func (s *AuthService) ensureNotLastAdmin(excludeID uint64) error {
	var n int64
	q := s.db.Model(&model.User{}).Where("role = ? AND disabled = ?", model.RoleAdmin, false)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return ErrLastAdmin
	}
	return nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
