---
name: gin
description: Expert guidance for building HTTP web applications and APIs with the Gin web framework in Go. Use when working with Gin routers, middleware, route handlers, request binding, response rendering, template engines, authentication, or any HTTP server development using github.com/gin-gonic/gin.
license: MIT
---

# Gin Web Framework

Expert guidance for building modern HTTP web applications and REST APIs using the Gin web framework in Go.

## Quick Start

### Basic Server

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default() // Creates router with Logger and Recovery middleware
    
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "pong"})
    })
    
    r.Run(":8080") // Listen and serve on 0.0.0.0:8080
}
```

### Router Setup

**Default router** (includes Logger + Recovery middleware):
```go
r := gin.Default()
```

**New router** (no middleware):
```go
r := gin.New()
```

**Custom middleware**:
```go
r := gin.New()
r.Use(gin.Logger())
r.Use(gin.Recovery())
r.Use(customMiddleware())
```

## Routing

### HTTP Methods

```go
r.GET("/get", getHandler)
r.POST("/post", postHandler)
r.PUT("/put", putHandler)
r.DELETE("/delete", deleteHandler)
r.PATCH("/patch", patchHandler)
r.HEAD("/head", headHandler)
r.OPTIONS("/options", optionsHandler)
```

### Path Parameters

```go
// Match /user/john
r.GET("/user/:name", func(c *gin.Context) {
    name := c.Param("name")
    c.String(http.StatusOK, "Hello %s", name)
})

// Match /user/john/books
r.GET("/user/:name/:action", func(c *gin.Context) {
    name := c.Param("name")
    action := c.Param("action")
    c.String(http.StatusOK, "%s is %s", name, action)
})
```

### Wildcard Routes

```go
// Match /static/css/style.css or /static/js/app.js
r.GET("/static/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.String(http.StatusOK, filepath)
})
```

### Route Groups

```go
v1 := r.Group("/v1")
{
    v1.GET("/users", listUsers)
    v1.POST("/users", createUser)
    v1.GET("/users/:id", getUser)
}

v2 := r.Group("/v2")
{
    v2.GET("/users", listUsersV2)
    v2.POST("/users", createUserV2)
}
```

## Request Handling

### Query Parameters

```go
r.GET("/search", func(c *gin.Context) {
    // Default value if not provided
    query := c.DefaultQuery("q", "default")
    
    // Get query parameter (empty string if not exists)
    page := c.Query("page")
    
    // Get query parameter with existence check
    if name, ok := c.GetQuery("name"); ok {
        // name exists
    }
})
```

### Form Data

```go
r.POST("/form", func(c *gin.Context) {
    // Get form value
    username := c.PostForm("username")
    
    // Get with default value
    password := c.DefaultPostForm("password", "guest")
    
    // Get with existence check
    if email, ok := c.GetPostForm("email"); ok {
        // email exists
    }
})
```

### JSON Binding

```go
type Login struct {
    User     string `json:"user" binding:"required"`
    Password string `json:"password" binding:"required"`
}

r.POST("/login", func(c *gin.Context) {
    var json Login
    
    // Bind JSON - Returns 400 if validation fails
    if err := c.ShouldBindJSON(&json); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"user": json.User})
})
```

### Query Binding

```go
type Query struct {
    Name string `form:"name" binding:"required"`
    Age  int    `form:"age" binding:"gte=0"`
}

r.GET("/query", func(c *gin.Context) {
    var query Query
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, query)
})
```

### URI Binding

```go
type User struct {
    ID int `uri:"id" binding:"required"`
}

r.GET("/users/:id", func(c *gin.Context) {
    var user User
    if err := c.ShouldBindUri(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"id": user.ID})
})
```

### File Upload

**Single file**:
```go
r.POST("/upload", func(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Save file
    c.SaveUploadedFile(file, "uploads/" + file.Filename)
    c.JSON(http.StatusOK, gin.H{"filename": file.Filename})
})
```

**Multiple files**:
```go
r.POST("/upload-multiple", func(c *gin.Context) {
    form, _ := c.MultipartForm()
    files := form.File["files"]
    
    for _, file := range files {
        c.SaveUploadedFile(file, "uploads/" + file.Filename)
    }
    c.JSON(http.StatusOK, gin.H{"count": len(files)})
})
```

## Response Rendering

### JSON Response

```go
// gin.H is shortcut for map[string]interface{}
c.JSON(http.StatusOK, gin.H{
    "message": "success",
    "data": data,
})

// Or use struct
c.JSON(http.StatusOK, User{Name: "John", Age: 30})

// Pretty JSON with indentation
c.IndentedJSON(http.StatusOK, data)

// Secure JSON (prevents JSON hijacking)
c.SecureJSON(http.StatusOK, data)

// ASCII-only JSON
c.AsciiJSON(http.StatusOK, data)

// Pure JSON (don't replace special chars)
c.PureJSON(http.StatusOK, data)
```

### XML Response

```go
c.XML(http.StatusOK, gin.H{"message": "success"})
```

### YAML Response

```go
c.YAML(http.StatusOK, gin.H{"message": "success"})
```

### TOML Response

```go
c.TOML(http.StatusOK, gin.H{"message": "success"})
```

### String Response

```go
c.String(http.StatusOK, "Hello %s", name)
```

### HTML Response

```go
r.LoadHTMLGlob("templates/*")

r.GET("/index", func(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "Main Page",
    })
})
```

### File Response

```go
// Serve file
c.File("/path/to/file.pdf")

// Serve file as attachment (download)
c.FileAttachment("/path/to/file.pdf", "download.pdf")

// Serve from filesystem
c.FileFromFS("index.html", http.Dir("./public"))
```

### Redirect

```go
// HTTP redirect
c.Redirect(http.StatusMovedPermanently, "https://google.com")

// Router redirect
c.Request.URL.Path = "/new-path"
r.HandleContext(c)
```

## Middleware

### Custom Middleware

```go
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "authorization required",
            })
            return
        }
        
        // Continue to next handler
        c.Next()
    }
}

// Apply globally
r.Use(AuthRequired())

// Apply to route group
authorized := r.Group("/api")
authorized.Use(AuthRequired())
{
    authorized.GET("/users", listUsers)
}

// Apply to specific route
r.GET("/protected", AuthRequired(), protectedHandler)
```

### Middleware Execution Flow

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before request
        
        c.Next() // Execute remaining handlers
        
        // After request
    }
}

func AbortMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !authorized {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        c.Next()
    }
}
```

### Built-in Middleware

**Basic Auth**:
```go
authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
    "admin": "secret",
    "user":  "password",
}))
```

**CORS** (requires `github.com/gin-contrib/cors`):
```go
import "github.com/gin-contrib/cors"

r.Use(cors.Default())
// Or custom config
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://example.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Origin"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
}))
```

## Context Methods

### Set/Get Values

```go
// Store value in context
c.Set("user", user)

// Retrieve value
if val, exists := c.Get("user"); exists {
    user := val.(User)
}

// MustGet (panics if not exists)
user := c.MustGet("user").(User)
```

### Client Information

```go
// Client IP
ip := c.ClientIP()

// Content type
contentType := c.ContentType()

// Request header
userAgent := c.GetHeader("User-Agent")

// Check if WebSocket
isWebsocket := c.IsWebsocket()
```

### Cookies

```go
// Set cookie
c.SetCookie(
    "session",      // name
    "value",        // value
    3600,           // max age (seconds)
    "/",            // path
    "localhost",    // domain
    false,          // secure
    true,           // httpOnly
)

// Get cookie
value, err := c.Cookie("session")
```

### Request Body

```go
// Read body bytes (can only be read once)
bodyBytes, _ := c.GetRawData()

// To allow multiple reads, bind to GetRawData and bind again
c.Set(gin.BodyBytesKey, bodyBytes)
```

## Error Handling

### Custom Error Responses

```go
r.GET("/error", func(c *gin.Context) {
    // Attach error to context
    c.Error(errors.New("something went wrong"))
    
    // Abort with error
    c.AbortWithError(http.StatusInternalServerError, errors.New("error"))
    
    // Abort with JSON
    c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
        "error": "invalid input",
    })
})
```

### Error Middleware

```go
r.Use(func(c *gin.Context) {
    c.Next()
    
    // Check for errors after handlers executed
    if len(c.Errors) > 0 {
        c.JSON(http.StatusInternalServerError, gin.H{
            "errors": c.Errors,
        })
    }
})
```

## HTML Templates

### Load Templates

```go
// Load all templates
r.LoadHTMLGlob("templates/**/*")

// Load specific files
r.LoadHTMLFiles("templates/index.html", "templates/about.html")
```

### Custom Delimiters

```go
r.Delims("{[{", "}]}")
```

### Custom Functions

```go
import "html/template"

r.SetFuncMap(template.FuncMap{
    "formatDate": func(t time.Time) string {
        return t.Format("2006-01-02")
    },
})
```

### Render Template

```go
r.GET("/", func(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "Home",
        "user": user,
    })
})
```

## Static Files

```go
// Serve single file
r.StaticFile("/favicon.ico", "./resources/favicon.ico")

// Serve directory
r.Static("/assets", "./assets")

// Serve from embedded FS (Go 1.16+)
r.StaticFS("/public", http.FS(embeddedFS))
```

## Testing

```go
import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestPingRoute(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    r := setupRouter()
    
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/ping", nil)
    r.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "pong", w.Body.String())
}
```

## Advanced Features

### Graceful Shutdown

```go
import (
    "context"
    "net/http"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    r := gin.Default()
    
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
}
```

### Custom Validators

```go
import "github.com/go-playground/validator/v10"

type Booking struct {
    CheckIn  time.Time `binding:"required,bookabledate"`
    CheckOut time.Time `binding:"required,gtfield=CheckIn"`
}

var bookableDate validator.Func = func(fl validator.FieldLevel) bool {
    date, ok := fl.Field().Interface().(time.Time)
    if ok {
        return date.After(time.Now())
    }
    return false
}

// Register custom validator
if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
    v.RegisterValidation("bookabledate", bookableDate)
}
```

### Run Multiple Services

```go
func main() {
    r1 := gin.Default()
    r2 := gin.Default()
    
    s1 := &http.Server{Addr: ":8080", Handler: r1}
    s2 := &http.Server{Addr: ":8081", Handler: r2}
    
    g := new(errgroup.Group)
    g.Go(func() error { return s1.ListenAndServe() })
    g.Go(func() error { return s2.ListenAndServe() })
    
    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }
}
```

### Auto TLS (HTTPS)

```go
import "github.com/gin-gonic/autotls"

func main() {
    r := gin.Default()
    
    // Automatically obtain TLS certificates from Let's Encrypt
    log.Fatal(autotls.Run(r, "example1.com", "example2.com"))
}
```

## Common Patterns

### API Versioning

```go
func main() {
    r := gin.Default()
    
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", v1GetUsers)
        v1.POST("/users", v1CreateUser)
    }
    
    v2 := r.Group("/api/v2")
    {
        v2.GET("/users", v2GetUsers)
        v2.POST("/users", v2CreateUser)
    }
    
    r.Run()
}
```

### Structured Responses

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Error(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Code:    code,
        Message: message,
    })
}
```

### Pagination

```go
type PaginationQuery struct {
    Page     int `form:"page" binding:"min=1"`
    PageSize int `form:"page_size" binding:"min=1,max=100"`
}

r.GET("/users", func(c *gin.Context) {
    var query PaginationQuery
    query.Page = 1
    query.PageSize = 10
    
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    offset := (query.Page - 1) * query.PageSize
    users := getUsersFromDB(query.PageSize, offset)
    
    c.JSON(http.StatusOK, gin.H{
        "page":      query.Page,
        "page_size": query.PageSize,
        "data":      users,
    })
})
```

## Configuration

### Gin Mode

```go
// Set mode programmatically
gin.SetMode(gin.ReleaseMode)  // Production
gin.SetMode(gin.DebugMode)    // Development
gin.SetMode(gin.TestMode)     // Testing

// Or via environment variable
// export GIN_MODE=release
```

### Disable Console Color

```go
gin.DisableConsoleColor()
```

### Custom Logger

```go
import (
    "io"
    "os"
)

// Write logs to file
f, _ := os.Create("gin.log")
gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

// Disable logs
gin.DefaultWriter = io.Discard
```

### Trusted Proxies

```go
r := gin.Default()
r.SetTrustedProxies([]string{"192.168.1.0/24"})
```

## Performance Tips

1. **Use `gin.New()` instead of `gin.Default()`** if you don't need logger/recovery middleware
2. **Disable console colors in production**: `gin.DisableConsoleColor()`
3. **Set release mode**: `gin.SetMode(gin.ReleaseMode)`
4. **Use `c.ShouldBind*` instead of `c.Bind*`** for better error handling control
5. **Reuse gin.Context carefully** - Use `c.Copy()` for goroutines
6. **Limit request body size** with middleware or nginx

## Examples Reference

See the `references/examples.md` file for additional working examples covering:
- WebSocket integration
- Server-Sent Events (SSE)
- gRPC integration
- Real-time chat
- OIDC authentication
- Rate limiting
- Reverse proxy
- And more

For complete API reference, see `references/api.md`.
