package handlers

import (
	"ManyACG/service"
	"ManyACG/telegram/utils"
	"context"

	"github.com/mymmrac/telego"
)

func Help(ctx context.Context, bot *telego.Bot, message telego.Message) {
	helpText := `使用方法:

/setu - 随机图片(NSFW)
/random - 随机全年龄图片
/search - 搜索相似图片
/info - 计算图片信息
/stats - 获取统计数据
`

	if IsChannelAvailable {
		helpText += `/file - 回复一条频道的消息获取原图文件 <index>`
	}
	helpText += `
关键词参数使用 '|' 分隔或关系, 使用空格分隔与关系, 示例:

/random 萝莉|白丝 猫耳|原创

表示搜索包含"萝莉"或"白丝", 且包含"猫耳"或"原创"的图片.
Inline 查询支持同样的参数格式.
`
	isAdmin, _ := service.IsAdmin(ctx, message.From.ID)
	if isAdmin {
		helpText += `/set_admin - 设置|删除管理员
/del - 删除图片 <消息id>
/delete - 删除整个作品
/r18 - 设置作品R18标记
/tags - 更新作品标签(覆盖原有标签)
/addtags - 添加作品标签
/deltags - 删除作品标签
/fetch - 手动开始一次抓取
/process_pictures_hashsize - 处理无哈希和尺寸的图片
/process_pictures_storage - 处理图片存储(生成缩略图, 迁移用)

发送作品链接可以获取信息或发布到频道
`
	}
	helpText += "源码: https://github.com/krau/ManyACG"
	utils.ReplyMessage(bot, message, helpText)
}
