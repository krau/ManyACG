package telegram

import "github.com/mymmrac/telego"

var (
	CommonCommands = []telego.BotCommand{
		{
			Command:     "start",
			Description: "开始涩涩",
		},
		{
			Command:     "files",
			Description: "获取作品所有原图文件",
		},
		{
			Command:     "setu",
			Description: "来点涩图",
		},
		{
			Command:     "random",
			Description: "随机1张全年龄图片",
		},
		{
			Command:     "search",
			Description: "搜索相似图片",
		},
		{
			Command:     "info",
			Description: "获取作品图片和信息",
		},
		{
			Command:     "hash",
			Description: "计算图片信息",
		},
		{
			Command:     "stats",
			Description: "统计数据",
		},
		{
			Command:     "help",
			Description: "食用指南",
		},
		{
			Command:     "hybrid",
			Description: "基于语义与关键字混合搜索作品",
		},
		{
			Command:     "similar",
			Description: "获取与回复的图片相似的作品",
		},
	}

	AdminCommands = []telego.BotCommand{
		{
			Command:     "set_admin",
			Description: "设置管理员",
		},
		{
			Command:     "delete",
			Description: "删除作品",
		},
		{
			Command:     "r18",
			Description: "设置作品 R18",
		},
		{
			Command:     "title",
			Description: "设置作品标题",
		},
		{
			Command:     "tags",
			Description: "设置作品标签(覆盖)",
		},
		{
			Command:     "autotag",
			Description: "自动添加作品标签(AI)",
		},
		{
			Command:     "addtags",
			Description: "添加作品标签",
		},
		{
			Command:     "deltags",
			Description: "删除作品标签",
		},
		{
			Command:     "tagalias",
			Description: "为标签添加别名 <原标签名> <别名1> <别名2> ...",
		},
		{
			Command:     "post",
			Description: "发布作品 <url>",
		},
		{
			Command:     "refresh",
			Description: "删除作品缓存 <url>",
		},
		{
			Command:     "recaption",
			Description: "重新生成作品描述 <url>",
		},
	}
)
