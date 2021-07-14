package t

import (
	"github.com/gin-gonic/gin"
)

var LocaleKey = "locale"

func GetLocale(ctx *gin.Context) string {
	if localeStr, exists := ctx.Get(LocaleKey); exists {
		return localeStr.(string)
	}

	return "zh"
}

func SetLocale(ctx *gin.Context, locale string) {
	var l string
	switch locale {
	case "zh", "en":
		l = locale
	default:
		l = "zh"
	}
	ctx.Set(LocaleKey, l)
}
