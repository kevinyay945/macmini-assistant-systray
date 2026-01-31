# Gin Framework Examples

This file provides detailed examples for common Gin web framework patterns and use cases.

## Table of Contents

- [WebSocket Integration](#websocket-integration)
- [Server-Sent Events (SSE)](#server-sent-events-sse)
- [File Upload & Download](#file-upload--download)
- [Template Rendering](#template-rendering)
- [Route Grouping](#route-grouping)
- [Custom Validation](#custom-validation)
- [Authentication & Authorization](#authentication--authorization)
- [CORS Middleware](#cors-middleware)
- [Rate Limiting](#rate-limiting)
- [Graceful Shutdown](#graceful-shutdown)
- [Multiple Services](#multiple-services)

## WebSocket Integration

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "net/http"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins in development
    },
}

func wsHandler(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer conn.Close()

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Echo back the message
        if err := conn.WriteMessage(messageType, message); err != nil {
            break
        }
    }
}

func main() {
    r := gin.Default()
    r.GET("/ws", wsHandler)
    r.Run(":8080")
}
```

## Server-Sent Events (SSE)

```go
package main

import (
    "fmt"
    "io"
    "time"
    "github.com/gin-gonic/gin"
)

func streamEvents(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    clientGone := c.Request.Context().Done()
    c.Stream(func(w io.Writer) bool {
        select {
        case <-clientGone:
            return false
        case t := <-ticker.C:
            c.SSEvent("message", fmt.Sprintf("Current time: %s", t.Format(time.RFC3339)))
            return true
        }
    })
}

func main() {
    r := gin.Default()
    r.GET("/events", streamEvents)
    r.Run(":8080")
}
```

## File Upload & Download

### Single File Upload

```go
func uploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }

    // Validate file type
    if file.Header.Get("Content-Type") != "image/jpeg" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG images allowed"})
        return
    }

    // Limit file size (10MB)
    if file.Size > 10*1024*1024 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
        return
    }

    filename := filepath.Base(file.Filename)
    if err := c.SaveUploadedFile(file, "uploads/"+filename); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "filename": filename,
        "size":     file.Size,
    })
}
```

### Multiple File Upload

```go
func uploadMultipleFiles(c *gin.Context) {
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    files := form.File["files"]
    uploaded := []string{}

    for _, file := range files {
        filename := filepath.Base(file.Filename)
        if err := c.SaveUploadedFile(file, "uploads/"+filename); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        uploaded = append(uploaded, filename)
    }

    c.JSON(http.StatusOK, gin.H{
        "count": len(uploaded),
        "files": uploaded,
    })
}
```

### File Download

```go
func downloadFile(c *gin.Context) {
    filename := c.Param("filename")
    filepath := "uploads/" + filename

    // Check if file exists
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        return
    }

    c.FileAttachment(filepath, filename)
}
```

## Template Rendering

### Basic Template

```go
package main

import (
    "html/template"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

func formatAsDate(t time.Time) string {
    return t.Format("2006-01-02")
}

func main() {
    r := gin.Default()
    
    // Custom delimiters to avoid conflicts with Vue/Angular
    r.Delims("{[{", "}]}")
    
    // Register custom functions
    r.SetFuncMap(template.FuncMap{
        "formatAsDate": formatAsDate,
    })
    
    // Load templates
    r.LoadHTMLGlob("templates/**/*")
    
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "title": "Home Page",
            "now":   time.Now(),
            "user":  "John Doe",
        })
    })
    
    r.Run(":8080")
}
```

### Template with Layout

```html
<!-- templates/layouts/base.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{[{ .title }]}</title>
</head>
<body>
    {[{ template "content" . }]}
</body>
</html>

<!-- templates/pages/home.html -->
{[{ define "content" }]}
<h1>Welcome {[{ .user }]}</h1>
<p>Current date: {[{ .now | formatAsDate }]}</p>
{[{ end }]}
```

## Route Grouping

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Public routes
    public := r.Group("/api/v1")
    {
        public.GET("/health", healthCheck)
        public.POST("/login", login)
        public.POST("/register", register)
    }

    // Authenticated routes
    auth := r.Group("/api/v1")
    auth.Use(AuthMiddleware())
    {
        auth.GET("/profile", getProfile)
        auth.PUT("/profile", updateProfile)
        
        // Admin routes
        admin := auth.Group("/admin")
        admin.Use(AdminMiddleware())
        {
            admin.GET("/users", listUsers)
            admin.DELETE("/users/:id", deleteUser)
        }
    }

    r.Run(":8080")
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "No authorization header",
            })
            return
        }
        
        // Validate token (pseudo-code)
        userID, valid := validateToken(token)
        if !valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid token",
            })
            return
        }
        
        c.Set("userID", userID)
        c.Next()
    }
}

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt("userID")
        
        // Check if user is admin (pseudo-code)
        if !isAdmin(userID) {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "Admin access required",
            })
            return
        }
        
        c.Next()
    }
}
```

## Custom Validation

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
    "time"
)

type Booking struct {
    CheckIn  time.Time `json:"check_in" binding:"required,bookabledate"`
    CheckOut time.Time `json:"check_out" binding:"required,gtfield=CheckIn"`
    Guests   int       `json:"guests" binding:"required,min=1,max=10"`
}

// Custom validator
var bookableDate validator.Func = func(fl validator.FieldLevel) bool {
    date, ok := fl.Field().Interface().(time.Time)
    if ok {
        today := time.Now()
        // Must be future date
        if date.Before(today) {
            return false
        }
        // Must be within next year
        if date.After(today.AddDate(1, 0, 0)) {
            return false
        }
        return true
    }
    return false
}

func main() {
    r := gin.Default()

    // Register custom validator
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("bookabledate", bookableDate)
    }

    r.POST("/bookings", func(c *gin.Context) {
        var booking Booking
        
        if err := c.ShouldBindJSON(&booking); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Booking created",
            "booking": booking,
        })
    })

    r.Run(":8080")
}
```

## Authentication & Authorization

### JWT Authentication

```go
package main

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key")

type Claims struct {
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func generateToken(userID int, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func validateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrSignatureInvalid
}

func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Authorization header required",
            })
            return
        }

        // Remove "Bearer " prefix
        tokenString := authHeader[7:]

        claims, err := validateToken(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid token",
            })
            return
        }

        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Next()
    }
}

func main() {
    r := gin.Default()

    r.POST("/login", func(c *gin.Context) {
        var loginData struct {
            Email    string `json:"email" binding:"required,email"`
            Password string `json:"password" binding:"required"`
        }

        if err := c.ShouldBindJSON(&loginData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Validate credentials (pseudo-code)
        userID, valid := validateCredentials(loginData.Email, loginData.Password)
        if !valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }

        token, err := generateToken(userID, loginData.Email)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"token": token})
    })

    protected := r.Group("/api")
    protected.Use(JWTAuthMiddleware())
    {
        protected.GET("/profile", func(c *gin.Context) {
            userID := c.GetInt("userID")
            email := c.GetString("email")
            
            c.JSON(http.StatusOK, gin.H{
                "user_id": userID,
                "email":   email,
            })
        })
    }

    r.Run(":8080")
}
```

## CORS Middleware

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func main() {
    r := gin.Default()

    // Default CORS config (allows all origins)
    // r.Use(cors.Default())

    // Custom CORS config
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    // Or allow all origins for development
    r.Use(cors.New(cors.Config{
        AllowAllOrigins: true,
        AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:    []string{"*"},
    }))

    r.GET("/api/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "CORS enabled"})
    })

    r.Run(":8080")
}
```

## Rate Limiting

```go
package main

import (
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
)

// Simple in-memory rate limiter
type RateLimiter struct {
    visitors map[string]*Visitor
    mu       sync.RWMutex
    rate     int
    window   time.Duration
}

type Visitor struct {
    count      int
    lastReset  time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        visitors: make(map[string]*Visitor),
        rate:     rate,
        window:   window,
    }
    
    // Cleanup old visitors periodically
    go rl.cleanup()
    
    return rl
}

func (rl *RateLimiter) cleanup() {
    for {
        time.Sleep(time.Minute)
        rl.mu.Lock()
        for ip, v := range rl.visitors {
            if time.Since(v.lastReset) > rl.window {
                delete(rl.visitors, ip)
            }
        }
        rl.mu.Unlock()
    }
}

func (rl *RateLimiter) Allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    v, exists := rl.visitors[ip]
    if !exists {
        rl.visitors[ip] = &Visitor{
            count:     1,
            lastReset: time.Now(),
        }
        return true
    }
    
    if time.Since(v.lastReset) > rl.window {
        v.count = 1
        v.lastReset = time.Now()
        return true
    }
    
    if v.count >= rl.rate {
        return false
    }
    
    v.count++
    return true
}

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        
        if !rl.Allow(ip) {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            return
        }
        
        c.Next()
    }
}

func main() {
    r := gin.Default()
    
    // Allow 10 requests per minute
    limiter := NewRateLimiter(10, time.Minute)
    r.Use(RateLimitMiddleware(limiter))
    
    r.GET("/api/data", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Success"})
    })
    
    r.Run(":8080")
}
```

## Graceful Shutdown

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    r.GET("/", func(c *gin.Context) {
        time.Sleep(2 * time.Second) // Simulate work
        c.JSON(http.StatusOK, gin.H{"message": "Hello"})
    })
    
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }
    
    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()
    
    log.Println("Server started on :8080")
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Give ongoing requests 5 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

## Multiple Services

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "golang.org/x/sync/errgroup"
)

func main() {
    // API server
    apiRouter := gin.Default()
    apiRouter.GET("/api/users", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"users": []string{"Alice", "Bob"}})
    })
    
    // Admin server
    adminRouter := gin.Default()
    adminRouter.GET("/admin/stats", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"stats": "data"})
    })
    
    apiServer := &http.Server{
        Addr:    ":8080",
        Handler: apiRouter,
    }
    
    adminServer := &http.Server{
        Addr:    ":8081",
        Handler: adminRouter,
    }
    
    g, ctx := errgroup.WithContext(context.Background())
    
    // Start API server
    g.Go(func() error {
        log.Println("API server starting on :8080")
        if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            return err
        }
        return nil
    })
    
    // Start Admin server
    g.Go(func() error {
        log.Println("Admin server starting on :8081")
        if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            return err
        }
        return nil
    })
    
    // Wait for context cancellation or error
    g.Go(func() error {
        <-ctx.Done()
        
        log.Println("Shutting down servers...")
        
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := apiServer.Shutdown(shutdownCtx); err != nil {
            return err
        }
        if err := adminServer.Shutdown(shutdownCtx); err != nil {
            return err
        }
        
        return nil
    })
    
    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }
}
```

## Additional Resources

For more examples, see the official Gin examples repository:
- Basic usage
- Custom validators
- Form binding
- gRPC integration
- OIDC authentication
- OpenTelemetry integration
- Reverse proxy
- Secure web applications
- And many more

All examples are available in the `examples/` directory of the Gin project.
