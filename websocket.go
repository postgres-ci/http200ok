package http200ok

import (
	"golang.org/x/net/websocket"
)

type webSocket struct {
	ws *websocket.Conn
}

func (w *webSocket) Conn() *websocket.Conn {

	return w.ws
}

func (w *webSocket) SendJSON(v interface{}) error {

	return websocket.JSON.Send(w.ws, v)
}

func wsMiddleware() Handler {

	return func(c *Context) {

		wss := websocket.Server{

			Handler: func(ws *websocket.Conn) {

				c.WebSocket = webSocket{ws: ws}

				c.Next()
			},
		}

		wss.ServeHTTP(c.Response, c.Request)

		if c.WebSocket.ws == nil {

			c.Stop()
		}
	}
}
