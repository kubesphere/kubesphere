package v1alpha1

type LanguageCode string
type LocaleString string
type Locales map[LanguageCode]LocaleString

const (
	LanguageCodeEn      = "en"
	LanguageCodeZh      = "zh"
	DefaultLanguageCode = LanguageCodeEn
)

func (l Locales) Default() string {
	if val, ok := l[DefaultLanguageCode]; ok {
		return string(val)
	}
	if zh, ok := l[LanguageCodeZh]; ok {
		return string(zh)
	}
	// pick up the first value
	for _, ls := range l {
		return string(ls)
	}
	return ""
}

func NewLocales(enVal, zhVal string) Locales {
	return Locales{
		LanguageCodeEn: LocaleString(enVal),
		LanguageCodeZh: LocaleString(zhVal),
	}
}
