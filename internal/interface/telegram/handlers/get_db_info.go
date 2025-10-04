package handlers

// func GetStats(ctx *telegohandler.Context, message telego.Message) error {
// 	stats, err := service.GetDatabaseStats(ctx)
// 	if err != nil {
// 		common.Logger.Errorf("获取统计信息失败: %s", err)
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取统计信息失败: "+err.Error())
// 		return nil
// 	}
// 	text := fmt.Sprintf(
// 		"关于数据库可以公开的情报:\n\n总图片数: %d\n总标签数: %d\n总画师数: %d\n总作品数: %d",
// 		stats.TotalPictures, stats.TotalTags, stats.TotalArtists, stats.TotalArtworks, //stats.LastArtworkUpdate.Format("2006-01-02 15:04:05"),
// 	)
// 	utils.ReplyMessage(ctx, ctx.Bot(), message, text)
// 	return nil
// }
