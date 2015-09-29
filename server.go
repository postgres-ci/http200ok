package http200ok

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sync"
)

type Handler func(c *Context)

type method uint8

const (
	delete method = iota + 1
	get
	head
	post
	put
)

func New() *server {
	return &server{
		router: httprouter.New(),
	}
}

type server struct {
	router   *httprouter.Router
	handlers []Handler
}

func (s *server) Use(handler Handler) {

	s.handlers = append(s.handlers, handler)
}

func (s *server) Delete(pattern string, handlers ...Handler) {

	s.add(delete, pattern, handlers)
}

func (s *server) Get(pattern string, handlers ...Handler) {

	s.add(get, pattern, handlers)
}

func (s *server) Head(pattern string, handlers ...Handler) {

	s.add(head, pattern, handlers)
}

func (s *server) Post(pattern string, handlers ...Handler) {

	s.add(post, pattern, handlers)
}

func (s *server) Put(pattern string, handlers ...Handler) {

	s.add(put, pattern, handlers)
}

func (s *server) WebSocket(pattern string, handlers ...Handler) {

	i := len(handlers) - 1

	s.add(get, pattern, append(handlers[:i], append([]Handler{wsMiddleware()}, handlers[i:]...)...))
}

func (s *server) add(method method, pattern string, handlers []Handler) {

	handler := func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {

		c := Context{
			mutex:    sync.Mutex{},
			Response: rw,
			Request:  req,
			params:   params,
			handlers: append(s.handlers, handlers...),
			values:   make(map[string]interface{}),
		}

		c.run()
	}

	switch method {
	case delete:
		s.router.DELETE(pattern, handler)
	case get:
		s.router.GET(pattern, handler)
	case head:
		s.router.HEAD(pattern, handler)
	case post:
		s.router.POST(pattern, handler)
	case put:
		s.router.PUT(pattern, handler)
	}
}

func (s *server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	s.router.ServeHTTP(rw, req)
}
