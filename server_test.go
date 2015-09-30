package http200ok

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestServerUse(t *testing.T) {

	var used bool

	app := New()
	app.Use(func(_ *Context) {
		used = true
	})
	app.Get("/", func(c *Context) {

		used = true
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	res, err := client.Get(ts.URL)

	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, used)
	}
}

func TestServerGet(t *testing.T) {

	app := New()
	app.Get("/", func(c *Context) {

		fmt.Fprint(c.Response, "GetTest")
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	res, err := client.Get(ts.URL)

	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)

		if assert.NoError(t, err) {

			assert.Contains(t, string(body), "GetTest")
		}
	}
}

func TestServerPost(t *testing.T) {

	app := New()

	var isPost bool

	app.Post("/post/", func(c *Context) {

		isPost = c.IsPost()

		fmt.Fprint(c.Response, "PostTest")
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	res, err := client.Post(ts.URL+"/post/", "application/x-www-form-urlencoded", bytes.NewReader([]byte{}))

	if assert.NoError(t, err) {

		if assert.True(t, isPost) {

			assert.Equal(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)

			if assert.NoError(t, err) {

				assert.Contains(t, string(body), "PostTest")
			}
		}
	}
}

func TestServerPut(t *testing.T) {

	app := New()

	app.Put("/put/", func(c *Context) {

		fmt.Fprint(c.Response, "PutTest")
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if req, err := http.NewRequest("PUT", ts.URL+"/put/", bytes.NewReader([]byte{})); assert.NoError(t, err) {

		res, err := client.Do(req)

		if assert.NoError(t, err) {

			assert.Equal(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)

			if assert.NoError(t, err) {

				assert.Contains(t, string(body), "PutTest")
			}
		}
	}
}

func TestServerDelete(t *testing.T) {

	app := New()

	app.Delete("/delete/", func(c *Context) {

		fmt.Fprint(c.Response, "DeleteTest")
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if req, err := http.NewRequest("DELETE", ts.URL+"/delete/", bytes.NewReader([]byte{})); assert.NoError(t, err) {

		res, err := client.Do(req)

		if assert.NoError(t, err) {

			assert.Equal(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)

			if assert.NoError(t, err) {

				assert.Contains(t, string(body), "DeleteTest")
			}
		}
	}
}

func TestServerHead(t *testing.T) {

	app := New()

	app.Head("/head/", func(c *Context) {

		fmt.Fprint(c.Response, "HeadTest")
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if req, err := http.NewRequest("HEAD", ts.URL+"/head/", bytes.NewReader([]byte{})); assert.NoError(t, err) {

		res, err := client.Do(req)

		if assert.NoError(t, err) {

			assert.Equal(t, http.StatusOK, res.StatusCode)

			body, err := ioutil.ReadAll(res.Body)

			if assert.NoError(t, err) {

				assert.Len(t, body, 0)
			}
		}
	}
}

func TestServerWebSocket(t *testing.T) {

	type T struct {
		Message string
	}

	var ws *websocket.Conn

	app := New()

	app.WebSocket("/ws/", func(c *Context) {

		ws = c.WebSocket.Conn()

		c.WebSocket.SendJSON(T{Message: "TestWebSocket"})
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	res, err := client.Get(ts.URL + "/ws/")

	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}

	if url, err := url.Parse(ts.URL); assert.NoError(t, err) {

		ws, err := websocket.Dial(fmt.Sprintf("ws://%s/ws/", url.Host), "", ts.URL)

		if assert.NoError(t, err) {

			if assert.NotNil(t, ws) {

				if assert.IsType(t, (*websocket.Conn)(nil), ws) {

					var message T

					if err := websocket.JSON.Receive(ws, &message); assert.NoError(t, err) {

						assert.Equal(t, "TestWebSocket", message.Message)
					}
				}
			}
		}
	}
}

func TestServerParams(t *testing.T) {

	var (
		expected = []string{"A", "B", "C", "D"}
		params   []string
	)

	app := New()

	app.Get("/params/:Param/", func(c *Context) {

		params = append(params, c.RequestParam("Param"))
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	for _, param := range expected {

		if res, err := client.Get(ts.URL + "/params/" + param + "/"); assert.NoError(t, err) {

			assert.Equal(t, http.StatusOK, res.StatusCode)
		}
	}

	if assert.Len(t, params, len(expected)) {

		for k, v := range expected {

			assert.Equal(t, v, params[k])
		}
	}
}

func TestServerNotFound(t *testing.T) {

	app := New()
	app.Get("/", func(c *Context) {})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if res, err := client.Get(ts.URL + "/404/"); assert.NoError(t, err) {

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	}

}

func TestServerMethodNotAllowed(t *testing.T) {

	app := New()
	app.Get("/", func(c *Context) {})
	app.Post("/post/", func(c *Context) {})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if res, err := client.Get(ts.URL + "/post/"); assert.NoError(t, err) {

		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	}

	res, err := client.Post(ts.URL, "application/x-www-form-urlencoded", bytes.NewReader([]byte{}))

	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	}
}
