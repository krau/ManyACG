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

节点角色由 `[app] role` 指定：`main` 为主节点，`worker` 为非主节点，默认 `main`，保留未配置角色的旧单机用法；环境变量为 `MANYACG_APP_ROLE`。查询和纯边缘节点必须显式配置 `role = "worker"`。能力列表不改变节点角色；角色和能力修改后需重启。

节点通过 `[app] capabilities` 声明查询和边缘执行能力，各项可独立组合；环境变量 `MANYACG_APP_CAPABILITIES` 用逗号分隔（如 `read,fetch,download`）。默认 `["read"]`；显式空列表 `[]` 不提供这些能力，但不会取消主节点职责。

| 能力 | 提供 |
| --- | --- |
| `read` | 公共只读查询与媒体接口（`/atom`、`/artwork`、`/picture`、`/artist`、`/tag`） |
| `fetch` | 作为边缘节点注册到主节点，代抓单作品详情 |
| `download` | 作为边缘节点注册到主节点，代下原始直链文件 |

公共写接口、数据库迁移、发布恢复，以及按各自开关启用的 Bot、调度、本地图像索引和搜索索引维护，均为 `role = "main"` 的专属职责，不属于能力枚举。非主节点即使开启相关开关也不会执行这些职责。

常见配置：`role = "main"` 配合 `["read"]` = 主节点；`role = "worker"` 配合 `["read"]` = 只读查询/媒体节点；`role = "worker"` 配合 `["fetch", "download"]` = 纯边缘节点。主节点也可声明 `fetch`/`download` 向其他派发节点贡献算力，但须配置自己的 Agent。主节点角色不隐含 `read`，`role = "main"` 配合 `[]` 不提供公共查询接口。启用 REST 时，任何节点都提供 `/myip`。

| 配置 | 数据库和媒体存储 | KV | Bot、调度、本地图像索引、搜索索引维护 | 自动迁移与发布恢复 |
| --- | --- | --- | --- | --- |
| `role = "main"`，任意能力列表 | 初始化 | 初始化 | 按各自开关启用 | 迁移由 `database.auto_migrate` 控制；执行发布恢复 |
| `role = "worker"`，含 `read` | 初始化 | 不初始化 | 不启用 | 不执行 |
| `role = "worker"`，不含 `read` | 不初始化 | 不初始化 | 不启用 | 不执行 |

非主节点声明 `read` 时必须启用 REST。纯边缘节点不需要 REST、数据库、KV、媒体存储或搜索服务，只需要可写缓存目录。声明了 `fetch`/`download` 必须配置 `internal.agent.gateway`；配置 `internal.gateway.addr` 的节点必须为主节点或声明 `read`。查询节点可以接入边缘下载，但不会因此获得主节点职责。不提供任何已启用接口的实例会拒绝启动。

多节点部署只能有一个活跃主节点。`app.role` 不提供自动选主或分布式锁，也不会阻止误启动第二个主节点；数据库复制、负载均衡及 CDN 需单独配置。

内部 RPC 与委托：

- 声明 `fetch`/`download` 的节点主动连接 `internal.agent.gateway` 注册能力、名称与并发数（`internal.agent.concurrency`，默认 1，派发节点最多接受 8），随后接收任务。同一公钥只允许一个活跃会话；每个边缘节点应使用独立密钥。纯边缘节点不监听端口，不需要公网 IP，位于 NAT、家用宽带或防火墙后也能接入；叠加 `read` 或主节点角色时仍可提供相应服务。
- 主节点配置 `internal.gateway.addr` 后开启内部 RPC 端口，按能力与空闲度轮询派发；主节点不需要配置任何边缘节点地址，边缘节点上下线立即生效。
- 失败回落：没有空闲边缘节点，或委托任务失败、超时、连接中断时，主节点立即在本地完成同一任务；不重试该节点，也不切换到其他节点。请求取消时会同时取消边缘任务，不再本地重试。
- 委托方仍负责 API key 鉴权、配额、缓存入库、计费、`cache_id`、压缩与存储写入；边缘节点只执行详情抓取或原始直链下载并回传结果。
- 边缘节点收到任务后永远本地执行，不会再次派发，因此任意注册关系都不会形成环路。
- 主节点只按 `internal.gateway.authorized_keys` 中人工登记的公钥允许边缘节点接入；边缘节点用 `internal.agent.private_key` 对应的私钥证明身份。可选 SSH 加密直接复用这份授权；主节点主机密钥自动生成，边缘节点通过 TOFU 自动记录首次主机身份，无需双方手工交换公钥或配置共享 token。节点名称只用于区分日志，不作为授权依据。
- 内部 RPC 使用独立端口，与公开 REST 端口无关，也不存在任何 HTTP 内部路由。默认不加密，只能用于可信网络或已有加密隧道；开启 SSH 后该端口只接受 SSH 加密连接，不自动降级。只有正常启动服务时才监听该端口，`index` 和 `imseek` 子命令不会占用它。
- 只有开启 `internal.gateway.addr` 的节点会使用边缘算力，其他节点一律本地抓取与下载；派发节点不做选主，重复开启会让边缘节点分散注册到不同节点。
- 主节点只接收任务所属连接的结果，通过已启用图源的链接匹配规则核对作品，并检查作者和必需媒体字段。非法结果按委托失败处理，回落到本地抓取；不另设图源、CDN 或媒体格式白名单。
- 原始文件下载仅允许公网 HTTP(S) 目标及各自默认端口，每次重定向和 DNS 结果都要满足限制。默认每个文件最多 1 GiB，由正整数 `internal.dispatch.max_download_bytes` 调整。主节点接收和边缘下载都执行上限，边缘还会取自身上限与任务上限的较小值；压缩响应按解压后的大小限制，上传声明长度必须与实际字节数一致。
- 新下载器只使用显式配置的 `http_client.proxy`，不使用环境变量代理。配置代理时，目标仍需通过公网检查；代理自身的解析与路由由管理员负责限制，不能允许其绕过公网出网边界访问内网。
- 公钥认证证明节点身份，不证明已授权节点返回的内容真实。基础校验不能保证媒体与源站一致，授权仍只应授予可信节点。

接入边缘节点：

1. 只需在边缘节点手工生成专用密钥，例如在部署目录执行 `ssh-keygen -t ed25519 -N '' -f ./edge_key`。私钥需无口令且只对运行账号可读；每个节点使用独立密钥，不要传输私钥。可使用现有 Ed25519、ECDSA 或至少 2048 位的 RSA 服务密钥。
2. 将 `edge_key.pub` 的完整公钥行人工加入主节点 `authorized_keys`。支持空行和注释，不支持公钥行前的 options 或证书；空文件、格式错误或读取失败都会拒绝新的认证。
3. 边缘节点配置主节点地址和自己的私钥路径后启动即可。主节点每次认证都会重新读取 `authorized_keys`，添加公钥无需重启；移除公钥只阻止后续认证，不会踢掉已有会话，立即撤权需断开该节点或重启主节点。
4. 需要加密时，在主节点 `[internal.gateway]` 和边缘节点 `[internal.agent]` 中都设置 `ssh = true`。主节点首次启用时自动生成并保存主机密钥，边缘节点首次连接自动保存主机公钥；无需运行系统 SSH 服务，也无需手工分发主节点公钥。

SSH 状态默认保存在各实例工作目录的 `data/internalrpc/`：主节点为 `ssh_host_ed25519_key`，边缘节点为 OpenSSH 格式的 `known_hosts`。可分别用 `internal.gateway.host_key`、`internal.agent.known_hosts` 修改路径，留空即使用默认值；这些路径不是必填认证配置。请持久化该目录，勿放入会定期清理的缓存；Unix 新文件使用 `0600`、新目录使用 `0700`，已有主机私钥必须为 `0600`，状态目录和文件不能允许其他用户写入。Windows 使用部署账号的文件 ACL。

TOFU 会自动接受并保存首次看到的主机公钥，后续连接和服务重启都沿用该记录；主机身份以配置的地址和端口区分。主机密钥变化、记录损坏或状态无法读写时拒绝连接，不覆盖旧记录、不回退明文。合法换钥时，先通过可信渠道确认新指纹，再用 `ssh-keygen -R '[main.example.com]:8085' -f ./data/internalrpc/known_hosts` 删除对应旧记录，下一次连接才会重新首次信任；使用自定义路径时替换 `-f` 参数。

TOFU 无法识别首次连接时的中间人，请确保首次连接环境可信；它不是事先核验的主机身份。SSH 加密也不改变任务结果的信任边界，授权节点仍需可信。

主节点配置示例（也可省略 `[app]`，使用默认主节点角色和 `read` 能力）：

```toml
[app]
role = "main"
capabilities = ["read"]

[database]
type = "pgsql"
dsn = "host=db-primary user=manyacg_writer dbname=manyacg port=5432 sslmode=verify-full"
auto_migrate = true

[rest]
enable = true
addr = ":8080"
base = "https://api.example.com/api/v1"

[internal.gateway]
addr = ":8085"
authorized_keys = "./authorized_keys"
ssh = true

[internal.dispatch]
fetch_timeout = 30
download_timeout = 300
max_download_bytes = 1073741824

[storage]
cache_dir = "./cache-writer"
```

只读查询节点配置示例：

```toml
[app]
role = "worker"
capabilities = ["read"]

[database]
type = "pgsql"
dsn = "host=db-replica user=manyacg_reader dbname=manyacg port=5432 sslmode=verify-full"
auto_migrate = false

[rest]
enable = true
addr = ":8080"
base = "https://api.example.com/api/v1"

[storage]
cache_dir = "./cache-reader"
```

边缘节点配置示例（只出网，贡献抓取与下载算力）：

```toml
[app]
role = "worker"
capabilities = ["fetch", "download"]

[internal.agent]
gateway = "main.example.com:8085"
name = "edge-tokyo-1"
concurrency = 1
private_key = "./edge_key"
ssh = true

[internal.dispatch]
max_download_bytes = 1073741824

[storage]
cache_dir = "./cache-edge"
```

DSN 中的主机、用户和 TLS 设置需按部署环境填写。非主节点的 `read` 实例应使用只读数据库账号并连接读副本；主节点连接主库。各实例使用独立缓存目录；媒体存储必须对查询节点可读，主节点本地磁盘上的文件不会自动复制。查询节点需要配置相应存储驱动，`storage.telegram` 用于媒体读取，与交互 Bot 的开关不同。纯边缘节点只需要可写缓存目录和到派发节点内部 RPC 端口的出网连通性。

部署或升级顺序：

本分支的多节点功能仍处于预发布阶段，不提供同一开发分支不同构建之间的迁移步骤。升级查询或纯边缘节点时，唯一需要的角色配置变更是显式加入 `app.role = "worker"`（或 `MANYACG_APP_ROLE=worker`）；主节点和单机默认使用 `main`，无需增加角色字段。能力列表不会改变节点角色。

1. 维护期间先停掉所有 ManyACG 节点，用 `role = "main"`、连接主库且具备数据库迁移权限的配置执行 `./manyacg migrate`。该命令只更新数据库结构与配置的数据库索引，不启动 Bot、调度或媒体存储。
2. 等待数据库结构同步到读副本，再启动一个主节点。启用 MeiliSearch 时，主节点负责共享索引的初始化与更新；首次部署可在主节点执行 `./manyacg index` 导入已有作品。
3. 启动查询节点，将普通查询和媒体请求路由到声明了 `read` 的节点；将 `/api/v1/artwork/fetch` 和 `/api/v1/bot/send_artwork_info` 的 GET、POST 请求只路由到主节点。POST `/artwork/list` 和 `/artwork/random` 仍是读接口。
4. 启动边缘节点，它们会自动注册到开启 `internal.gateway.addr` 的节点；反向代理只需暴露公开 REST 端口，内部 RPC 端口仅对受信任网络开放。

非主节点的查询实例可使用仅有目标索引 `search` 权限的 MeiliSearch 密钥，必须与主节点使用相同的 Host 和 Index。启用搜索时，查询节点启动会用一次零结果查询验证目标索引、权限和连接，不创建索引、不修改搜索设置，也不补写图片 MIME。缺失索引或搜索服务故障会阻止查询节点启动；纯边缘节点不会连接搜索服务。读副本和搜索索引可能存在同步延迟，新发布作品不保证立即在所有查询节点可见。`migrate`、`index` 和所有 `imseek` 子命令（包括 `status`）只允许主节点执行。

主节点启动时会把上次异常退出遗留的 `posting` 缓存状态恢复为可发布状态，然后才启动 Bot 和调度。必须先确认旧主节点已停止，再启动替代实例；恢复可能重新发布已经发送 Telegram 消息但尚未写入数据库的作品，因此不保证 exactly-once。非主节点不执行发布恢复。

正常停机先停止接收新请求，再等待正在处理的请求和任务结束，最长等待 `app.shutdown_timeout` 秒（默认 8）；超过上限后未完成的工作会被放弃并继续退出，不会无限等待。若调大该值，需同时调大编排器的停止超时：容器默认 10 秒后强杀进程，会截断尚未发送完的响应。

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