package partial

import "context"

var (
	LocalizerKey     = "localizer"
	LocalizerDefault = &defaultLocalizer{locale: "en_US"}
)

type Localizer interface {
	Locale() string
}

func getLocalizer(ctx context.Context) Localizer {
	if loc, ok := ctx.Value(LocalizerKey).(Localizer); ok {
		return loc
	}
	return LocalizerDefault
}

type defaultLocalizer struct {
	locale string
}

func (d *defaultLocalizer) Locale() string {
	return d.locale
}
