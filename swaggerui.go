package swaggerui

import (
	"bytes"
	"embed"
	"io/fs"
	"io/ioutil"
	"net/http"
	"text/template"
)

//go:generate go run generate.go

//go:embed embed
var swagfs embed.FS

// Handler returns a handler that will serve a self-hosted Swagger UI.
func Handler(specURL string) http.Handler {
	static, err := fs.Sub(swagfs, "embed")
	if err != nil {
		panic(err)
	}

	indexFile, err := static.Open("index.html")
	if err != nil {
		panic(err)
	}
	indexTemplate, err := ioutil.ReadAll(indexFile)
	if err != nil {
		panic(err)
	}
	indexInfo, err := indexFile.Stat()
	if err != nil {
		panic(err)
	}

	t, err := template.New("index.html").Parse(string(indexTemplate))
	if err != nil {
		panic(err)
	}

	index := &bytes.Buffer{}
	if err := t.Execute(index, struct{ SwaggerURL string }{specURL}); err != nil {
		panic(err)
	}

	overlay := overlayFS{static, map[string]overlayFile{
		"index.html": {indexInfo, *index},
	}}

	return http.FileServer(http.FS(overlay))
}
