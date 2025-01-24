package handlers

import (
	"context"
	"fmt"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"

	"github.com/mymmrac/telego"
)

func Help(ctx context.Context, bot *telego.Bot, message telego.Message) {
	helpText := `使用方法:
/setu - 随机图片(NSFW)
/random - 随机全年龄图片
/search - 搜索相似图片
/info - 发送作品图片和信息
/hash - 计算图片信息
/stats - 获取统计数据
/files - 获取作品原图
/query - 混合搜索作品
`
	helpText += `
随机图片相关功能中支持使用以下格式的参数:

使用 '|' 分隔'或'关系, 使用 '空格' 分隔'与'关系, 示例:

/random 萝莉|白丝 猫耳|原创

表示搜索包含"萝莉"或"白丝", 且包含"猫耳"或"原创"的图片.

Inline 查询(在任意聊天框中@本bot)支持同样的参数格式.
`
	isAdmin, _ := service.IsAdmin(ctx, message.From.ID)
	if isAdmin {
		helpText += `
管理员命令:
/set_admin - 设置|删除管理员
/delete - 删除整个作品
/r18 - 设置作品R18标记
/title - 设置作品标题
/tags - 更新作品标签(覆盖原有标签)
/addtags - 添加作品标签
/deltags - 删除作品标签
/tagalias - 为标签添加别名
/dump - 输出 json 格式作品信息
/recaption - 重新生成作品描述

process_pictures_hashsize - 处理无哈希和尺寸的图片
process_pictures_storage - 处理图片存储(生成缩略图, 迁移用)
fix_twitter_artists - 修复Twitter作者信息(更新所有推特作品的作者信息)

发送作品链接可以获取信息或发布到频道
`
	}
	helpText += fmt.Sprintf("\n版本: %s, 构建日期 %s, 提交 %s\nhttps://github.com/krau/ManyACG", common.Version, common.BuildTime, common.Commit[:7])
	utils.ReplyMessage(bot, message, helpText)
}
