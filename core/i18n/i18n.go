package i18n

import (
	"fmt"
	"reflect"
)

var (
	locales = &localeStore{store: make(map[string]*locale)}
)

type locale struct {
	id      int
	lang    string
	message map[string]string
}

type localeStore struct {
	langs []string
	store map[string]*locale
}

// Get locale from localeStore use specify lang string
func (d *localeStore) getLocale(lang string) (*locale, bool) {
	if l, ok := d.store[lang]; ok {
		return l, true
	}
	return nil, false
}

// Get target language string
func (d *localeStore) Get(lang, key string) (string, bool) {
	if locale, ok := d.getLocale(lang); ok {
		if value, ok := locale.message[key]; ok {
			return value, true
		}
	}
	return "", false
}

func (d *localeStore) Add(lc *locale) bool {
	if _, ok := d.store[lc.lang]; ok {
		return false
	}
	lc.id = len(d.langs)
	d.langs = append(d.langs, lc.lang)
	d.store[lc.lang] = lc
	return true
}

// List all locale languages
func ListLangs() []string {
	langs := make([]string, len(locales.langs))
	copy(langs, locales.langs)
	return langs
}

// Check language name if exist
func IsExist(lang string) bool {
	_, ok := locales.store[lang]
	return ok
}

// Check language name if exist
func IndexLang(lang string) int {
	if lc, ok := locales.store[lang]; ok {
		return lc.id
	}
	return -1
}

// Get language by index id
func GetLangByIndex(index int) string {
	if index < 0 || index >= len(locales.langs) {
		return ""
	}
	return locales.langs[index]
}

// A Locale describles the information of localization.
type Locale struct {
	Lang string
}

// Tr translate content to target language.
func (l Locale) Tr(key string, args ...interface{}) string {
	return Tr(l.Lang, key, args...)
}

// Index get lang index of LangStore
func (l Locale) Index() int {
	return IndexLang(l.Lang)
}

// Tr translate content to target language.
func Tr(lang, key string, args ...interface{}) string {
	value, ok := locales.Get(lang, key)
	if ok && len(args) > 0 {
		params := make([]interface{}, 0, len(args))
		for _, arg := range args {
			if arg != nil {
				val := reflect.ValueOf(arg)
				if val.Kind() == reflect.Slice {
					for i := 0; i < val.Len(); i++ {
						params = append(params, val.Index(i).Interface())
					}
				} else {
					params = append(params, arg)
				}
			}
		}
		return fmt.Sprintf(value, params...)
	}
	return fmt.Sprintf(value)
}
