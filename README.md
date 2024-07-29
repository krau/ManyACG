<p align="center">
<img src="https://socialify.git.ci/krau/manyacg/image?description=1&font=Jost&forks=1&issues=1&language=1&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxZW0iIGhlaWdodD0iMWVtIiB2aWV3Qm94PSIwIDAgMjQgMjQiPjxwYXRoIGZpbGw9IiM4ODg4ODgiIGQ9Ik0xMiAyQzYuNDggMiAyIDYuNDggMiAxMnM0LjQ4IDEwIDEwIDEwczEwLTQuNDggMTAtMTBTMTcuNTIgMiAxMiAybTQuNjQgNi44Yy0uMTUgMS41OC0uOCA1LjQyLTEuMTMgNy4xOWMtLjE0Ljc1LS40MiAxLS42OCAxLjAzYy0uNTguMDUtMS4wMi0uMzgtMS41OC0uNzVjLS44OC0uNTgtMS4zOC0uOTQtMi4yMy0xLjVjLS45OS0uNjUtLjM1LTEuMDEuMjItMS41OWMuMTUtLjE1IDIuNzEtMi40OCAyLjc2LTIuNjlhLjIuMiAwIDAgMC0uMDUtLjE4Yy0uMDYtLjA1LS4xNC0uMDMtLjIxLS4wMmMtLjA5LjAyLTEuNDkuOTUtNC4yMiAyLjc5Yy0uNC4yNy0uNzYuNDEtMS4wOC40Yy0uMzYtLjAxLTEuMDQtLjItMS41NS0uMzdjLS42My0uMi0xLjEyLS4zMS0xLjA4LS42NmMuMDItLjE4LjI3LS4zNi43NC0uNTVjMi45Mi0xLjI3IDQuODYtMi4xMSA1LjgzLTIuNTFjMi43OC0xLjE2IDMuMzUtMS4zNiAzLjczLTEuMzZjLjA4IDAgLjI3LjAyLjM5LjEyYy4xLjA4LjEzLjE5LjE0LjI3Yy0uMDEuMDYuMDEuMjQgMCAuMzgiLz48L3N2Zz4%3D&name=1&owner=1&pattern=Solid&pulls=1&stargazers=1&theme=Auto" alt="manyacg-bot" width="640" height="320" />
</p>
<div align="center">

# ManyACG

Work in progress...

</div>

ManyACG 是针对 Telegram 的 ACG 图片分享频道而设计的 Bot, 兼具爬虫和频道管理功能.

DEMO: [@kirakabot](https://kirakabot.t.me)

## ⭐ 特性

⚡ 原生跨平台, 轻易部署.

✨ 可扩展的图源支持, 从多种平台主动拉取或手动推送图片.

💾 原图文件的多协议存储端支持.

🖼️ 基于图像内容的查重与择优.

😋 编不上了...

> [!NOTE]\
> 本项目处于早期开发阶段, 有较多的破坏性更改, 请您在升级版本前自行阅读提交记录, 并妥善备份数据.

## 部署

依赖:

数据库: MongoDB 7.0+ ( mongodb.com 的免费实例足够较小规模使用 )

### 二进制部署

在 [Releases](https://github.com/krau/manyacg-bot/releases) 页面下载对应平台的二进制文件, 并解压.

编辑配置文件 `config.toml`, 或下载 [配置文件模板](https://github.com/krau/ManyACG/blob/main/config.example.toml) 并重命名为 `config.toml`.

```toml
[api] # RESTful API
enable = false # 是否启用
address = "0.0.0.0:39818" # 监听地址
auth = true # 是否启用认证
token = "123123qwq" # 认证 Token

[fetcher] # 爬虫配置
max_concurrent = 1 # 最大并发数
limit = 30 # 每次从每个图源最大拉取数量

[log] # 日志配置
level = "TRACE" # 日志级别
file_path = "logs/trace.log" # 日志文件路径
backup_num = 5 # 日志文件备份数量

[source] # 图源配置
proxy = "" # 向图源发起请求时使用的代理, 支持 http/https/socks5

[source.pixiv]
enable = true
proxy = "pixiv.re" # Pixiv 反向代理域名
urls = [] # Pixiv RSS 地址
intervel = 60 # 拉取间隔, 单位: 分钟
sleep = 1 # 拉取间隔内每次请求间隔, 单位: 秒
[[source.pixiv.cookies]] # Pixiv Cookies, 可在浏览器打开 F12 -> Application -> Cookies 中找到
name = "PHPSESSID"
value = "123_qwq"
[[source.pixiv.cookies]]
name = "yuid_b"
value = "1A2B3C"

[source.twitter]
enable = true
fx_twitter_domain = "fxtwitter.com" # FxTwitter 主域名

[source.bilibili]
enable = true

[source.danbooru]
enable = true

[source.kemono]
enable = true
session = "" # Kemono Session Key, 可选

[storage] # 原图存储策略配置
type = "webdav" # 存储类型, 目前仅支持 WebDAV
cache_dir = "./cache" # 缓存文件目录
cache_ttl = 114514 # 缓存过期时间, 单位: 秒

[storage.webdav]
url = "https://example.com/dav"
username = "dav@example.com"
password = "1231145"
path = "/" # 存储路径

[telegram]
token = "token"
api_url = "" # Telegram API 地址, 默认 https://api.telegram.org
username = "@manyacg" # 频道用户名
chat_id = -100114514 # 频道 ID. username 和 chat_id 仅需设置一个, 若频道为公开频道, 请使用 username
sleep = 1 # 向频道中发布作品间隔, 单位: 秒
group_id = -100114514 # 频道附属群组 id, 可选
admins = [114514]

[database]
uri = "mongodb://admin:mongo@127.0.0.1:27017/?authSource=admin" # 当 uri 不为空时, 优先使用 uri 直接连接数据库
host = "127.0.0.1"
port = 27017
user = "krau"
password = "1919810"
database = "manyacg"
```

## 更新

### 二进制更新

使用 ManyACG update 可自动下载最新适合当前系统的 Release.
