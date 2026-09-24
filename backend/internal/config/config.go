package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr string

	MySQLDSN string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	// AppSecret 用于签发附件短期访问链接（HMAC）；为空时每次启动随机生成
	AppSecret string
	// TrustedProxies 反向代理地址/网段，用于从 X-Forwarded-For 取真实客户端 IP（登录限流依赖）
	TrustedProxies []string
	// CORSOrigins 允许跨域的前端源；为空=不下发 CORS 头（同源部署 / Vite 代理无需配置）
	CORSOrigins []string

	AmapKey          string
	AmapSecurityCode string

	SessionTTL time.Duration

	AllowRegister        bool
	ResetTokenInResponse bool
	ResetTokenTTLHours   int

	DeepSeekAPIKey      string
	DeepSeekBaseURL     string
	DeepSeekModel       string // 文字识别，如 deepseek-chat
	DeepSeekVisionModel string // 识图，如 deepseek-flash

	// 实际加载到的配置文件路径（空表示未找到，仅用默认值/系统环境变量）
	LoadedFrom string
}

// Load 优先级（高 → 低）：
//  1. 进程已有环境变量（系统/容器注入，适合临时覆盖）
//  2. 项目配置文件（默认 backend/.env，仅作用于本进程，不改系统）
//  3. 代码内默认值
//
// 配置文件路径可通过 CONFIG_FILE 指定；否则按常见位置自动查找。
func Load() *Config {
	loaded := loadDotEnvFiles()
	cfg := &Config{
		ServerAddr: get("SERVER_ADDR", ":8080"),

		MySQLDSN: get("MYSQL_DSN",
			"user:password@tcp(127.0.0.1:3306)/mybookkeeping?charset=utf8mb4&parseTime=True&loc=Local"),

		RedisAddr:     get("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: get("REDIS_PASSWORD", ""),
		RedisDB:       getInt("REDIS_DB", 0),

		MinIOEndpoint:  get("MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinIOAccessKey: get("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: get("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    get("MINIO_BUCKET", "mybookkeeping"),
		MinIOUseSSL:    getBool("MINIO_USE_SSL", false),

		AppSecret:      get("APP_SECRET", ""),
		TrustedProxies: getList("TRUSTED_PROXIES", "127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"),
		CORSOrigins:    getList("CORS_ORIGINS", ""),

		AmapKey:          get("AMAP_KEY", ""),
		AmapSecurityCode: get("AMAP_SECURITY_CODE", ""),

		SessionTTL: time.Duration(getInt("SESSION_TTL_DAYS", 7)) * 24 * time.Hour,

		AllowRegister:        getBool("ALLOW_REGISTER", true),
		ResetTokenInResponse: getBool("RESET_TOKEN_IN_RESPONSE", false),
		ResetTokenTTLHours:   getInt("RESET_TOKEN_TTL_HOURS", 2),

		DeepSeekAPIKey:      get("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:     get("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:       get("DEEPSEEK_MODEL", "deepseek-chat"),
		DeepSeekVisionModel: get("DEEPSEEK_VISION_MODEL", "deepseek-flash"),

		LoadedFrom: loaded,
	}
	if loaded != "" {
		log.Printf("config loaded from %s", loaded)
	} else {
		log.Printf("config: no .env found, using defaults / process env (copy backend/.env.example → backend/.env)")
	}
	return cfg
}

func loadDotEnvFiles() string {
	var candidates []string
	if p := os.Getenv("CONFIG_FILE"); p != "" {
		candidates = append(candidates, p)
	}
	candidates = append(candidates,
		".env",
		"config.env",
		filepath.Join("backend", ".env"),
	)
	// 从可执行文件旁再试一次（生产单文件部署）
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, ".env"),
			filepath.Join(dir, "config.env"),
		)
	}

	seen := map[string]bool{}
	for _, path := range candidates {
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		if err := applyDotEnvFile(path); err == nil {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}
	return ""
}

// applyDotEnvFile 把 KEY=VALUE 写入进程环境；已存在的系统环境变量不覆盖。
func applyDotEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		// 去掉成对引号
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue // 系统/已有环境变量优先
		}
		_ = os.Setenv(key, val)
	}
	return sc.Err()
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getList(key, def string) []string {
	var out []string
	for _, p := range strings.Split(get(key, def), ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}
