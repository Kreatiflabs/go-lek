# Type-Safe Handlers

Go-Lek uses Go generics to eliminate manual boilerplate for binding, validation, and JSON serialization.

---

## Handler Wrappers

### `golek.H[Req, Res]`
Binds request body, validates struct tags, executes handler, and writes JSON response.

```go
type CreateProductRequest struct {
    Title string  `json:"title" validate:"required,min=3"`
    Price float64 `json:"price" validate:"gte=0"`
}

type ProductResponse struct {
    ID    int     `json:"id"`
    Title string  `json:"title"`
    Price float64 `json:"price"`
}

r.Post("/products", golek.H(func(c *golek.Context, req *CreateProductRequest) (*ProductResponse, error) {
    // req is already parsed and validated!
    return &ProductResponse{
        ID:    1,
        Title: req.Title,
        Price: req.Price,
    }, nil
}))
```

---

### `golek.HNoReq[Res]`
For endpoints without request body (e.g. `GET`).

```go
r.Get("/products", golek.HNoReq(func(c *golek.Context) (*[]ProductResponse, error) {
    products := fetchProducts()
    return &products, nil
}))
```

---

### `golek.HNoRes[Req]`
For endpoints with request body but without response body (returns `204 No Content` on success).

```go
r.Delete("/products/{id}", golek.HNoRes(func(c *golek.Context, req *DeleteProductReq) error {
    return deleteProduct(req.ID)
}))
```

---

### `golek.Wrap`
For classic context-based handlers:

```go
r.Get("/ping", golek.Wrap(func(c *golek.Context) {
    c.JSON(200, map[string]string{"message": "pong"})
}))
```
