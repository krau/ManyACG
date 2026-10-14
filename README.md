<div align="center">
  
# ManyACG

![ManyACG_banner](https://github.com/user-attachments/assets/1d2d7835-18c1-4a50-9cb9-c14ae69659be)

Collect, Download, Organize and Share your Favorite Anime Pictures.

</div
  
---

这里是 ManyACG 的后端代码.

ManyACG 是为收集与整理二次元插画作品而生的项目, 目前主要通过 Telegram Bot 完成数据交互.

在充当 Telegram 插画频道的爬虫与管理 Bot 的同时, ManyACG 还能使用已存入数据库的作品构建一个自己的二次元图片分享网站.

> 前端代码 -> [ManyACG/web](https://github.com/ManyACG/web)

![manyacg-web](https://github.com/user-attachments/assets/670a6092-1406-4f51-ab2b-49a6d9be286f)

## Demo

- Bot - [@KirakaBot](https://t.me/kirakabot)
- 频道 - [@MoreACG](https://t.me/MoreACG)
- 网站 - [ManyACG](https://manyacg.top)

## 特性

- **多图源支持**
  - [x] [Pixiv](https://www.pixiv.net/)
  - [x] [Twitter](https://x.com/)
  - [x] [Danbooru](https://danbooru.donmai.us/)
  - [x] [Bilibili](https://www.bilibili.com/)
  - [x] [Kemono](https://www.kemono.cr/)
  - [x] [Yandere](https://yande.re/)
  - [x] [Nhentai](https://nhentai.net/)
- **原图多存储端支持**
  - [x] 本地存储
  - [x] WebDAV
  - [x] Telegram
- 基于图像哈希的去重与以图搜图
- 带有逻辑控制的关键词搜图
- 以 Telegram 所接受的最高质量发送图片
- 支持动图和视频
- 基于 AI 的图片标签生成 -> [konatagger](https://github.com/krau/konatagger)
- 集成 [MeiliSearch](https://www.meilisearch.com/) , 支持混合搜索与相似作品检索.
- 轻量, 原生跨平台, 部署简单

## 部署

### 安装FFmpeg(可选)

ManyACG 需要使用 FFmpeg 来从动图序列合成视频, 请在自己的系统上安装, 以下是一些系统的安装示例:

Ubuntu/Debian:

```bash
sudo apt install ffmpeg -y
```

Arch Linux:

```bash
sudo pacman -S ffmpeg --noconfirm
```

[其他/任意 Linux 发行版安装 FFmepg 参考](https://krau.top/posts/linux-install-ffmpeg)

Windows:

1. 在 [gyan.dev](https://www.gyan.dev/ffmpeg/builds/) 下载 [ffmpeg-release-full.7z](https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-full.7z)
2. 解压并将 `bin` 目录添加到环境变量 `PATH`

### 从二进制文件部署 ManyACG

在 [release](https://github.com/krau/ManyACG/releases) 页面下载与自己系统和架构对应的文件, 解压.

在与解压出的二进制文件的相同目录下创建 `config.toml` 文件, 修改各项配置.

#### 最简配置

如果你只需要将 ManyACG 作为一个 Telegram 频道的自动发图与管理 Bot 使用, 使用以下简单的配置即可:

```toml
[telegram]
bot_token="token"
admins = [123456789] # 你的 Telegram 用户 ID
username = "@moreacg" # 频道用户名(如有)
chat_id = -1001234567890 # 频道 ID, 与 username 二选一

[source.pixiv]
# 建议配置 pixiv cookies, 可以提高作品的爬取成功率
[[source.pixiv.cookies]]
name = "PHPSESSID"
value = ""
[[source.pixiv.cookies]]
name = "yuid_b"
value = ""

# 如果你不需要存储原图, 以下配置也可以删除
[storage]
original_type = "telegram"
[storage.telegram]
enable = true
token = "用于发送原图的 Bot 的 Token" # 可以与 telegram.bot_token 相同
chat_id = -1001234567890 # 用于存储原图的频道 ID
```

赋予二进制文件执行权限并运行即可:

```bash
chmod +x manyacg
./manyacg
```

#### 多节点部署

节点通过 `[app] capabilities` 声明自己提供的能力，各项能力独立，可任意组合；环境变量 `MANYACG_APP_CAPABILITIES` 用逗号分隔（如 `read,fetch,download`）。未配置时默认 `["read", "write"]`；显式空列表 `[]` 不提供业务能力，也不具备写权限。修改能力后需重启。

| 能力 | 提供 |
| --- | --- |
| `read` | 公共只读查询与媒体接口（`/atom`、`/artwork`、`/picture`、`/artist`、`/tag`） |
| `write` | 公共写接口、数据库迁移、发布恢复，以及按各自开关启用的 Bot、调度、本地图像索引和搜索索引维护 |
| `fetch` | 内部接口 `POST /api/v1/internal/artwork/fetch`，代抓单作品详情 |
| `download` | 内部接口 `POST /api/v1/internal/artwork/download`，代下原始直链文件 |

常见组合：`["read", "write"]` = 完整节点；`["read"]` = 只读查询/媒体节点；`["read", "fetch", "download"]` = 查询与委托协助节点；`["download"]` = 纯下载节点。`write` 不隐含 `read`，`["write"]` 不提供公共查询接口。启用 REST 时，任何节点都提供 `/myip`；其余接口只按对应能力注册。

| 配置 | 数据库和媒体存储 | KV | Bot、调度、本地图像索引、搜索索引维护 | 自动迁移与发布恢复 |
| --- | --- | --- | --- | --- |
| 含 `write` | 初始化 | 初始化 | 按各自开关启用 | 迁移由 `database.auto_migrate` 控制；执行发布恢复 |
| 含 `read`，不含 `write` | 初始化 | 不初始化 | 不启用 | 不执行 |
| 只有 `fetch`、`download` 或 `[]` | 不初始化 | 不初始化 | 不启用 | 不执行 |

`fetch`、`download` 和不含 `write` 的 `read` 节点必须启用 REST；内部委托能力还必须配置 `source.remote.token`。不提供任何已启用接口的实例会拒绝启动。含 `write` 的节点可以只启用 Telegram、不启用 REST。

多节点部署只能有一个活跃的 `write` 节点，其余节点不声明 `write`。能力配置不会自动选主，也不会阻止误启动第二个写节点。数据库复制、负载均衡及 CDN 需要单独配置。

出站委托按操作分别配置：

- `[source.remote.fetch] endpoints` 把单作品详情抓取委托给提供 `fetch` 的节点，仅供 `write` 节点使用。调用方仍负责 API key 鉴权、配额检查、缓存入库、计费和 `cache_id` 写入；协助节点只返回作品详情。不配置时使用本地图源。
- `[source.remote.download] endpoints` 把**原始直链下载**委托给提供 `download` 的节点，可供 `read` 或 `write` 节点使用。图片原图、动图 zip 和视频等直链文件由协助节点下载后回传字节；调用方保留压缩、上传存储和入库等职责。已存在于媒体存储的文件仍由调用方直接读取。不配置时在本地下载。
- 两个节点池独立轮询。每次请求只选择一个 endpoint，不重试、不跟随节点响应的重定向，失败直接报错，不回落到本地。`source.remote.fetch.timeout` 默认 30 秒，`source.remote.download.timeout` 默认 300 秒；调用方分别用于对应请求，协助节点分别用于对应处理。
- endpoint 是协助节点的 REST API 基地址，例如 `https://fetch-1.example.com/api/v1`。双方配置同一个 `source.remote.token`。节点可同时提供内部能力和使用出站委托，但收到的内部请求始终在本地执行，不再次委托。
- 两个内部接口使用 Bearer token，独立于公共 API key 和配额。不要向公网开放，只允许受信任节点访问。下载接口会按调用方提供的 http(s) 直链出网抓取，必须保密 token。

完整节点配置示例（也可省略 `[app]` 使用默认能力）：

```toml
[app]
capabilities = ["read", "write"]

[database]
type = "pgsql"
dsn = "host=db-primary user=manyacg_writer dbname=manyacg port=5432 sslmode=verify-full"
auto_migrate = true

[rest]
enable = true
addr = ":8080"
base = "https://api.example.com/api/v1"

[source.remote]
token = "请替换为随机长串"

[source.remote.fetch]
endpoints = ["https://fetch-1.example.com/api/v1", "https://fetch-2.example.com/api/v1"]
timeout = 30

[source.remote.download]
endpoints = ["https://download-1.example.com/api/v1", "https://download-2.example.com/api/v1"]
timeout = 300

[storage]
cache_dir = "./cache-writer"
```

只读节点配置示例（承担查询、详情抓取与下载）：

```toml
[app]
capabilities = ["read", "fetch", "download"]

[database]
type = "pgsql"
dsn = "host=db-replica user=manyacg_reader dbname=manyacg port=5432 sslmode=verify-full"
auto_migrate = false

[rest]
enable = true
addr = ":8080"
base = "https://api.example.com/api/v1"

[source.remote]
token = "请替换为随机长串"

[source.remote.fetch]
timeout = 30

[source.remote.download]
timeout = 300

[storage]
cache_dir = "./cache-reader"
```

DSN 中的主机、用户和 TLS 设置需按部署环境填写。含 `read` 且不含 `write` 的节点应使用只读数据库账号并连接读副本；写节点连接主库。各实例使用独立缓存目录；媒体存储必须对查询节点可读，写节点本地磁盘上的文件不会自动复制。查询节点需要配置相应存储驱动，`storage.telegram` 用于媒体读取，与交互 Bot 的开关不同。纯 `fetch`、`download` 节点不需要数据库、KV、媒体存储或搜索服务；下载节点只需要可写缓存目录。

部署或升级顺序：

1. 维护期间先停掉所有 ManyACG 节点，用连接主库且具备 `write` 能力与数据库迁移权限的配置执行 `./manyacg migrate`。该命令只更新数据库结构与配置的数据库索引，不启动 Bot、调度或媒体存储。
2. 等待数据库结构同步到读副本，再启动一个写节点。启用 MeiliSearch 时，写节点负责共享索引的初始化与更新；首次部署可在写节点执行 `./manyacg index` 导入已有作品。
3. 启动查询和委托协助节点，将普通查询和媒体请求路由到声明了 `read` 的节点；将 `/api/v1/artwork/fetch` 和 `/api/v1/bot/send_artwork_info` 的 GET、POST 请求只路由到写节点。POST `/artwork/list` 和 `/artwork/random` 仍是读接口。

不含 `write` 的查询节点可使用仅有目标索引 `search` 权限的 MeiliSearch 密钥，必须与写节点使用相同的 Host 和 Index。启用搜索时，查询节点启动会用一次零结果查询验证目标索引、权限和连接，不创建索引、不修改搜索设置，也不补写图片 MIME。缺失索引或搜索服务故障会阻止查询节点启动；纯委托协助节点不会连接搜索服务。读副本和搜索索引可能存在同步延迟，新发布作品不保证立即在所有查询节点可见。`migrate`、`index` 和所有 `imseek` 子命令（包括 `status`）都需要 `write` 能力。

写节点启动时会把上次异常退出遗留的 `posting` 缓存状态恢复为可发布状态，然后才启动 Bot 和调度。必须先确认旧写节点已停止，再启动替代实例；恢复可能重新发布已经发送 Telegram 消息但尚未写入数据库的作品，因此不保证 exactly-once。不含 `write` 的节点不执行发布恢复。

#### 安装为服务

适用于 Linux 系统, 以 systemd 为例:

`/etc/systemd/system/manyacg.service`

```ini
[Unit]
Description=ManyACG
After=network.target

[Service]
Type=simple
WorkingDirectory=/path/to/manyacg
ExecStart=/path/to/manyacg/manyacg
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now manyacg
```

---

## 从 v0 迁移

如果你之前使用的是 v0 版本的 ManyACG, 请下载最新的 v0.x 版本 release, 并修改配置文件, 添加迁移目标数据库配置:

```toml
[migrate]
target = "sqlite" # pgsql/mysql/sqlite
dsn = "file:manyacg_migrate.db" # 连接字符串
# 示例: pgsql dsn
# dsn = "host=localhost user=postgres password=yourpassword dbname=manyacg port=5432 sslmode=disable"
```

然后运行

```bash
./manyacg db migrate
```

数据迁移完成后, 将配置文件也改为使用新的配置格式, 详情参考上方的部署章节.