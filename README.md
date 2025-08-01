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
- **可选的存储原图**, 多存储端支持
  - [x] 本地存储
  - [x] WebDAV
  - [x] Alist
  - [x] Telegram
- 基于图像哈希的去重与以图搜图
- 带有逻辑控制的关键词搜图
- 以 Telegram 所接受的最高质量发送图片
- Web API
- 基于 AI 的图片标签生成 -> [konatagger](https://github.com/krau/konatagger)
- 集成 [MeiliSearch](https://www.meilisearch.com/) , 支持混合搜索与相似作品检索.
- 轻量, 原生跨平台, 部署简单

## 部署

### 安装依赖组件

#### MongoDB

项目需要启用了副本集的 MongoDB 作为数据库, [MongoDB Cloud](https://www.mongodb.com/) 提供的免费实例足够使用, 也可以选择自行搭建.

你可以参考这个 repo 使用 docker compose 快速启动一个 MongoDB 副本集: [mongodb-rs-compose](https://github.com/krau/mongodb-rs-compose)

#### FFmpeg

项目使用 [FFmpeg](https://ffmpeg.org/) 进行一些图像处理, 请在自己的系统上安装, 以下是一些系统的安装示例:

Ubuntu/Debian:

```bash
sudo apt install ffmpeg -y
```

[其他/任意 Linux 发行版安装 FFmepg 参考](https://krau.top/posts/linux-install-ffmpeg)

Windows:

1. 在 [gyan.dev](https://www.gyan.dev/ffmpeg/builds/) 下载 [ffmpeg-release-full.7z](https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-full.7z)
2. 解压并将 `bin` 目录添加到环境变量 `PATH`

### 从二进制文件部署 ManyACG

完成数据库和 FFmpeg 的安装后, 需要为准备使用的 Bot 设置一个头像, 然后在 [release](https://github.com/krau/ManyACG/releases) 页面下载与自己系统和架构对应的文件, 解压.

在与解压出的二进制文件的相同目录下创建 `config.toml` 文件, 修改各项配置.

#### 最简配置

如果你只需要将 ManyACG 作为一个 Telegram 频道的自动发图与管理 Bot 使用, 使用以下简单的配置即可:

```toml
[database]
database = "manyacg"
uri = "mongodb://admin:password@localhost:27017"

[telegram]
token="token"
admins = [123456789]
channel = true
username = "@moreacg"

# 配置 pixiv cookies 可以提高大部分作品的爬取成功率
[source.pixiv]
[[source.pixiv.cookies]]
name = "PHPSESSID"
value = ""
[[source.pixiv.cookies]]
name = "yuid_b"
value = ""

# 如果你不需要存储原图, 以下配置也可以删除
[storage]
original_type = "local"

[storage.local]
enable = true
path = "./downloads"
```

#### 完整配置

如果你需要使用 ManyACG 的全部功能, 请参考 [config.all.toml](https://github.com/krau/ManyACG/blob/main/config.all.toml) 文件.

更详细的配置可以参考 `config` 目录源码

---

赋予二进制文件执行权限并运行即可:

```bash
chmod +x manyacg
./manyacg
```

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
systemctl enable manyacg
systemctl start manyacg
```

### 使用 Docker 部署 ManyACG

下载 [docker-compose.yml](https://github.com/krau/ManyACG/blob/main/docker-compose.yml) 和 [.env](https://github.com/krau/ManyACG/blob/main/.env) 文件, 修改 `.env` 文件中的配置.

```bash
docker compose up -d
```
