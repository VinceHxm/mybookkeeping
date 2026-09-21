# 我的账本 (myBookkeeping)



个人向精简记账：首页直达「记一笔」，MySQL + Redis Session + MinIO 附件 + 高德选点。

**产品 / UI / 流程规范**（改功能前请先读）：[`docs/DESIGN.md`](./docs/DESIGN.md)

## 技术栈



- 后端：Go · Gin · GORM · MySQL · Redis · MinIO

- 前端：Vue 3 · Vite · Vuetify · ECharts · 高德 JSAPI



## 环境准备



1. **Go 1.22+**、**Node.js 18+**

2. MySQL / Redis：已配置远程（见下方默认连接）

3. **MinIO**（附件）：默认已指向远程；也可 `cd deploy && docker compose up -d minio`

4. （可选）**DeepSeek**：在 `backend/.env` 填 `DEEPSEEK_API_KEY`，启用文字识别 + 票据识图



## 配置说明

在 `backend/` 下维护 **`.env` 文件**即可（已提供一份，也可从 `.env.example` 复制），**不必写系统环境变量**。

```powershell
cd backend
# 按需编辑 .env，例如填入 DEEPSEEK_API_KEY=sk-xxx
go run ./cmd/server
# 日志会出现：config loaded from ...\.env
```

优先级：系统/进程环境变量（临时覆盖）→ `backend/.env` → 代码默认值。  
自定义文件路径：`$env:CONFIG_FILE="D:\path\to\my.env"`。

| 配置项 | 说明 |
|------|------|
| MYSQL_DSN / REDIS_* / MINIO_* | 基础设施 |
| AMAP_KEY / AMAP_SECURITY_CODE | 高德 |
| ALLOW_REGISTER | 是否开放注册（默认 true） |
| RESET_TOKEN_IN_RESPONSE | 忘记密码是否直接返回令牌（无 SMTP 时用） |
| DEEPSEEK_API_KEY | DeepSeek API Key；空则关闭 AI |
| DEEPSEEK_MODEL | 文字识别模型，默认 `deepseek-chat` |
| DEEPSEEK_VISION_MODEL | 识图模型，默认 `deepseek-flash` |



## 首次启动



### 1. 创建管理员用户（可选，也可前端注册）



```powershell

cd backend

go mod tidy

go run ./cmd/init-user -username=admin -password=你的密码

```

> `init-user` 创建/重置的用户角色为 **admin**。普通注册用户为 `user`。若库中尚无管理员，启动时会自动把 `admin` 用户名或最早用户提升为 admin。



### 2. 启动后端



```powershell

cd backend

go run ./cmd/server

```



健康检查：http://127.0.0.1:8080/api/health



### 3. 启动前端



```powershell

cd frontend

npm install

npm run dev

```



浏览器打开 http://127.0.0.1:5173 。



## 功能一览



- 登录 / 注册 / 忘记密码（令牌重置）/ 改密

- 账户 CRUD、归档、信用卡额度/账单日/对账

- 二级分类 + 图标 / 排序；标签

- 记一笔：支出 · 收入 · 转账；附件；高德选点；标签

- 记账模板、周期自动记账（后台每分钟）

- 明细：日期/账户/分类/标签/类型/关键词筛选 + 分页

- 统计：自定义区间、按账户、分类饼图、趋势

- 用户设置：默认账户、周起始、主题、收支颜色

- AI：DeepSeek 文字识别 + 票据识图（需配置 Key）

## 关于 AI

- 文字：`POST /api/ai/recognize-text`，模型 `DEEPSEEK_MODEL`（默认 deepseek-chat）
- 识图：`POST /api/ai/recognize-image`（multipart `file` + 可选 `hint`），模型 `DEEPSEEK_VISION_MODEL`（默认 deepseek-flash）
- 前端：记一笔页「文字识别 / 票据识图」；识图成功后会尽量把原图挂为附件

## 目录



```

myBookkeeping/

  backend/     Go API

  frontend/    Vue 应用

  deploy/      docker-compose（MinIO）

  README.md

```

