package t

import (
	"encoding/json"
	"errors"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/translator/langs"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	Booted bool
	bundle *i18n.Bundle

	zhLocalizer *i18n.Localizer
	enLocalizer *i18n.Localizer

	matcher = language.NewMatcher([]language.Tag{
		language.Chinese,
		language.English,
	})
)

func Init() {
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustParseMessageFileBytes(langs.ZH.Bytes(), "zh.json")
	bundle.MustParseMessageFileBytes(langs.EN.Bytes(), "en.json")

	zhLocalizer = i18n.NewLocalizer(bundle, "zh")
	enLocalizer = i18n.NewLocalizer(bundle, "en")
	Booted = true
}

func RTrans(key string, replace interface{}, locale string) string {
	if !Booted {
		return key
	}

	localize, err := GetLocalizer(locale).
		Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    key,
				Other: key,
			},
			TemplateData: replace,
		})
	if err != nil {
		mlog.Debug(err)
	}

	return localize
}

func GetLocalizer(locale string) *i18n.Localizer {
	switch locale {
	case "zh":
		return zhLocalizer
	case "en":
		return enLocalizer
	default:
		return zhLocalizer
	}
}

func Trans(key string, locale string) string {
	return RTrans(key, nil, locale)
}

func TransError(err error, locale string) error {
	return errors.New(Trans(err.Error(), locale))
}

func TransToError(key string, locale string) error {
	return errors.New(Trans(key, locale))
}

func RTransToError(key string, replace interface{}, locale string) error {
	return errors.New(RTrans(key, replace, locale))
}

