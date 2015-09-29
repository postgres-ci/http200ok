package render

import (
	"fmt"
	"github.com/flosch/pongo2"
	"net/http"
	"os"
	"path/filepath"
)

type Context map[string]interface{}

var templates = make(map[string]*pongo2.Template)

func FromDir(root string) {

	filepath.Walk(root, func(path string, fi os.FileInfo, _ error) error {

		if fi == nil || fi.IsDir() {

			return nil
		}

		if name, err := filepath.Rel(root, path); err == nil {

			templates[name] = pongo2.Must(pongo2.FromFile(path))

		} else {

			panic(err)
		}

		return nil
	})
}

func FromString(name, content string) {

	templates[name] = pongo2.Must(pongo2.FromString(content))
}

func HTML(rw http.ResponseWriter, name string, context Context) error {

	if template, found := templates[name]; found {

		return template.ExecuteWriterUnbuffered(pongo2.Context(context), rw)
	}

	return fmt.Errorf("template '%s' not found", name)
}
