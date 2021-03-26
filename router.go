package socker

import (
	"sync"
)

type Handler func(msg Data) (out interface{}, handler Handler)

type Router struct {
	sync.RWMutex
	Handlers map[string]Handler
}

func (h *Router) Register(api string, handler Handler) {
	h.Lock()
	defer h.Unlock()
	h.Handlers[api] = handler
}

func (h *Router) Get(api string) Handler {
	h.RLock()
	defer h.RUnlock()
	return h.Handlers[api]
}

func (h *Router) Remove(api string) {
	h.Lock()
	defer h.Unlock()
	delete(h.Handlers, api)
}

func (h *Router) RemoveAll() {
	h.Handlers = map[string]Handler{}
}
