# swaggerui

Embedded, self-hosted [Swagger UI](https://swagger.io/tools/swagger-ui/) for Go servers.

This package provides `swaggerui.Handler`, which you can use to serve an embedded copy of
[Swagger UI](https://swagger.io/tools/swagger-ui/).

## Example usage

```go
package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"

	"github.com/allenai/go-swaggerui"
)

//go:embed openapi.json
var jsonSpec []byte

func main() {
	// Redirect /api to /api/ when serving UI.
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/", http.StatusFound)
	})

	// Serve Swagger UI using a relative path to the schema exposed below.
	http.Handle("/api/", http.StripPrefix("/api", swaggerui.Handler("v1/openapi.json")))

	// Serve the JSON-encoded schema.
	http.HandleFunc("/api/v1/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonSpec)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## Updating Swagger UI

Run `go generate ./...` and follow the printed instructions.
