<p align="center">
  <img src="logo.jpg" alt="Go-Lek Logo" width="220">
</p>

<p align="center">
  <strong>Go-Lek</strong> (Jawa: <em>golek</em> = mencari) — A modern, type-safe Go HTTP framework.
</p>

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

---

## 💡 Why Go-Lek?

Go-Lek bridges the gap between raw **`net/http`** performance and high-level framework productivity:

- **100% `net/http` compatible**: Composable with standard Go middleware.
- **Type-safe generic handlers**: `golek.H()` handles auto-binding and validation without runtime type assertion surprises.
- **Zero external dependencies**: Clean, secure, and minimal binary footprint.
- **Radix tree routing**: Fast URL routing with support for named parameters and wildcards.
- **Auto OpenAPI 3.0 generation**: Keep documentation synced with your codebase.
- **Batteries included**: CORS, rate limiting, gzip compression, panic recovery, and request tracing out of the box.

---

## ⚡ Feature Comparison

| Feature | Chi | Gin | Go-Lek |
|---|:---:|:---:|:---:|
| Radix tree router | ✅ | ✅ | ✅ |
| `net/http` compatible | ✅ | ❌ | ✅ |
| External dependencies | Zero | Many | **Zero** |
| **Type-safe generic handlers** | ❌ | ❌ | ✅ |
| **Auto request binding** | ❌ | ✅ | ✅ |
| **Auto struct validation** | ❌ | ✅ | ✅ |
| **Auto OpenAPI 3.0 generation** | ❌ | ❌ | ✅ |
| **Built-in graceful shutdown** | ❌ | ❌ | ✅ |
| **Full built-in middleware suite** | Partial | Partial | ✅ |

---

## 📦 Installation

```bash
go get github.com/kreatiflabs/go-lek
```
