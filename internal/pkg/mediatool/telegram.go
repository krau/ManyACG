package mediatool

// telegramPhotoAspectRatioTarget 是压缩 Telegram photo 时实际使用的宽高比上限.
// 它小于 TelegramMaxPhotoAspectRatio, 为 Telegram 派生缩略图时的取整留出余量.
// 一般情况下这种比例的图非常少, 所以与原图不一致是可以接受的.
const telegramPhotoAspectRatioTarget = 19

type telegramPhotoPlan struct {
	scale        float64 // 等比缩放系数
	width        int     // 缩放后的图片宽高
	height       int
	canvasWidth  int // 最终画布宽高, 图片居中放置
	canvasHeight int
}

func (p telegramPhotoPlan) needsTransform() bool {
	return p.scale < 1 || p.width != p.canvasWidth || p.height != p.canvasHeight
}

// https://core.telegram.org/bots/api#sendphoto :
func planTelegramPhoto(width, height, maxEdge int) telegramPhotoPlan {
	plan := telegramPhotoPlan{
		scale:        1,
		width:        width,
		height:       height,
		canvasWidth:  width,
		canvasHeight: height,
	}
	if width <= 0 || height <= 0 {
		return plan
	}
	long, short := max(width, height), min(width, height)
	maxLongSide := TelegramMaxPhotoTotalSideLength
	padded := long > TelegramMaxPhotoAspectRatio*short
	if padded {
		maxLongSide = TelegramMaxPhotoTotalSideLength * telegramPhotoAspectRatioTarget / (telegramPhotoAspectRatioTarget + 1)
	}
	if maxEdge > 0 && long > maxEdge {
		plan.scale = float64(maxEdge) / float64(long)
	}
	if float64(long)*plan.scale > float64(maxLongSide) {
		plan.scale = float64(maxLongSide) / float64(long)
	}
	if plan.scale < 1 {
		plan.width = max(int(float64(width)*plan.scale), 1)
		plan.height = max(int(float64(height)*plan.scale), 1)
	}
	plan.canvasWidth, plan.canvasHeight = plan.width, plan.height
	if padded && plan.width >= plan.height {
		plan.canvasHeight = max(ceilDiv(plan.width, telegramPhotoAspectRatioTarget), plan.height)
	} else if padded {
		plan.canvasWidth = max(ceilDiv(plan.height, telegramPhotoAspectRatioTarget), plan.width)
	}
	return plan
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
