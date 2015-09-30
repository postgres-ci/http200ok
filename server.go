package http200ok

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"sync"
)

type Handler func(c *Context)
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

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
		errorHandler: func(rw http.ResponseWriter, _ *http.Request, err error) {

			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			log.Println(err)
		},
		notFoundHandler: func(rw http.ResponseWriter, req *http.Request) {

			http.NotFound(rw, req)

		},
		methodNotAllowedHandler: func(rw http.ResponseWriter, _ *http.Request) {

			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		},
	}
}

type server struct {
	router   *httprouter.Router
	handlers []Handler

	errorHandler            ErrorHandler
	notFoundHandler         http.HandlerFunc
	methodNotAllowedHandler http.HandlerFunc
}

func (s *server) SetErrorHandler(handler ErrorHandler) {
	s.errorHandler = handler
}

func (s *server) SetNotFoundHandler(handler http.HandlerFunc) {
	s.notFoundHandler = handler
}

func (s *server) SetMethodNotAllowedHandler(handler http.HandlerFunc) {
	s.methodNotAllowedHandler = handler
}

func (s *server) Use(handler ...Handler) {

	s.handlers = append(s.handlers, handler...)
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

	s.router.PanicHandler = func(rw http.ResponseWriter, req *http.Request, err interface{}) {

		s.errorHandler(rw, req, fmt.Errorf("%v", err))
	}

	s.router.NotFound = s.notFoundHandler
	s.router.MethodNotAllowed = s.methodNotAllowedHandler

	s.router.ServeHTTP(rw, req)
}
