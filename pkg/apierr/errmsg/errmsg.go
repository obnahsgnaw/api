package errmsg

import (
	"errors"
	"fmt"
	"github.com/obnahsgnaw/application/pkg/utils"
	"strconv"
	"strings"
)

type Language string
type LanguageMessage map[string]interface{}

var (
	En Language = "en"
	Zh Language = "zh"
)

type LocalMessage struct {
	projectId   string
	defLanguage Language
	data        map[string]map[Language][]LanguageMessage
}

func (s *LocalMessage) SetDefaultLanguage(l Language) {
	if l != "" {
		s.defLanguage = l
	}
}

func (s *LocalMessage) SetProjectId(projectId int) {
	s.projectId = strconv.Itoa(projectId)
}

func (s *LocalMessage) Load(lange Language, data []byte) error {
	var v LanguageMessage
	if utils.ParseJson(data, &v) {
		if _, ok := s.data[s.projectId]; !ok {
			s.data[s.projectId] = make(map[Language][]LanguageMessage)
		}
		s.data[s.projectId][lange] = append(s.data[s.projectId][lange], v)
		return nil
	}
	return errors.New("invalid language data")
}

func (s *LocalMessage) Translate(lang Language, target string, params ...interface{}) string {
	if lang == "" {
		lang = s.defLanguage
	}
	projectId := "0"
	if strings.ContainsAny(target, "@") {
		targets := strings.Split(target, "@")
		projectId = targets[0]
		target = targets[1]
		if projectId == "" {
			projectId = "0"
		}
	}
	var projectMsgs map[Language][]LanguageMessage
	var msgs []LanguageMessage
	var ok bool
	if projectMsgs, ok = s.data[projectId]; !ok {
		projectMsgs, _ = s.data["0"]
	}
	if projectMsgs != nil && len(projectMsgs) > 0 {
		if msgs, ok = projectMsgs[lang]; !ok {
			if msgs, ok = projectMsgs[s.defLanguage]; !ok {
				if msgs, ok = projectMsgs[En]; !ok {
					return target
				}
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

func (s *LocalMessage) Merge(l *LocalMessage) {
	if _, ok := s.data[s.projectId]; !ok {
		s.data[l.projectId] = make(map[Language][]LanguageMessage)
	}
	for lang, langMsgs := range l.data[l.projectId] {
		if _, ok := s.data[l.projectId][lang]; !ok {
			s.data[l.projectId][lang] = langMsgs
		} else {
			s.data[l.projectId][lang] = append(s.data[l.projectId][lang], langMsgs...)
		}
	}
}

func New() *LocalMessage {
	return &LocalMessage{projectId: "0", defLanguage: En, data: make(map[string]map[Language][]LanguageMessage)}
}
