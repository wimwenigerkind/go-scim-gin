# go-scim-gin

A SCIM v2.0 implementation for Go using [Gin](https://github.com/gin-gonic/gin),
implementing [RFC 7643](https://www.rfc-editor.org/rfc/rfc7643) (Schema) and
[RFC 7644](https://www.rfc-editor.org/rfc/rfc7644) (Protocol).

> **Status:** Work in progress.

## Features

### Users

- [x] `GET    /scim/v2/Users` list ([RFC 7644 §3.4.2](https://www.rfc-editor.org/rfc/rfc7644#section-3.4.2))
- [x] `GET    /scim/v2/Users/:id` read ([RFC 7644 §3.4.1](https://www.rfc-editor.org/rfc/rfc7644#section-3.4.1))
- [x] `POST   /scim/v2/Users` create ([RFC 7644 §3.3](https://www.rfc-editor.org/rfc/rfc7644#section-3.3))
- [ ] `PUT    /scim/v2/Users/:id` replace ([RFC 7644 §3.5.1](https://www.rfc-editor.org/rfc/rfc7644#section-3.5.1))
- [ ] `PATCH  /scim/v2/Users/:id` partial update ([RFC 7644 §3.5.2](https://www.rfc-editor.org/rfc/rfc7644#section-3.5.2))
- [ ] `DELETE /scim/v2/Users/:id` delete ([RFC 7644 §3.6](https://www.rfc-editor.org/rfc/rfc7644#section-3.6))

### Protocol

- [x] Bearer-token authentication ([RFC 7644 §2](https://www.rfc-editor.org/rfc/rfc7644#section-2))
- [x] `application/scim+json` content type ([RFC 7644 §3.1](https://www.rfc-editor.org/rfc/rfc7644#section-3.1))
- [x] Server-controlled `meta` ([RFC 7643 §3.1](https://www.rfc-editor.org/rfc/rfc7643#section-3.1))
- [x] `Location` header on create ([RFC 7644 §3.3](https://www.rfc-editor.org/rfc/rfc7644#section-3.3))
- [ ] SCIM error responses ([RFC 7644 §3.12](https://www.rfc-editor.org/rfc/rfc7644#section-3.12))
- [x] `WWW-Authenticate` on 401 ([RFC 7235 §4.1](https://www.rfc-editor.org/rfc/rfc7235#section-4.1))

## Usage

```go
import scim "github.com/wimwenigerkind/go-scim-gin"

provider := &MyProvider{} // implements scim.UserProvider
handler := scim.NewHandler(provider, "your-bearer-token")

r := gin.Default()
handler.RegisterRoutes(r)
r.Run(":8080")
```

## License

Licensed under the [MIT License](LICENSE).
