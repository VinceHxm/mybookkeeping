# 我的账本 (myBookkeeping)

个人向精简记账：首页一眼看到本月收支，一键进入「记一笔」。相对原版 ezBookkeeping 砍掉过重能力，保留够用的账户 / 分类 / 标签 / 转账 / 模板 / 周期 / 附件 / 高德选点 / AI 辅助。

产品、交互与 UI 约定见 [`docs/DESIGN.md`](./docs/DESIGN.md)（改功能前请先读）。

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.22 · Gin · GORM · MySQL · Redis Session · MinIO |
| 前端 | Vue 3 · Vite · Vuetify · Pinia · ECharts · 高德 JSAPI |
| AI | DeepSeek（文字多笔识别 / 票据识图；Key 为空则关闭） |
| 部署 | 单镜像 Docker（前后端一体）；依赖可外置 |

## 功能一览

- **认证**：登录 / 注册 / 忘记密码（令牌重置）/ 改密；角色 `user` / `admin`
- **首页**：本月收支汇总 + 大圆形「记一笔」+ 最近流水；信用账户临近还款提醒
- **记一笔**：主类型「收支 | 转账 | 还款」；支出/收入；附件；标签；高德选点（单点 / 起终点路线）；模板 Chip
- **账户**：现金 / 银行卡 / 信用等；额度、账单日、还款日；敏感字段脱敏；存放说明与现场照片
- **分类 / 标签**：二级分类 + MDI 图标；新用户自动写入默认分类与常用标签
- **模板 / 周期**：记账模板；周期自动记账（后台每分钟）
- **计费规则**：刷卡折、立减、时段/节假日/乘次/金额阶梯等；模板可绑定；记一笔可预览建议价
- **明细 / 统计**：多条件筛选分页；区间统计、分类饼图、趋势
- **设置**：周起始、主题（light/dark/system）、收支颜色
- **账户管理**：默认还款账户、默认消费账户（卡片徽章区分）
- **AI**：文字多笔识别（确认后再入库）；票据识图填回单笔表单
- **管理**（admin）：用户列表、改角色、停用/启用、删除；节假日数据手动刷新

## 环境准备

1. **Go 1.22+**、**Node.js 18+**（Docker 构建用 Node 22）
2. **MySQL**、**Redis**（Session）
3. **MinIO**（附件存储；可本机 `docker compose`，也可外置）
4. （可选）**高德** Key + 安全密钥；**DeepSeek** API Key

## 配置说明

在 `backend/` 下维护 `.env`（从示例复制），启动时自动加载，一般不必设系统环境变量。

```powershell
cd backend
copy .env.example .env
# 编辑 .env：填入 MYSQL_DSN、REDIS_*、MINIO_*，以及可选的 AMAP_* / DEEPSEEK_API_KEY
```

`deploy/.env` 供 Docker Compose 使用，内容可与 `backend/.env` 保持一致。

优先级：**系统/进程环境变量** → **`.env` 文件** → **代码默认值（本地占位）**。  
自定义路径：`$env:CONFIG_FILE="D:\path\to\my.env"`。

| 配置项 | 说明 |
|--------|------|
| `MYSQL_DSN` / `REDIS_*` / `MINIO_*` | 基础设施 |
| `AMAP_KEY` / `AMAP_SECURITY_CODE` | 高德选点；空则地图能力不可用 |
| `ALLOW_REGISTER` | 是否开放注册（默认 true） |
| `RESET_TOKEN_IN_RESPONSE` | 忘记密码是否直接返回令牌，默认 false；生产勿开，由管理员在用户管理中重置 |
| `APP_SECRET` | 附件签名短链密钥；空则每次启动随机生成 |
| `TRUSTED_PROXIES` | 反向代理地址/网段，用于识别真实 IP（登录限流），默认本机+内网 |
| `CORS_ORIGINS` | 跨域白名单；同源部署留空 |
| `DEEPSEEK_API_KEY` | DeepSeek；空则关闭 AI |
| `DEEPSEEK_MODEL` | 文字识别，默认 `deepseek-chat` |
| `DEEPSEEK_VISION_MODEL` | 识图，默认 `deepseek-flash` |
| `SESSION_TTL_DAYS` | Session 有效天数，默认 7 |

> **安全**：真实密码、API Key 只放在本地 `.env`，不要提交。仓库内 `.env.example` 仅为占位示例。
>
> **部署要求**：只对外暴露 HTTPS 反向代理；MySQL / Redis / MinIO 端口仅监听内网或本机（远程维护走 SSH 隧道）。MinIO 桶保持私有，附件一律经后端签名短链访问。
> 所有接收 ID 的写接口须调用 `assertRefsOwned` 校验归属，并在 `service/tenant_isolation_test.go` 补越权用例。

## 本地启动

### 1. 后端

```powershell
cd backend
go mod tidy
# 可选：创建管理员（角色 admin）；也可前端注册后再在用户管理中提升
go run ./cmd/init-user -username=admin -password=你的密码
go run ./cmd/server
```

健康检查：<http://127.0.0.1:9180/api/health>（本地默认端口，见 `backend/.env` 的 `SERVER_ADDR`）

若库中尚无管理员，启动时会自动把用户名为 `admin` 的账号或最早用户提升为 admin。

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

## AI 说明

| 能力 | 接口 / 入口 | 说明 |
|------|-------------|------|
| 文字多笔 | `POST /api/ai/recognize-text`；页 `/transactions/recognize` | 确认列表勾选后再批量创建 |
| 票据识图 | `POST /api/ai/recognize-image`；记一笔内弹窗 | 填回当前单笔；尽量把原图挂为附件 |

未配置 `DEEPSEEK_API_KEY` 时相关入口不可用。

## 目录结构

```
myBookkeeping/
  backend/          # Go API（cmd/server、cmd/init-user）
  frontend/         # Vue 3 应用
  deploy/           # docker-compose（整站 + 可选 MinIO）
  docs/DESIGN.md    # 产品与工程规范
  Dockerfile        # 前后端多阶段构建
  README.md
```

## License

MIT
