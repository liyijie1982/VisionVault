package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"skybase/internal/config"
	"skybase/internal/domain/auth"
)

const SessionCookieName = "skybase_session"

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrInvalidVerificationCode = errors.New("invalid verification code")
var ErrUnauthorized = errors.New("unauthorized")

type LoginResult struct {
	User auth.User `json:"user"`
}

type CaptchaResult struct {
	Enabled bool   `json:"captchaEnabled"`
	UUID    string `json:"uuid"`
	Img     string `json:"img"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type AuthService struct {
	cfg         config.Config
	userService *UserService
	deptService *DepartmentService
	roleService *RoleService
	redis       *redis.Client
	ttl         time.Duration
	captchaTTL  time.Duration
}

type AuthSession struct {
	Token string
	User  auth.User
}

type authSessionPayload struct {
	UserID int64 `json:"userId"`
}

func NewAuthService(cfg config.Config, userService *UserService, deptService *DepartmentService, roleService *RoleService) *AuthService {
	ttl, err := time.ParseDuration(cfg.Redis.SessionTTL)
	if err != nil || ttl <= 0 {
		ttl = 12 * time.Hour
	}
	captchaTTL, err := time.ParseDuration(cfg.Auth.CaptchaTTL)
	if err != nil || captchaTTL <= 0 {
		captchaTTL = 5 * time.Minute
	}

	return &AuthService{
		cfg:         cfg,
		userService: userService,
		deptService: deptService,
		roleService: roleService,
		redis: redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		ttl:        ttl,
		captchaTTL: captchaTTL,
	}
}

func (s *AuthService) Ping(ctx context.Context) error {
	return s.redis.Ping(ctx).Err()
}

func (s *AuthService) EnsureBootstrap(ctx context.Context) error {
	departments, err := s.deptService.List(ctx)
	if err != nil {
		return err
	}

	deptID := int64(0)
	if len(departments) > 0 {
		deptID = departments[0].ID
	} else {
		dept, createErr := s.deptService.Create(ctx, DepartmentMutation{
			ParentID: 0,
			Name:     "System",
			Leader:   "System",
			Sort:     1,
			Status:   1,
		})
		if createErr != nil {
			return createErr
		}
		deptID = dept.ID
	}

	_, _, err = s.userService.GetByUsername(ctx, s.cfg.Auth.AdminUsername)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return err
	}

	_, err = s.userService.Create(ctx, UserMutation{
		DeptID:   deptID,
		Username: s.cfg.Auth.AdminUsername,
		Nickname: "System Administrator",
		RealName: "System Administrator",
		Phone:    "13900009999",
		Email:    "admin@skybase.local",
		Password: s.cfg.Auth.AdminPassword,
		Status:   1,
	})
	if err != nil {
		return err
	}

	userRecord, _, err := s.userService.GetByUsername(ctx, s.cfg.Auth.AdminUsername)
	if err != nil {
		return err
	}
	return s.userService.ChangePassword(ctx, userRecord.ID, s.cfg.Auth.AdminPassword)
}

func (s *AuthService) Login(username, password, uuid, code, loginIP string) (AuthSession, error) {
	if err := s.VerifyCaptcha(uuid, code); err != nil {
		return AuthSession{}, err
	}

	record, passwordHash, err := s.userService.GetByUsername(context.Background(), username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthSession{}, ErrInvalidCredentials
		}
		return AuthSession{}, err
	}
	if record.Status != 1 {
		return AuthSession{}, ErrInvalidCredentials
	}
	if err := comparePassword(passwordHash, password); err != nil {
		return AuthSession{}, ErrInvalidCredentials
	}

	token, err := randomToken(32)
	if err != nil {
		return AuthSession{}, err
	}

	payload, err := json.Marshal(authSessionPayload{UserID: record.ID})
	if err != nil {
		return AuthSession{}, err
	}
	if err := s.redis.Set(context.Background(), sessionKey(token), payload, s.ttl).Err(); err != nil {
		return AuthSession{}, err
	}

	now := time.Now()
	if err := s.userService.UpdateLoginMeta(context.Background(), record.ID, loginIP, now); err != nil {
		return AuthSession{}, err
	}

	user, err := s.userFromRecordID(context.Background(), record.ID)
	if err != nil {
		return AuthSession{}, err
	}

	return AuthSession{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) CurrentUser(token string) (auth.User, error) {
	if token == "" {
		return auth.User{}, ErrUnauthorized
	}

	payload, err := s.redis.Get(context.Background(), sessionKey(token)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return auth.User{}, ErrUnauthorized
		}
		return auth.User{}, err
	}

	var session authSessionPayload
	if err := json.Unmarshal(payload, &session); err != nil {
		return auth.User{}, ErrUnauthorized
	}

	user, err := s.userFromRecordID(context.Background(), session.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return auth.User{}, ErrUnauthorized
		}
		return auth.User{}, err
	}
	if err := s.redis.Expire(context.Background(), sessionKey(token), s.ttl).Err(); err != nil {
		return auth.User{}, err
	}

	return user, nil
}

func (s *AuthService) ChangePassword(token string, input ChangePasswordInput) (auth.User, error) {
	user, err := s.CurrentUser(token)
	if err != nil {
		return auth.User{}, err
	}

	record, passwordHash, err := s.userService.GetByUsername(context.Background(), user.Username)
	if err != nil {
		return auth.User{}, err
	}
	if err := comparePassword(passwordHash, input.CurrentPassword); err != nil {
		return auth.User{}, ErrCurrentPasswordIncorrect
	}
	if err := s.userService.ChangePassword(context.Background(), record.ID, input.NewPassword); err != nil {
		return auth.User{}, err
	}

	return s.userFromRecordID(context.Background(), record.ID)
}

func (s *AuthService) Logout(token string) {
	if token == "" {
		return
	}

	_ = s.redis.Del(context.Background(), sessionKey(token)).Err()
}

func (s *AuthService) SessionTTL() time.Duration {
	return s.ttl
}

func (s *AuthService) GenerateCaptcha() (CaptchaResult, error) {
	code, err := randomCaptchaCode(4)
	if err != nil {
		return CaptchaResult{}, err
	}

	uuid, err := randomToken(16)
	if err != nil {
		return CaptchaResult{}, err
	}

	if err := s.redis.Set(context.Background(), captchaKey(uuid), code, s.captchaTTL).Err(); err != nil {
		return CaptchaResult{}, err
	}

	return CaptchaResult{
		Enabled: true,
		UUID:    uuid,
		Img:     renderCaptcha(code),
	}, nil
}

func (s *AuthService) VerifyCaptcha(uuid, code string) error {
	uuid = trimSpace(uuid)
	code = trimSpace(code)
	if uuid == "" || code == "" {
		return ErrInvalidVerificationCode
	}

	key := captchaKey(uuid)
	stored, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrInvalidVerificationCode
		}
		return err
	}

	_ = s.redis.Del(context.Background(), key).Err()
	if !stringsEqualFold(stored, code) {
		return ErrInvalidVerificationCode
	}
	return nil
}

func (s *AuthService) userFromRecordID(ctx context.Context, id int64) (auth.User, error) {
	record, err := s.userService.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return auth.User{}, err
		}
		return auth.User{}, err
	}

	createdAt, _ := time.ParseInLocation(mysqlTimeFormat, record.CreatedAt, time.Local)
	updatedAt, _ := time.ParseInLocation(mysqlTimeFormat, record.UpdatedAt, time.Local)
	roleKeys, roleNames, menuKeys, err := s.roleService.FindUserAccess(ctx, record.ID)
	if err != nil {
		return auth.User{}, err
	}

	return auth.User{
		ID:                    record.ID,
		Username:              record.Username,
		Nickname:              record.Nickname,
		RealName:              record.RealName,
		Phone:                 record.Phone,
		Email:                 record.Email,
		Status:                record.Status,
		DeptID:                record.DeptID,
		RoleKeys:              roleKeys,
		RoleNames:             roleNames,
		MenuKeys:              menuKeys,
		PasswordResetRequired: record.PasswordResetRequired,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}, nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func sessionKey(token string) string {
	return "skybase:session:" + token
}

func captchaKey(uuid string) string {
	return "skybase:captcha:" + uuid
}

func randomCaptchaCode(size int) (string, error) {
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	buf := make([]byte, size)
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}

func renderCaptcha(code string) string {
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="132" height="44" viewBox="0 0 132 44"><rect width="132" height="44" rx="10" fill="#F5F7FA"/><path d="M8 30 C22 10, 38 38, 54 18 S84 8, 98 24 114 34, 124 12" stroke="#BED3EA" stroke-width="2" fill="none" opacity="0.9"/><path d="M10 14 C26 32, 42 8, 58 26 S92 36, 120 18" stroke="#D7A66A" stroke-width="1.6" fill="none" opacity="0.55"/><circle cx="22" cy="12" r="2" fill="#9BB8D3"/><circle cx="108" cy="32" r="2.4" fill="#DCC3A0"/><circle cx="92" cy="10" r="1.8" fill="#C8D8E8"/><text x="18" y="30" font-family="Verdana, Arial, sans-serif" font-size="24" font-weight="700" letter-spacing="7" fill="#1F2A37">%s</text></svg>`,
		html.EscapeString(code),
	)
	return base64.StdEncoding.EncodeToString([]byte(svg))
}

func trimSpace(value string) string {
	return strings.TrimSpace(value)
}

func stringsEqualFold(a, b string) bool {
	return strings.EqualFold(trimSpace(a), trimSpace(b))
}
