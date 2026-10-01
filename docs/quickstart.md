# Quick Start

Get your first Go-Lek server running in under a minute.

## 1. Install

```bash
go get github.com/kreatiflabs/go-lek
```

## 2. Hello World

Create `main.go`:

```go
package main

import (
    "net/http"
    golek "github.com/kreatiflabs/go-lek"
    "github.com/kreatiflabs/go-lek/middleware"
)

func main() {
    r := golek.New()

    // Attach middlewares
    r.Use(middleware.Logger)
    r.Use(middleware.Recovery)

    // Standard HTTP Handler
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from Go-Lek!"))
    })

    // Start server with graceful shutdown
    r.ListenAndServe(":3000")
}
```

## 3. Run

```bash
go run main.go
```

Test it:
```bash
curl http://localhost:3000/
```
