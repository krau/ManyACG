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

ManyACG 只有两种节点角色：默认完整节点和 `reader`。通过 `[app] mode` 或环境变量 `MANYACG_APP_MODE` 选择角色；未配置或空值表示完整节点，唯一合法的非空值是 `reader`，非法值会阻止启动。角色在实例启动时确定，修改后需重启。

`database.auto_migrate` 默认开启；设置为 `false` 可让完整节点跳过启动迁移。reader 无论该值如何都不迁移数据库。完整节点启动恢复 `posting` 不依赖迁移开关。

| 角色 | 数据库自动迁移 | Bot、发布调度、本地图像索引 | 抓取与写入接口 | 启动时恢复 `posting` |
| --- | --- | --- | --- | --- |
| 完整节点（默认） | 由 `database.auto_migrate` 控制 | 按原有开关启用 | 提供 | 执行 |
| `reader` | 始终不执行 | 始终关闭 | 不提供 | 不执行 |

默认完整节点负责 Telegram Bot、发布和搜索索引更新；多节点部署只允许一个活跃的完整节点，多个 `reader` 提供查询和媒体读取。模式配置不会自动选主，也不会阻止误启动第二个完整节点。数据库复制、负载均衡及 CDN 需要单独配置。

完整节点配置示例（`mode` 可省略）：

```toml
[database]
type = "pgsql"
dsn = "host=db-primary user=manyacg_writer dbname=manyacg port=5432 sslmode=verify-full"
auto_migrate = true

[rest]
enable = true
addr = ":8080"
base = "https://api.example.com/api/v1"

[storage]
cache_dir = "./cache-writer"
```

读节点配置示例：

```toml
[app]
mode = "reader"

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

DSN 中的主机、用户和 TLS 设置需按部署环境填写。读节点应使用只读数据库账号并连接读副本；完整节点连接主库。各实例使用独立的缓存目录，不能共用正在使用的缓存文件；媒体存储必须对读节点可读，完整节点本地磁盘上的文件不会自动复制。读节点仍需要配置相应存储驱动，`storage.telegram` 用于媒体读取，与交互 Bot 的开关不同。

部署或升级顺序：

1. 维护期间先停掉所有 ManyACG 节点，用连接主库且具备迁移权限的配置执行 `./manyacg migrate`。该命令只更新数据库结构与配置的数据库索引，不启动 Bot、调度或媒体存储；`reader` 模式不能执行它。
2. 等待数据库结构同步到读副本，再启动一个完整节点。启用 MeiliSearch 时，完整节点负责共享索引的初始化与更新；首次部署可在完整节点执行 `./manyacg index` 导入已有作品。
3. 启动多个 `reader`，将普通查询和媒体请求路由到它们；将 `/api/v1/artwork/fetch` 和 `/api/v1/bot/send_artwork_info` 的 GET、POST 请求只路由到完整节点。POST `/artwork/list` 和 `/artwork/random` 仍是读接口，可路由到读节点。

读节点可使用仅有目标索引 `search` 权限的 MeiliSearch 密钥，必须与完整节点使用相同的 Host 和 Index。reader 启动时会用一次零结果查询验证目标索引、权限和连接；不会创建索引、修改搜索设置或补写图片 MIME。缺失索引或搜索服务故障会阻止 reader 启动。读副本和搜索索引可能存在同步延迟，新发布作品不保证立即在所有 reader 可见。`migrate`、`index` 和所有 `imseek` 子命令（包括 `status`）只能在完整节点运行；reader 不提供本地图像索引功能。

完整节点启动时会把上次异常退出遗留的 `posting` 缓存状态恢复为可发布状态，然后才启动 Bot 和调度。必须先确认旧完整节点已停止，再启动替代实例；恢复可能重新发布已经发送 Telegram 消息但尚未写入数据库的作品，因此不保证 exactly-once。reader 不执行发布恢复。

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