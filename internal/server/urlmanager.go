package server

import (
	"github.com/obnahsgnaw/application/pkg/utils"
	"sync"
)

// UrlManager some server callback url manager, like auth、 perm
type UrlManager struct {
	sync.Mutex
	urls map[UrlType]map[string]UrlItem
}

type UrlType string

type UrlItem struct {
	url string
	ext map[string]string
}

func NewUrlManger() *UrlManager {
	return &UrlManager{urls: make(map[UrlType]map[string]UrlItem)}
}

func (m *UrlManager) Add(typ UrlType, host, address string, ext map[string]string) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.urls[typ]; !ok {
		m.urls[typ] = make(map[string]UrlItem)
	}
	m.urls[typ][host] = UrlItem{
		url: address,
		ext: ext,
	}
}

func (m *UrlManager) Remove(typ UrlType, host string) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.urls[typ]; ok {
		if _, ok = m.urls[typ][host]; ok {
			delete(m.urls[typ], host)
		}
	}
}

func (m *UrlManager) Get(typ UrlType) (urls []UrlItem) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.urls[typ]; ok {
		for _, item := range m.urls[typ] {
			urls = append(urls, item)
		}
	}
	return
}

func (m *UrlManager) GetRand(typ UrlType) *UrlItem {
	l := m.Get(typ)

	if len(l) == 0 {
		return nil
	}

	return &l[utils.RandInt(len(l))]
}

func (m *UrlManager) Had(typ UrlType) bool {
	return len(m.Get(typ)) > 0
}
