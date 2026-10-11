# 我的账本 (myBookkeeping)

个人向精简记账：首页一眼看到本月收支，一键进入「记一笔」。基于 ezBookkeeping 思路重写，砍掉企业账、多币种账户、复杂报表等过重能力，但把个人真正高频的场景做深——**信用卡账单与分期、团购券、理财账户、外币订阅、交通计费、AI 识别**。

产品、交互与 UI 约定见 [`docs/DESIGN.md`](./docs/DESIGN.md)（改功能前请先读）。本文只讲功能概览、启动与配置。

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.22 · Gin · GORM（AutoMigrate）· MySQL · Redis Session · MinIO |
| 前端 | Vue 3 · Vite · Vuetify · Pinia · ECharts · 高德 JSAPI |
| AI | DeepSeek（文字多笔识别 / 票据识图；Key 为空则关闭） |
| 部署 | 单镜像 Docker（前后端一体）；MySQL / Redis / MinIO 外置，或 SQLite 单容器零依赖 |

## 功能一览

### 记账

- **单一入口**：首页大圆形「记一笔」；主类型「收支 | 转账 | 还款」，收支下再切支出 / 收入
- **快捷模板**：首页最多 9 个模板，点按二次确认即入账；记一笔页模板 Chip 只填表不落库
- **分类 / 标签**：两级分类 + MDI 图标；记一笔选择器旁「+」快速新建；新用户自动写入默认分类与常用标签
- **附件**：图片存 MinIO（私有桶），经后端签名短链访问；缩略图可全屏预览、多图切换
- **地点**：高德单点定位或起终点路线（打车 / 公交地铁选站）；最近地点缓存、模板预设地点，减少高德调用
- **周期记账**：后台每分钟扫描自动生成流水；可选「金额待确认」，首页提醒、编辑保存即确认
- **外币支付**：记外币币种与金额 + 实际人民币扣款，展示支付汇率（不接汇率接口）
- **团购券**：支出可标待核销 + 过期日；一键核销、一键退款（自动生成「退款」收入并关联原券）；首页快过期提醒
- **计费规则**：刷卡折、立减、时段 / 节假日免费、乘次阶梯、金额阶梯、多优惠叠加策略；模板可绑定，记一笔可预览建议价（手填金额为准）

### 账户

- **类型**：现金 / 银行卡 / 信用 / 理财 / 其他；余额可直接手改
- **账户明细页**：按月分组展示单个账户的流入 / 流出与余额变化
- **信用卡账单**：按账单日只读计算「本期已还 / 本期待还 / 整体待还」，出账后的退款计入本期已还；还款日前一天或当天首页提醒
- **分期**：刷卡可填分期期数、利息与首期出账（随刷卡 / 下期）；建账可补录存量分期；账户页展开「分期中 N 笔」明细
- **还款建议**：从提醒或「去还款」进入自动预填本期待还和默认还款账户；可选按期数拆分、提前还款
- **理财账户**：买入 / 赎回走转账；「校准市值」生成不计收支的 `revalue` 流水；展示净投入 / 累计收益 / 收益率
- **默认账户**：默认还款账户、默认消费账户，记一笔切类型时自动带出
- **账户档案**：开户行、卡号、户名、存放说明与现场照片；卡号 / 户名默认脱敏，查看明文需输入登录密码

### 明细与统计

- **首页**：本月收支、还款 / 待确认 / 团购券提醒、本月支出构成、近 7 天支出、最近 5 条流水；一次请求 `GET /home/overview` 取齐
- **明细**：多条件筛选分页，支持按待确认、团购券状态过滤
- **统计**：时间预设（本月 / 上月 / 近 3 月 / 今年 / 去年 / 自定义）+ 账户筛选；概览卡片 + 五个分页
  - 构成：分类占比饼图与排行
  - 趋势：收支柱状趋势
  - 日历：按月日历看每日支出 / 收入，附日均与最高的一天
  - 地点：基于高德的消费地点分布
  - 资产：净资产 / 总资产 / 总负债走势（按日 / 月 / 年），按当前余额倒推各期末余额

### AI 辅助

| 能力 | 接口 / 入口 | 说明 |
|------|-------------|------|
| 文字多笔 | `POST /api/ai/recognize-text-batch`；页 `/transactions/recognize` | 返回可编辑确认列表，勾选后再批量创建 |
| 票据识图 | `POST /api/ai/recognize-image`；记一笔内弹窗 | 填回当前单笔表单，可挂原图为附件 |
| 单笔文字 | `POST /api/ai/recognize-text` | 兼容接口 |

Prompt 会带上用户当前的账户 / 分类 / 标签名单。未配置 `DEEPSEEK_API_KEY` 时相关入口不可用。

### 账号与管理

- **认证**：登录 / 注册 / 忘记密码 / 改密；登录失败按真实 IP 限流
- **设置**：周起始、主题（light / dark / system）、收支颜色
- **管理员**（`admin`）：用户列表、改角色、停用 / 启用、重置密码、删除（级联清数据）；节假日数据手动刷新

## 环境准备

1. **Go 1.22+**、**Node.js 18+**（Docker 构建用 Node 22）
2. **MySQL**、**Redis**（Session）
3. **MinIO**（附件存储；可本机 `docker compose`，也可外置）

> 不想装 MySQL / Redis / MinIO：设 `DB_DRIVER=sqlite` 即可，数据库、会话、附件全部落在 `DATA_DIR`（默认 `backend/data/`），见下文「单容器一体化部署」。
4. （可选）**高德** Key + 安全密钥；**DeepSeek** API Key

## 配置说明

在 `backend/` 下维护 `.env`（从示例复制），启动时自动加载，一般不必设系统环境变量。

```powershell
cd backend
copy .env.example .env
# 编辑 .env：填入 MYSQL_DSN、REDIS_*、MINIO_*，以及可选的 AMAP_* / DEEPSEEK_API_KEY
```

`deploy/.env` 供 Docker Compose 使用，内容可与 `backend/.env` 保持一致。

优先级：**系统 / 进程环境变量** → **`.env` 文件** → **代码默认值（本地占位）**。  
自定义路径：`$env:CONFIG_FILE="D:\path\to\my.env"`。

| 配置项 | 说明 |
|--------|------|
| `SERVER_ADDR` | 监听地址；本地默认 `:9180`，Docker 内为 `:8080` |
| `DB_DRIVER` | `mysql`（默认）/ `sqlite`；sqlite 时下两项默认切到 `db` / `local` |
| `SESSION_STORE` | `redis` / `db`（会话、登录限流、重置令牌存业务库） |
| `STORAGE_DRIVER` | `minio` / `local`（存 `DATA_DIR/attachments`）/ `none` |
| `DATA_DIR` / `SQLITE_PATH` | 本地数据目录（默认 `data`）/ SQLite 文件（默认 `DATA_DIR/mybookkeeping.db`） |
| `MYSQL_DSN` | MySQL 连接串（连接池 20 常驻，建连 5s / 读写 30s 超时） |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | Session 存储（`SESSION_STORE=redis` 时） |
| `MINIO_ENDPOINT` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` / `MINIO_BUCKET` / `MINIO_USE_SSL` | 附件存储（`STORAGE_DRIVER=minio` 时）；桶保持私有 |
| `APP_SECRET` | 附件签名短链密钥；空则每次启动随机生成（重启后旧链接失效） |
| `TRUSTED_PROXIES` | 反向代理地址 / 网段，用于识别真实 IP（登录限流），默认本机 + 内网 |
| `CORS_ORIGINS` | 跨域白名单；同源部署留空 |
| `AMAP_KEY` / `AMAP_SECURITY_CODE` | 高德地图；空则选点与地点统计不可用 |
| `SESSION_TTL_DAYS` | Session 有效天数，默认 7 |
| `ALLOW_REGISTER` | 是否开放注册，默认 true |
| `RESET_TOKEN_IN_RESPONSE` | 忘记密码是否直接返回令牌，默认 false；生产勿开，由管理员重置密码 |
| `RESET_TOKEN_TTL_HOURS` | 重置令牌有效小时数，默认 2 |
| `DEEPSEEK_API_KEY` | DeepSeek；空则关闭 AI |
| `DEEPSEEK_BASE_URL` | 默认 `https://api.deepseek.com` |
| `DEEPSEEK_MODEL` | 文字识别，默认 `deepseek-chat` |
| `DEEPSEEK_VISION_MODEL` | 识图，默认 `deepseek-flash` |

> **安全**：真实密码、API Key 只放在本地 `.env`，不要提交。仓库内 `.env.example` 仅为占位示例。
>
> **部署要求**：只对外暴露 HTTPS 反向代理；MySQL / Redis / MinIO 端口仅监听内网或本机（远程维护走 SSH 隧道）。MinIO 桶保持私有，附件一律经后端签名短链访问。

## 本地启动

### 1. 后端

```powershell
cd backend
go mod tidy
# 可选：创建管理员（角色 admin）；也可前端注册后再在用户管理中提升
go run ./cmd/init-user -username=admin -password=你的密码
go run ./cmd/server
```

健康检查：<http://127.0.0.1:9180/api/health>，返回 MinIO / AI / 注册开关状态。

- 表结构由 GORM AutoMigrate 自动维护，**改了模型字段后需重启后端**。
- 若库中尚无管理员，启动时会自动把用户名为 `admin` 的账号或最早用户提升为 admin。

### 2. 前端

```powershell
cd frontend
npm install
npm run dev
```

浏览器打开 <http://127.0.0.1:5173>。开发时 Vite 会把 `/api` 代理到后端。

### 3.（可选）本机 MinIO

```powershell
cd deploy
docker compose -f docker-compose.minio.yml up -d
```

按 `.env` 中的 endpoint / 密钥对接即可。

## 测试与构建

```powershell
# 后端单元测试（账单 / 分期、理财、团购券、统计、首页、多租户越权等）
cd backend
go test ./...

# 前端类型检查 + 生产构建
cd frontend
npm run build
```

> 所有接收 ID 的写接口须调用 `assertRefsOwned` 校验归属，并在 `service/tenant_isolation_test.go` 补越权用例。

## Docker 部署（前后端一体）

单容器跑 API + 静态前端；MySQL / Redis / MinIO 走 `.env` 外置地址。

```powershell
cd deploy
copy .env.example .env
# 编辑 .env 指向你的 MySQL / Redis / MinIO
docker compose up -d --build
```

默认映射：`http://主机:9002` → 容器内 `:8080`。

首次可用容器内工具创建管理员：

```powershell
docker exec -it mybookkeeping /app/init-user -username=admin -password=你的密码
```

### 单容器一体化部署（SQLite，无外部依赖）

不需要 MySQL / Redis / MinIO，所有数据在 `deploy/data/`（`mybookkeeping.db` + `attachments/`），备份复制该目录即可（建议先停容器）。

```bash
cd deploy
cp .env.example .env          # 只需按需填 APP_SECRET、AMAP_*、DEEPSEEK_*
mkdir -p data && sudo chown 1000:1000 data
docker compose -f docker-compose.lite.yml up -d --build
```

> MySQL 模式与 SQLite 模式是两套独立数据，切换不会自动迁移。SQLite 适合个人 / 家庭少量用户；多人高并发写入仍建议 MySQL。

## 目录结构

```
myBookkeeping/
  backend/
    cmd/server/        # API 入口（路由注册、后台任务）
    cmd/init-user/     # 创建 / 重置管理员
    internal/
      config/          # .env 加载
      database/        # MySQL / SQLite / Redis 连接与迁移
      kv/              # 会话等短期数据：Redis 或业务库
      handler/         # HTTP 层，按资源分文件（account / transaction / stats …）
      service/         # 业务逻辑（账单、分期、理财、团购券、统计、计费 …）
      model/           # GORM 模型
      middleware/      # 登录态与管理员校验
      storage/         # 附件存储：MinIO / 本地目录
  frontend/
    src/views/         # 页面（只做组装）
    src/components/    # 功能块组件（含 stats/ 统计分页）
    src/composables/   # 跨组件业务逻辑
    src/stores/        # Pinia 状态与字典缓存
  deploy/              # docker-compose（整站 + 可选 MinIO）
  docs/DESIGN.md       # 产品与工程规范
  Dockerfile           # 前后端多阶段构建
```

## License

MIT
