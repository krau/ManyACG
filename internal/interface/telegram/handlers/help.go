package handlers

import (
	"fmt"

	"github.com/krau/ManyACG/internal/common/version"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/service"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func Help(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	helpText := `使用方法:
/setu - 随机图片(NSFW)
/random - 随机全年龄图片
/search - 搜索相似图片
/info - 发送作品图片和信息
/hash - 计算图片信息
/stats - 获取统计数据
/files - 获取作品原图
/hybrid - 混合搜索作品
/similar - 搜索相似作品
`
	helpText += `
随机图片相关功能中支持使用以下格式的参数:
使用 '|' 分隔'或'关系, 使用 '空格' 分隔'与'关系, 示例:

/random 萝莉|白丝 猫耳|原创

表示搜索包含"萝莉"或"白丝", 且包含"猫耳"或"原创"的图片.
Inline 查询(在任意聊天框中@本bot)支持同样的参数格式.
`
	isAdmin, _ := serv.IsAdminByTgID(ctx, message.From.ID)
	if isAdmin {
		helpText += `
管理员命令:
/set_admin - 设置|删除管理员
/delete - 删除整个作品
/r18 - 设置作品R18标记
/title - 设置作品标题
/tags - 更新作品标签(覆盖原有标签)
/autotag - 自动tag作品
/addtags - 添加作品标签
/deltags - 删除作品标签
/tagalias - 为标签添加别名
/dump - 输出 json 格式作品信息
/recaption - 重新生成作品描述
`
	}
	helpText += fmt.Sprintf("\n版本: %s, 构建日期 %s, 提交 %s\nhttps://github.com/krau/ManyACG", version.Version, version.BuildTime, version.Commit[:7])
	utils.ReplyMessage(ctx, message, helpText)
	return nil
}
