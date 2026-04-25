# Vue Nav Page

一个轻量、可配置的站点导航页：前端用 Vue 3 写成单页静态页面，后端用 Go 写了一个零依赖静态文件服务器。

适合用来做个人导航页、内部门户、常用工具入口页，直接改 `config.json` 就能维护站点列表。

## 特性

- **配置驱动**：网站标题、分组、链接都在 `config.json` 里维护
- **Vue 3 前端**：无构建步骤，浏览器直接加载 `assets/vue.global.prod.js`
- **Go 静态服务器**：单个二进制，跨平台，资源占用很低
- **搜索过滤**：支持按名称、描述、URL 搜索
- **健康检查**：内置 `/health`
- **无后端数据库**：部署简单，拷贝目录即可运行

## 项目结构

```text
.
├── index.html                  # Vue 单页主体，包含样式和交互逻辑
├── config.json                 # 当前运行配置，可直接修改
├── config.example.json         # 示例配置
├── main.go                     # Go 静态文件服务器
├── go.mod                      # Go module
├── assets/
│   └── vue.global.prod.js      # Vue 3 生产版运行时
├── scripts/
│   ├── build.sh                # 多平台构建脚本
│   └── install-systemd.sh      # Linux systemd 安装脚本
└── deploy/
    └── vue-nav-page.service    # systemd service 示例
```

## 快速开始

### 1. 安装 Go

需要 Go 1.22 或更高版本。

Ubuntu 示例：

```bash
sudo apt update
sudo apt install -y golang-go
```

### 2. 本地运行

```bash
go run ./main.go -host 127.0.0.1 -port 21000
```

打开：

```text
http://127.0.0.1:21000/
```

健康检查：

```bash
curl http://127.0.0.1:21000/health
```

返回：

```json
{"ok":true,"app":"nav-server"}
```

### 3. 编译运行

```bash
go build -trimpath -ldflags='-s -w' -o nav-server main.go
./nav-server -host 0.0.0.0 -port 21000
```

## 配置说明

配置文件为 `config.json`，和 `nav-server` 二进制放在同一个目录。

示例：

```json
{
  "site": {
    "title": "站点导航",
    "subtitle": "常用公共网站入口，干净、快速、可配置",
    "brand": "NAV"
  },
  "groups": [
    {
      "name": "搜索与资讯",
      "items": [
        {
          "name": "Bing",
          "url": "https://www.bing.com",
          "desc": "微软搜索与 AI 搜索入口"
        }
      ]
    }
  ],
  "server": {
    "host": "0.0.0.0",
    "port": 21000
  }
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `site.title` | 页面标题 |
| `site.subtitle` | 页面副标题 |
| `site.brand` | 右上角圆形品牌标识 |
| `groups[].name` | 分组名称 |
| `groups[].items[].name` | 站点名称 |
| `groups[].items[].url` | 站点链接 |
| `groups[].items[].desc` | 站点描述 |
| `server.host` | 默认监听地址 |
| `server.port` | 默认监听端口 |

> 修改站点列表后，刷新浏览器即可生效。修改 `server.host` / `server.port` 后，需要重启服务。

## 启动参数优先级

服务端监听地址和端口支持三种配置方式，优先级从高到低：

1. 命令行参数：`-host`、`-port`
2. 环境变量：`NAV_HOST`、`NAV_PORT`
3. `config.json` 中的 `server.host`、`server.port`

示例：

```bash
NAV_HOST=0.0.0.0 NAV_PORT=21000 ./nav-server
# 或
./nav-server -host 0.0.0.0 -port 21000
```

## systemd 部署

仓库内置安装脚本：

```bash
sudo ./scripts/install-systemd.sh
```

默认会部署到：

```text
/srv/vue-nav-page
```

默认端口：

```text
21000
```

你也可以通过环境变量覆盖：

```bash
sudo APP_DIR=/srv/vue-nav-page PORT=8080 HOST=0.0.0.0 ./scripts/install-systemd.sh
```

常用管理命令：

```bash
sudo systemctl status vue-nav-page
sudo systemctl restart vue-nav-page
sudo journalctl -u vue-nav-page -f
```

## 多平台构建

```bash
./scripts/build.sh
```

会输出：

```text
dist/nav-server-linux-amd64
dist/nav-server-windows-amd64.exe
dist/nav-server-darwin-arm64
```

## 前端说明

`index.html` 是 Vue 3 单页应用，直接通过：

```html
<script src="./assets/vue.global.prod.js"></script>
```

加载 Vue，无需 Vite/Webpack 等构建工具。

页面启动后会请求：

```text
./config.json
```

然后渲染：

- 页面标题
- 副标题
- 品牌标识
- 分组列表
- 站点卡片
- 搜索过滤结果

## 安全说明

Go 服务端只提供静态文件访问，并带有路径穿越保护：

- 只允许 `GET` / `HEAD`
- 禁止访问根目录之外的文件
- 默认对静态资源设置 `Cache-Control: no-store`

## License

MIT
