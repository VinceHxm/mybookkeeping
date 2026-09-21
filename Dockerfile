# 基础镜像走 DaoCloud 加速（国内飞牛直连 docker.io 常超时）
# 若已在 Docker 配置 registry-mirrors，也可改回官方名：node/golang/alpine

# ---- frontend ----
FROM docker.m.daocloud.io/library/node:22-alpine AS fe-builder
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- backend ----
FROM docker.m.daocloud.io/library/golang:1.22-alpine AS be-builder
WORKDIR /src
# 国内直连 proxy.golang.org 常超时，改用七牛/官方备选
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct \
    GOSUMDB=sum.golang.google.cn
RUN apk add --no-cache git ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/init-user ./cmd/init-user

# ---- runtime ----
FROM docker.m.daocloud.io/library/alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S -g 1000 app && adduser -S -G app -u 1000 app \
 && mkdir -p /app/public && chown -R app:app /app
WORKDIR /app
ENV TZ=Asia/Shanghai \
    CONFIG_FILE=/app/.env \
    SERVER_ADDR=:8080
COPY --from=be-builder --chown=app:app /out/server /app/server
COPY --from=be-builder --chown=app:app /out/init-user /app/init-user
COPY --from=fe-builder --chown=app:app /src/dist /app/public
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
