package errmsg

import (
	"errors"
	"fmt"
	"github.com/obnahsgnaw/application/pkg/utils"
	"strings"
)

type Language string
type LanguageMessage map[string]interface{}

var (
	En Language = "en"
	Zh Language = "zh"
)

type LocalMessage struct {
	defLanguage Language
	data        map[Language][]LanguageMessage
}

func (s *LocalMessage) SetDefaultLanguage(l Language) {
	if l != "" {

	}
}

func (s *LocalMessage) Load(lange Language, data []byte) error {
	var v LanguageMessage
	if utils.ParseJson(data, &v) {
		s.data[lange] = append(s.data[lange], v)
		return nil
	}
	return errors.New("invalid language data")
}

func (s *LocalMessage) Translate(lang Language, target string, params ...interface{}) string {
	if lang == "" {
		lang = s.defLanguage
	}
	var msgs []LanguageMessage
	var ok bool
	if msgs, ok = s.data[lang]; !ok {
		if msgs, ok = s.data[s.defLanguage]; !ok {
			if msgs, ok = s.data[En]; !ok {
				return target
			}
		}
	}
	for _, msg := range msgs {
		if strings.ContainsAny(target, ".") {
			targets := strings.Split(target, ".")
			if v, ok := msg[targets[0]]; ok {
				if vv, ok := v.(string); ok {
					return fmt.Sprintf(vv, params...)
				}
				if vv, ok := v.(map[string]interface{}); ok {
					if vvv, ok := vv[targets[1]]; ok {
						if vvvv, ok := vvv.(string); ok {
							return fmt.Sprintf(vvvv, params...)
						}
					}
				}
			}
		} else {
			if v, ok := msg[target]; ok {
				if vv, ok := v.(string); ok {
					return fmt.Sprintf(vv, params...)
				}
				if vv, ok := v.(map[string]interface{}); ok {
					if vvv, ok := vv["default"]; ok {
						if vvvv, ok := vvv.(string); ok {
							return fmt.Sprintf(vvvv, params...)
						}
					}
				}
			}
		}
	}

	return target
}

func New() *LocalMessage {
	return &LocalMessage{data: make(map[Language][]LanguageMessage), defLanguage: En}
}
