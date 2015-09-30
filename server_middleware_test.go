package http200ok

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerMiddlewareUse(t *testing.T) {

	var (
		urls  = []string{"/", "/a/", "/b/", "/c/"}
		used1 = make(map[string]bool)
		used2 = make(map[string]bool)
	)

	app := New()
	app.Use(func(c *Context) {

		used1[c.Request.RequestURI] = true
	})

	app.Use(func(c *Context) {

		used2[c.Request.RequestURI] = true
	})

	for _, url := range urls {

		app.Get(url, func(_ *Context) {})
	}

	ts := httptest.NewServer(app)

	client := &http.Client{}

	for _, url := range urls {

		if res, err := client.Get(ts.URL + url); assert.NoError(t, err) {

			assert.Equal(t, http.StatusOK, res.StatusCode)
		}
	}

	if assert.Len(t, used1, len(urls)) && assert.Len(t, used2, len(urls)) {

		for _, url := range urls {

			_, ok := used1[url]

			if assert.True(t, ok) {

				_, ok := used2[url]

				assert.True(t, ok)
			}
		}
	}
}

func TestServerMiddlewareOrder(t *testing.T) {

	var (
		expected = map[int]string{
			0: "A",
			1: "B",
			2: "C",
			3: "D",
		}

		order []string
	)

	app := New()
	app.Use(
		func(_ *Context) {

			order = append(order, expected[0])
		},
		func(_ *Context) {

			order = append(order, expected[1])
		},
	)

	app.Get("/",
		func(_ *Context) {
			order = append(order, expected[2])
		},

		func(_ *Context) {
			order = append(order, expected[3])
		},
	)

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if res, err := client.Get(ts.URL); assert.NoError(t, err) {

		if assert.Equal(t, http.StatusOK, res.StatusCode) && assert.Len(t, order, len(expected)) {

			for k, v := range expected {

				assert.Equal(t, v, order[k])
			}
		}
	}
}

func TestServerMiddlewareContext(t *testing.T) {

	type User struct {
		UserID int
	}

	var (
		ok   bool
		noOk bool
		user *User
	)

	app := New()
	app.Use(func(c *Context) {

		_, noOk = c.Get("user").(*User)
	})
	app.Use(func(c *Context) {

		c.Set("user", &User{UserID: 42})
	})

	app.Get("/", func(c *Context) {

		user, ok = c.Get("user").(*User)
	})

	ts := httptest.NewServer(app)

	client := &http.Client{}

	if res, err := client.Get(ts.URL); assert.NoError(t, err) {

		if assert.Equal(t, http.StatusOK, res.StatusCode) {

			if assert.True(t, ok) && assert.False(t, noOk) && assert.NotNil(t, user) {

				assert.IsType(t, (*User)(nil), user)
			}
		}
	}
}
