# syntax=docker/dockerfile:1
#
# 两个最终阶段，运行层相同（工具链 + start.sh + skills）：
#
#   local（默认）— 在镜像内从本仓库源码编译前端和 Go 二进制。
#     docker compose 使用这一阶段，不再拉取 autumn27/artex。
#   prebuilt — 放入构建上下文里预先交叉编译好的 dist/<arch>/artex，
#     供 .github/workflows/release.yml 使用（需 --target prebuilt）。
#
# 在线自更新已关闭。start.sh 只做进程守护，不下载、不换装。
FROM node:22-bookworm AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
ENV HUSKY=0
RUN npm ci
COPY web/ ./
ENV NEXT_EXPORT=1
RUN npm run build:static

FROM golang:1.26-bookworm AS gobuild
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/out server/webui/dist
ENV CGO_ENABLED=0
RUN go build -tags embedui -trimpath -ldflags "-s -w" -o /out/artex ./cmd/artex

FROM python:3.12-slim-bookworm AS runtime
# 常用工具：ripgrep / curl / vim，加一批 recon 常备件（按需增删）。
# Node 从 NodeSource 装 20.x：bookworm 自带的 apt nodejs 是 18，Playwright 要求 >=20。
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates ripgrep curl wget vim git jq unzip \
      dnsutils iputils-ping netcat-openbsd inetutils-telnet whois nmap \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*
# 预装 Playwright MCP 与 CLI（全局），运行时不再 npx 联网下载。
# @playwright/mcp：browser MCP 直接 `npx @playwright/mcp`（已全局装好，无需 -y/@latest）。
# @playwright/cli：提供 playwright-cli，装完顺带 --help 验证可执行。
# 再装 playwright（提供浏览器管理），装完用 --with-deps 预置 chromium 及其系统依赖，
# 这样容器内 MCP/CLI 首次启动即可用，不再联网下载浏览器。
RUN npm install -g @playwright/mcp@latest @playwright/cli@latest playwright@latest \
    && playwright-cli --help \
    && playwright install --with-deps chromium \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY start.sh /app/start.sh
COPY skills/ /app/skills/
# data/（SQLite + jwt.key）持久化点
VOLUME ["/app/data"]
EXPOSE 8787 8788

# 发布流水线：使用 CI 已交叉编译好的二进制，避免在镜像里再编一次。
FROM runtime AS prebuilt
ARG TARGETARCH
COPY dist/${TARGETARCH}/artex /app/artex
RUN chmod +x /app/artex /app/start.sh
ENTRYPOINT ["/app/start.sh"]
CMD ["-addr", ":8787", "-proxy", ":8788"]

# 本地 / docker compose：从本仓库源码构建。放在最后，作为 docker build 的默认阶段。
FROM runtime AS local
COPY --from=gobuild /out/artex /app/artex
RUN chmod +x /app/artex /app/start.sh
ENTRYPOINT ["/app/start.sh"]
CMD ["-addr", ":8787", "-proxy", ":8788"]
