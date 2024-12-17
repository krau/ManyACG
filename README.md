# ManyACG

Kawaii is All You Need! ᕕ(◠ڼ◠)ᕗ

Demo:

- [Telegram @KirakaBot](https://t.me/kirakabot)
- [Telegram @MoreACG Channel](https://t.me/MoreACG)
- [ManyACG Website](https://manyacg.top)

## 特性

- **多图源支持**
  - [x] [Pixiv](https://www.pixiv.net/)
  - [x] [Twitter](https://twitter.com/)
  - [x] [Danbooru](https://danbooru.donmai.us/)
  - [x] [Bilibili](https://www.bilibili.com/)
  - [x] [Kemono](https://www.kemono.su/)
  - [x] [Yandere](https://yande.re/)
- **可选的存储原图**, 多存储端支持
  - [x] 本地存储
  - [x] WebDAV
  - [x] [Alist](https://alistgo.com/)
- 基于图像哈希的去重与以图搜图
- 带有逻辑控制的关键词搜图
- 以 Telegram 所接受的最高质量发送图片
- Web API
- 轻量, 原生跨平台, 部署简单 (大概)
  ...

## 部署

项目需要启用了副本集的 MongoDB 作为数据库, [MongoDB Cloud](https://www.mongodb.com/) 提供的免费实例足够使用, 也可以选择自行搭建.

项目使用 [FFmpeg](https://ffmpeg.org/) 进行一些图像处理, 请在自己的系统上安装, 以下是一些系统的安装示例:

Ubuntu/Debian:

```bash
sudo apt install ffmpeg -y
```

[其他/任意 Linux 发行版安装 FFmepg 参考](https://krau.top/posts/linux-install-ffmpeg)

Windows:

1. 在 [gyan.dev](https://www.gyan.dev/ffmpeg/builds/) 下载 [ffmpeg-release-full.7z](https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.7z)
2. 解压并将 `bin` 目录添加到环境变量 `PATH`

完成数据库和 FFmpeg 的安装后, 需要为准备使用的 Bot 设置一个头像, 然后在 [release](https://github.com/krau/ManyACG/releases) 页面下载与自己系统和架构对应的文件, 解压.

在与解压出的二进制文件的相同目录下创建 `config.toml` 文件, 参考以下内容修改各项配置:

```toml
# 数据库
[database]
# 数据库名
database = "manyacg"
# 连接 uri
uri = "mongodb+srv://user:pass@mongodb.com/"
# 当未配置 uri 时使用下列四项配置连接数据库
host = "mongodb.com"
port = 27017
user = "user"
password = "pass"

# 日志
[log]
# 等级
level = "DEBUG"
# 输出文件
file_path = "logs/trace.log"
# 日志备份份数
backup_num = 5

# 图源配置
[source]
# 请求代理, 支持 http(s), socks5
proxy = "http://user:pass@127.0.0.1:7890"

# pixiv
[source.pixiv]
enable = true
# 用于解决防盗链的代理
proxy = "pixiv.re"
# 自动从此连接列表中爬图, 兼容 rsshub pixiv 相关路由
urls = [
    'https://rsshub.app/pixiv/user/bookmarks/114514',
    'https://rsshub.app/pixiv/user/illustfollows',
]
# 爬取间隔, 单位: 分钟
intervel = 120
# 单个作品请求间隔, 单位: 秒
sleep = 1
# pixiv cookies 配置, 可选
# 若不配置无法请求成功部分作品
[[source.pixiv.cookies]]
name = "PHPSESSID"
value = "value"
[[source.pixiv.cookies]]
name = "yuid_b"
value = "value"

# twitter
[source.twitter]
enable = true
# FxTwitter 主域名
fx_twitter_domain = "fxtwitter.com"
# 自动从此连接列表中爬图, 兼容 rsshub twitter 相关路由
urls = []

intervel = 120
sleep = 1

# bilibili, 无需额外配置
[source.bilibili]
enable = true

# danbooru
[source.danbooru]
enable = true

# kemono
[source.kemono]
enable = true

[source.yandere]
enable = true

# 抓取配置, 建议都保持默认
[fetcher]
# 最大并发数, 影响自动爬图
max_concurrent = 1
# 单次爬取限制量
limit = 50

# 存储端配置, 可选
[storage]
# 原图存储类型
original_type = "webdav"
# 普通尺寸图片存储类型
regular_type = "alist"
# 缩略图存储类型
thumb_type = "local"
# 缓存目录
cache_dir = "./cache"
# 缓存文件过期时间, 单位: 秒
# 不建议设置过短
cache_ttl = 3600

# webdav
[storage.webdav]
enable = false
url = "https://example.com/dav"
username = "user"
password = "password"
# 存储 base 路径
path = ""

# 本地存储
[storage.local]
enable = false
# 存储 base 路径
path = "./downloads"

# Alist
[storage.alist]
enable = false
username = "krau"
password = "password"
url = "https://alist.example.com"
# alist 的 token 过期时间, 用于自动刷新 token
token_expire = 86400
# 存储 base 路径
path = "/manyacg"

# Telegram 相关配置
[telegram]
# Bot API
api_url = "https://api.telegram.org"
token="bot_token"
# bot 管理员 user id
admins = [777000]
# 启用图片发布到频道
channel = true
# 频道 username
username = "@moreacg"
# 频道 chat_id , 支持私有频道, 与 username 配置至少一项即可
chat_id = -1000721
# 可选, 评论组 id
group_id = -100114514
# 图片发布间隔, 单位: 秒
# 过小的间隔会导致 Flood Limit 而无法成功发送图片
sleep = 5

# Web API 配置
[api]
# 启用 Web API
enable = false
# 监听地址
address = "127.0.0.1:39088"
# CORS 允许来源
allowed_origins = ["https://manyacg.top"]
# API Key
key = "5LiA5Liq5aSN5p2C55qE5a+G56CB"
# JWT 相关
secret = "5LiA5Liq5b6I6Zq+55qE5py65a+G"
realm = "manyacg"
token_expire = 43200 # 单位: 秒
refresh_token_expire = 43200
# API 缓存
[api.cache]
# 启用缓存
enable = false
# 使用 redis 而不是直接使用内存
redis = false
# 全局默认过期时间, 单位: 秒
memory_ttl = 10
# 为路由配置独立过期时间, 单位: 秒
[api.cache.ttl]
"/atom" = 600
"/artwork/random" = 5
"/artwork/:id" = 600
"/artwork/list" = 10
"/artwork/count" = 5
```

更详细的配置可以参考 `config` 目录源码

赋予二进制文件执行权限并运行即可:

```bash
chmod +x manyacg
./manyacg
```
