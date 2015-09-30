// +build ignore

package main

import (
	"fmt"
	"github.com/ops-console/http200ok"
	"github.com/ops-console/http200ok/render"
	"log"
	"net/http"
	"time"
)

type Message struct {
	Message string
}

type User struct {
	UserID int
}

func (u *User) IsAuth() bool {
	return u.UserID != 0
}

var tpl = `
<!DOCTYPE html>
<html>
	<head>
		<title></title>
		<script src="http://code.jquery.com/jquery-1.11.0.min.js"></script>
	</head>
	<body>
		<script>
			ws = new WebSocket('ws://127.0.0.1:9009/ws/');
			ws.onopen = function () {
				ws.onmessage = function (evt) {
					var data = JSON.parse(evt.data);
					$("#Message").html(data.Message);
				}
			};
		</script>
		<div id="Message"></div>
	</body>
</html>
`

func main() {

	render.FromString("index.html", tpl)

	app := http200ok.New()

	app.Use(func(c *http200ok.Context) {

		c.Set("CurrentUser", &User{UserID: 1})

		fmt.Println("I'm a global middleware", c.Request.RequestURI)
	})

	app.Get("/", func(c *http200ok.Context) {

		render.HTML(c.Response, "index.html", nil)
	})

	app.WebSocket("/ws/", func(c *http200ok.Context) {
		/*
			user, ok := c.Get("CurrentUser").(*User)

			if !ok || !user.IsAuth() {

				http.NotFound(c.Response, c.Request)

				c.Stop()

				return
			}
		*/
		fmt.Println("I'm a local middleware")

	}, func(c *http200ok.Context) {

		user, ok := c.Get("CurrentUser").(*User)

		if ok && user.IsAuth() {

			fmt.Println("OK: i'm a disco dancer")
		}

		var i = 0

		for {

			if err := c.WebSocket.SendJSON(Message{Message: fmt.Sprintf("Hello %d", i)}); err != nil {

				fmt.Println(err)

				return
			}

			i++

			<-time.After(time.Second)
		}

	})

	app.Get("/panic/", func(c *http200ok.Context) {

		panic("AAA")
	})

	app.SetErrorHandler(func(rw http.ResponseWriter, req *http.Request, err error) {

		http.Error(rw, fmt.Sprintf("Panic: %s", err.Error()), http.StatusInternalServerError)
	})

	app.SetNotFoundHandler(func(rw http.ResponseWriter, req *http.Request) {

		http.Error(rw, fmt.Sprintf("%s not found", req.RequestURI), http.StatusNotFound)
	})

	log.Fatal(http.ListenAndServe(":9009", app))
}
