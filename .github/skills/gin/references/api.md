======================================================================
 PACKAGE: github.com/gin-gonic/gin
======================================================================

package gin // import "github.com/gin-gonic/gin"

Package gin implements a HTTP web framework called gin.

See https://gin-gonic.com/ for more information about gin.

CONSTANTS

const (
	MIMEJSON              = binding.MIMEJSON
	MIMEHTML              = binding.MIMEHTML
	MIMEXML               = binding.MIMEXML
	MIMEXML2              = binding.MIMEXML2
	MIMEPlain             = binding.MIMEPlain
	MIMEPOSTForm          = binding.MIMEPOSTForm
	MIMEMultipartPOSTForm = binding.MIMEMultipartPOSTForm
	MIMEYAML              = binding.MIMEYAML
	MIMETOML              = binding.MIMETOML
)
    Content-Type MIME of the most common data formats.

const (
	// PlatformGoogleAppEngine when running on Google App Engine. Trust X-Appengine-Remote-Addr
	// for determining the client's IP
	PlatformGoogleAppEngine = "X-Appengine-Remote-Addr"
	// PlatformCloudflare when using Cloudflare's CDN. Trust CF-Connecting-IP for determining
	// the client's IP
	PlatformCloudflare = "CF-Connecting-IP"
	// PlatformFlyIO when running on Fly.io. Trust Fly-Client-IP for determining the client's IP
	PlatformFlyIO = "Fly-Client-IP"
)
    Trusted platforms

const (
	// DebugMode indicates gin mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates gin mode is release.
	ReleaseMode = "release"
	// TestMode indicates gin mode is test.
	TestMode = "test"
)
const AuthProxyUserKey = "proxy_user"
    AuthProxyUserKey is the cookie name for proxy_user credential in basic auth
    for proxy.

const AuthUserKey = "user"
    AuthUserKey is the cookie name for user credential in basic auth.

const BindKey = "_gin-gonic/gin/bindkey"
    BindKey indicates a default bind key.

const BodyBytesKey = "_gin-gonic/gin/bodybyteskey"
    BodyBytesKey indicates a default body bytes key.

const ContextKey = "_gin-gonic/gin/contextkey"
    ContextKey is the key that a Context returns itself for.

const EnvGinMode = "GIN_MODE"
    EnvGinMode indicates environment name for gin mode.

const Version = "v1.10.0"
    Version is the current gin framework's version.


VARIABLES

var DebugPrintFunc func(format string, values ...interface{})
    DebugPrintFunc indicates debug log output format.

var DebugPrintRouteFunc func(httpMethod, absolutePath, handlerName string, nuHandlers int)
    DebugPrintRouteFunc indicates debug log output format.

var DefaultErrorWriter io.Writer = os.Stderr
    DefaultErrorWriter is the default io.Writer used by Gin to debug errors

var DefaultWriter io.Writer = os.Stdout
    DefaultWriter is the default io.Writer used by Gin for debug output and
    middleware output like Logger() or Recovery(). Note that both Logger
    and Recovery provides custom ways to configure their output io.Writer.
    To support coloring in Windows use:

        import "github.com/mattn/go-colorable"
        gin.DefaultWriter = colorable.NewColorableStdout()


FUNCTIONS

func CreateTestContext(w http.ResponseWriter) (c *Context, r *Engine)
    CreateTestContext returns a fresh engine and context for testing purposes

func Dir(root string, listDirectory bool) http.FileSystem
    Dir returns a http.FileSystem that can be used by http.FileServer().
    It is used internally in router.Static(). if listDirectory == true, then it
    works the same as http.Dir() otherwise it returns a filesystem that prevents
    http.FileServer() to list the directory files.

func DisableBindValidation()
    DisableBindValidation closes the default validator.

func DisableConsoleColor()
    DisableConsoleColor disables color output in the console.

func EnableJsonDecoderDisallowUnknownFields()
    EnableJsonDecoderDisallowUnknownFields sets true for
    binding.EnableDecoderDisallowUnknownFields to call the DisallowUnknownFields
    method on the JSON Decoder instance.

func EnableJsonDecoderUseNumber()
    EnableJsonDecoderUseNumber sets true for binding.EnableDecoderUseNumber to
    call the UseNumber method on the JSON Decoder instance.

func ForceConsoleColor()
    ForceConsoleColor force color output in the console.

func IsDebugging() bool
    IsDebugging returns true if the framework is running in debug mode.
    Use SetMode(gin.ReleaseMode) to disable debug mode.

func Mode() string
    Mode returns current gin mode.

func SetMode(value string)
    SetMode sets gin mode according to input string.


TYPES

type Accounts map[string]string
    Accounts defines a key/value for user/pass list of authorized logins.

type Context struct {
	Request *http.Request
	Writer  ResponseWriter

	Params Params

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs

	// Accepted defines a list of manually accepted formats for content negotiation.
	Accepted []string

	// Has unexported fields.
}
    Context is the most important part of gin. It allows us to pass variables
    between middleware, manage the flow, validate the JSON of a request and
    render a JSON response for example.

func CreateTestContextOnly(w http.ResponseWriter, r *Engine) (c *Context)
    CreateTestContextOnly returns a fresh context base on the engine for testing
    purposes

func (c *Context) Abort()
    Abort prevents pending handlers from being called. Note that this will not
    stop the current handler. Let's say you have an authorization middleware
    that validates that the current request is authorized. If the authorization
    fails (ex: the password does not match), call Abort to ensure the remaining
    handlers for this request are not called.

func (c *Context) AbortWithError(code int, err error) *Error
    AbortWithError calls `AbortWithStatus()` and `Error()` internally.
    This method stops the chain, writes the status code and pushes the specified
    error to `c.Errors`. See Context.Error() for more details.

func (c *Context) AbortWithStatus(code int)
    AbortWithStatus calls `Abort()` and writes the headers with the specified
    status code. For example, a failed attempt to authenticate a request could
    use: context.AbortWithStatus(401).

func (c *Context) AbortWithStatusJSON(code int, jsonObj any)
    AbortWithStatusJSON calls `Abort()` and then `JSON` internally. This method
    stops the chain, writes the status code and return a JSON body. It also sets
    the Content-Type as "application/json".

func (c *Context) AddParam(key, value string)
    AddParam adds param to context and replaces path param key with given
    value for e2e testing purposes Example Route: "/user/:id" AddParam("id",
    1) Result: "/user/1"

func (c *Context) AsciiJSON(code int, obj any)
    AsciiJSON serializes the given struct as JSON into the response
    body with unicode to ASCII string. It also sets the Content-Type as
    "application/json".

func (c *Context) Bind(obj any) error
    Bind checks the Method and Content-Type to select a binding engine
    automatically, Depending on the "Content-Type" header different bindings are
    used, for example:

        "application/json" --> JSON binding
        "application/xml"  --> XML binding

    It parses the request's body as JSON if Content-Type == "application/json"
    using JSON or XML as a JSON input. It decodes the json payload into the
    struct specified as a pointer. It writes a 400 error and sets Content-Type
    header "text/plain" in the response if input is not valid.

func (c *Context) BindHeader(obj any) error
    BindHeader is a shortcut for c.MustBindWith(obj, binding.Header).

func (c *Context) BindJSON(obj any) error
    BindJSON is a shortcut for c.MustBindWith(obj, binding.JSON).

func (c *Context) BindQuery(obj any) error
    BindQuery is a shortcut for c.MustBindWith(obj, binding.Query).

func (c *Context) BindTOML(obj any) error
    BindTOML is a shortcut for c.MustBindWith(obj, binding.TOML).

func (c *Context) BindUri(obj any) error
    BindUri binds the passed struct pointer using binding.Uri. It will abort the
    request with HTTP 400 if any error occurs.

func (c *Context) BindWith(obj any, b binding.Binding) error
    BindWith binds the passed struct pointer using the specified binding engine.
    See the binding package.

    Deprecated: Use MustBindWith or ShouldBindWith.

func (c *Context) BindXML(obj any) error
    BindXML is a shortcut for c.MustBindWith(obj, binding.BindXML).

func (c *Context) BindYAML(obj any) error
    BindYAML is a shortcut for c.MustBindWith(obj, binding.YAML).

func (c *Context) ClientIP() string
    ClientIP implements one best effort algorithm to return the real client IP.
    It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted
    proxy or not. If it is it will then try to parse the headers defined in
    Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]). If the
    headers are not syntactically valid OR the remote IP does not correspond to
    a trusted proxy, the remote IP (coming from Request.RemoteAddr) is returned.

func (c *Context) ContentType() string
    ContentType returns the Content-Type header of the request.

func (c *Context) Cookie(name string) (string, error)
    Cookie returns the named cookie provided in the request or ErrNoCookie if
    not found. And return the named cookie is unescaped. If multiple cookies
    match the given name, only one cookie will be returned.

func (c *Context) Copy() *Context
    Copy returns a copy of the current context that can be safely used outside
    the request's scope. This has to be used when the context has to be passed
    to a goroutine.

func (c *Context) Data(code int, contentType string, data []byte)
    Data writes some data into the body stream and updates the HTTP code.

func (c *Context) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string)
    DataFromReader writes the specified reader into the body stream and updates
    the HTTP code.

func (c *Context) Deadline() (deadline time.Time, ok bool)
    Deadline returns that there is no deadline (ok==false) when c.Request has no
    Context.

func (c *Context) DefaultPostForm(key, defaultValue string) string
    DefaultPostForm returns the specified key from a POST urlencoded form
    or multipart form when it exists, otherwise it returns the specified
    defaultValue string. See: PostForm() and GetPostForm() for further
    information.

func (c *Context) DefaultQuery(key, defaultValue string) string
    DefaultQuery returns the keyed url query value if it exists, otherwise it
    returns the specified defaultValue string. See: Query() and GetQuery() for
    further information.

        GET /?name=Manu&lastname=
        c.DefaultQuery("name", "unknown") == "Manu"
        c.DefaultQuery("id", "none") == "none"
        c.DefaultQuery("lastname", "none") == ""

func (c *Context) Done() <-chan struct{}
    Done returns nil (chan which will wait forever) when c.Request has no
    Context.

func (c *Context) Err() error
    Err returns nil when c.Request has no Context.

func (c *Context) Error(err error) *Error
    Error attaches an error to the current context. The error is pushed to a
    list of errors. It's a good idea to call Error for each error that occurred
    during the resolution of a request. A middleware can be used to collect all
    the errors and push them to a database together, print a log, or append it
    in the HTTP response. Error will panic if err is nil.

func (c *Context) File(filepath string)
    File writes the specified file into the body stream in an efficient way.

func (c *Context) FileAttachment(filepath, filename string)
    FileAttachment writes the specified file into the body stream in an
    efficient way On the client side, the file will typically be downloaded with
    the given filename

func (c *Context) FileFromFS(filepath string, fs http.FileSystem)
    FileFromFS writes the specified file from http.FileSystem into the body
    stream in an efficient way.

func (c *Context) FormFile(name string) (*multipart.FileHeader, error)
    FormFile returns the first file for the provided form key.

func (c *Context) FullPath() string
    FullPath returns a matched route full path. For not found routes returns an
    empty string.

        router.GET("/user/:id", func(c *gin.Context) {
            c.FullPath() == "/user/:id" // true
        })

func (c *Context) Get(key string) (value any, exists bool)
    Get returns the value for the given key, ie: (value, true). If the value
    does not exist it returns (nil, false)

func (c *Context) GetBool(key string) (b bool)
    GetBool returns the value associated with the key as a boolean.

func (c *Context) GetDuration(key string) (d time.Duration)
    GetDuration returns the value associated with the key as a duration.

func (c *Context) GetFloat64(key string) (f64 float64)
    GetFloat64 returns the value associated with the key as a float64.

func (c *Context) GetHeader(key string) string
    GetHeader returns value from request headers.

func (c *Context) GetInt(key string) (i int)
    GetInt returns the value associated with the key as an integer.

func (c *Context) GetInt64(key string) (i64 int64)
    GetInt64 returns the value associated with the key as an integer.

func (c *Context) GetPostForm(key string) (string, bool)
    GetPostForm is like PostForm(key). It returns the specified key from a POST
    urlencoded form or multipart form when it exists `(value, true)` (even
    when the value is an empty string), otherwise it returns ("", false).
    For example, during a PATCH request to update the user's email:

            email=mail@example.com  -->  ("mail@example.com", true) := GetPostForm("email") // set email to "mail@example.com"
        	   email=                  -->  ("", true) := GetPostForm("email") // set email to ""
                                    -->  ("", false) := GetPostForm("email") // do nothing with email

func (c *Context) GetPostFormArray(key string) (values []string, ok bool)
    GetPostFormArray returns a slice of strings for a given form key, plus a
    boolean value whether at least one value exists for the given key.

func (c *Context) GetPostFormMap(key string) (map[string]string, bool)
    GetPostFormMap returns a map for a given form key, plus a boolean value
    whether at least one value exists for the given key.

func (c *Context) GetQuery(key string) (string, bool)
    GetQuery is like Query(), it returns the keyed url query value if it exists
    `(value, true)` (even when the value is an empty string), otherwise it
    returns `("", false)`. It is shortcut for `c.Request.URL.Query().Get(key)`

        GET /?name=Manu&lastname=
        ("Manu", true) == c.GetQuery("name")
        ("", false) == c.GetQuery("id")
        ("", true) == c.GetQuery("lastname")

func (c *Context) GetQueryArray(key string) (values []string, ok bool)
    GetQueryArray returns a slice of strings for a given query key, plus a
    boolean value whether at least one value exists for the given key.

func (c *Context) GetQueryMap(key string) (map[string]string, bool)
    GetQueryMap returns a map for a given query key, plus a boolean value
    whether at least one value exists for the given key.

func (c *Context) GetRawData() ([]byte, error)
    GetRawData returns stream data.

func (c *Context) GetString(key string) (s string)
    GetString returns the value associated with the key as a string.

func (c *Context) GetStringMap(key string) (sm map[string]any)
    GetStringMap returns the value associated with the key as a map of
    interfaces.

func (c *Context) GetStringMapString(key string) (sms map[string]string)
    GetStringMapString returns the value associated with the key as a map of
    strings.

func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string)
    GetStringMapStringSlice returns the value associated with the key as a map
    to a slice of strings.

func (c *Context) GetStringSlice(key string) (ss []string)
    GetStringSlice returns the value associated with the key as a slice of
    strings.

func (c *Context) GetTime(key string) (t time.Time)
    GetTime returns the value associated with the key as time.

func (c *Context) GetUint(key string) (ui uint)
    GetUint returns the value associated with the key as an unsigned integer.

func (c *Context) GetUint64(key string) (ui64 uint64)
    GetUint64 returns the value associated with the key as an unsigned integer.

func (c *Context) HTML(code int, name string, obj any)
    HTML renders the HTTP template specified by its file name. It also
    updates the HTTP code and sets the Content-Type as "text/html". See
    http://golang.org/doc/articles/wiki/

func (c *Context) Handler() HandlerFunc
    Handler returns the main handler.

func (c *Context) HandlerName() string
    HandlerName returns the main handler's name. For example if the handler is
    "handleGetUsers()", this function will return "main.handleGetUsers".

func (c *Context) HandlerNames() []string
    HandlerNames returns a list of all registered handlers for this context in
    descending order, following the semantics of HandlerName()

func (c *Context) Header(key, value string)
    Header is an intelligent shortcut for c.Writer.Header().Set(key, value).
    It writes a header in the response. If value == "", this method removes the
    header `c.Writer.Header().Del(key)`

func (c *Context) IndentedJSON(code int, obj any)
    IndentedJSON serializes the given struct as pretty JSON (indented +
    endlines) into the response body. It also sets the Content-Type as
    "application/json". WARNING: we recommend using this only for development
    purposes since printing pretty JSON is more CPU and bandwidth consuming.
    Use Context.JSON() instead.

func (c *Context) IsAborted() bool
    IsAborted returns true if the current context was aborted.

func (c *Context) IsWebsocket() bool
    IsWebsocket returns true if the request headers indicate that a websocket
    handshake is being initiated by the client.

func (c *Context) JSON(code int, obj any)
    JSON serializes the given struct as JSON into the response body. It also
    sets the Content-Type as "application/json".

func (c *Context) JSONP(code int, obj any)
    JSONP serializes the given struct as JSON into the response body.
    It adds padding to response body to request data from a server residing
    in a different domain than the client. It also sets the Content-Type as
    "application/javascript".

func (c *Context) MultipartForm() (*multipart.Form, error)
    MultipartForm is the parsed multipart form, including file uploads.

func (c *Context) MustBindWith(obj any, b binding.Binding) error
    MustBindWith binds the passed struct pointer using the specified binding
    engine. It will abort the request with HTTP 400 if any error occurs. See the
    binding package.

func (c *Context) MustGet(key string) any
    MustGet returns the value for the given key if it exists, otherwise it
    panics.

func (c *Context) Negotiate(code int, config Negotiate)
    Negotiate calls different Render according to acceptable Accept format.

func (c *Context) NegotiateFormat(offered ...string) string
    NegotiateFormat returns an acceptable Accept format.

func (c *Context) Next()
    Next should be used only inside middleware. It executes the pending handlers
    in the chain inside the calling handler. See example in GitHub.

func (c *Context) Param(key string) string
    Param returns the value of the URL param. It is a shortcut for
    c.Params.ByName(key)

        router.GET("/user/:id", func(c *gin.Context) {
            // a GET request to /user/john
            id := c.Param("id") // id == "john"
            // a GET request to /user/john/
            id := c.Param("id") // id == "/john/"
        })

func (c *Context) PostForm(key string) (value string)
    PostForm returns the specified key from a POST urlencoded form or multipart
    form when it exists, otherwise it returns an empty string `("")`.

func (c *Context) PostFormArray(key string) (values []string)
    PostFormArray returns a slice of strings for a given form key. The length of
    the slice depends on the number of params with the given key.

func (c *Context) PostFormMap(key string) (dicts map[string]string)
    PostFormMap returns a map for a given form key.

func (c *Context) ProtoBuf(code int, obj any)
    ProtoBuf serializes the given struct as ProtoBuf into the response body.

func (c *Context) PureJSON(code int, obj any)
    PureJSON serializes the given struct as JSON into the response body.
    PureJSON, unlike JSON, does not replace special html characters with their
    unicode entities.

func (c *Context) Query(key string) (value string)
    Query returns the keyed url query value if it exists, otherwise it returns
    an empty string `("")`. It is shortcut for `c.Request.URL.Query().Get(key)`

            GET /path?id=1234&name=Manu&value=
        	   c.Query("id") == "1234"
        	   c.Query("name") == "Manu"
        	   c.Query("value") == ""
        	   c.Query("wtf") == ""

func (c *Context) QueryArray(key string) (values []string)
    QueryArray returns a slice of strings for a given query key. The length of
    the slice depends on the number of params with the given key.

func (c *Context) QueryMap(key string) (dicts map[string]string)
    QueryMap returns a map for a given query key.

func (c *Context) Redirect(code int, location string)
    Redirect returns an HTTP redirect to the specific location.

func (c *Context) RemoteIP() string
    RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the
    IP (without the port).

func (c *Context) Render(code int, r render.Render)
    Render writes the response headers and calls render.Render to render data.

func (c *Context) SSEvent(name string, message any)
    SSEvent writes a Server-Sent Event into the body stream.

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error
    SaveUploadedFile uploads the form file to specific dst.

func (c *Context) SecureJSON(code int, obj any)
    SecureJSON serializes the given struct as Secure JSON into the response
    body. Default prepends "while(1)," to response body if the given struct is
    array values. It also sets the Content-Type as "application/json".

func (c *Context) Set(key string, value any)
    Set is used to store a new key/value pair exclusively for this context.
    It also lazy initializes c.Keys if it was not used previously.

func (c *Context) SetAccepted(formats ...string)
    SetAccepted sets Accept header data.

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
    SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
    The provided cookie must have a valid Name. Invalid cookies may be silently
    dropped.

func (c *Context) SetSameSite(samesite http.SameSite)
    SetSameSite with cookie

func (c *Context) ShouldBind(obj any) error
    ShouldBind checks the Method and Content-Type to select a binding engine
    automatically, Depending on the "Content-Type" header different bindings are
    used, for example:

        "application/json" --> JSON binding
        "application/xml"  --> XML binding

    It parses the request's body as JSON if Content-Type == "application/json"
    using JSON or XML as a JSON input. It decodes the json payload into the
    struct specified as a pointer. Like c.Bind() but this method does not set
    the response status code to 400 or abort if input is not valid.

func (c *Context) ShouldBindBodyWith(obj any, bb binding.BindingBody) (err error)
    ShouldBindBodyWith is similar with ShouldBindWith, but it stores the request
    body into the context, and reuse when it is called again.

    NOTE: This method reads the body before binding. So you should use
    ShouldBindWith for better performance if you need to call only once.

func (c *Context) ShouldBindBodyWithJSON(obj any) error
    ShouldBindBodyWithJSON is a shortcut for c.ShouldBindBodyWith(obj,
    binding.JSON).

func (c *Context) ShouldBindBodyWithTOML(obj any) error
    ShouldBindBodyWithTOML is a shortcut for c.ShouldBindBodyWith(obj,
    binding.TOML).

func (c *Context) ShouldBindBodyWithXML(obj any) error
    ShouldBindBodyWithXML is a shortcut for c.ShouldBindBodyWith(obj,
    binding.XML).

func (c *Context) ShouldBindBodyWithYAML(obj any) error
    ShouldBindBodyWithYAML is a shortcut for c.ShouldBindBodyWith(obj,
    binding.YAML).

func (c *Context) ShouldBindHeader(obj any) error
    ShouldBindHeader is a shortcut for c.ShouldBindWith(obj, binding.Header).

func (c *Context) ShouldBindJSON(obj any) error
    ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binding.JSON).

func (c *Context) ShouldBindQuery(obj any) error
    ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binding.Query).

func (c *Context) ShouldBindTOML(obj any) error
    ShouldBindTOML is a shortcut for c.ShouldBindWith(obj, binding.TOML).

func (c *Context) ShouldBindUri(obj any) error
    ShouldBindUri binds the passed struct pointer using the specified binding
    engine.

func (c *Context) ShouldBindWith(obj any, b binding.Binding) error
    ShouldBindWith binds the passed struct pointer using the specified binding
    engine. See the binding package.

func (c *Context) ShouldBindXML(obj any) error
    ShouldBindXML is a shortcut for c.ShouldBindWith(obj, binding.XML).

func (c *Context) ShouldBindYAML(obj any) error
    ShouldBindYAML is a shortcut for c.ShouldBindWith(obj, binding.YAML).

func (c *Context) Status(code int)
    Status sets the HTTP response code.

func (c *Context) Stream(step func(w io.Writer) bool) bool
    Stream sends a streaming response and returns a boolean indicates "Is client
    disconnected in middle of stream"

func (c *Context) String(code int, format string, values ...any)
    String writes the given string into the response body.

func (c *Context) TOML(code int, obj any)
    TOML serializes the given struct as TOML into the response body.

func (c *Context) Value(key any) any
    Value returns the value associated with this context for key, or nil if no
    value is associated with key. Successive calls to Value with the same key
    returns the same result.

func (c *Context) XML(code int, obj any)
    XML serializes the given struct as XML into the response body. It also sets
    the Content-Type as "application/xml".

func (c *Context) YAML(code int, obj any)
    YAML serializes the given struct as YAML into the response body.

type ContextKeyType int

const ContextRequestKey ContextKeyType = 0
type Engine struct {
	RouterGroup

	// RedirectTrailingSlash enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// RedirectFixedPath if enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// HandleMethodNotAllowed if enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// ForwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool

	// AppEngine was deprecated.
	// Deprecated: USE `TrustedPlatform` WITH VALUE `gin.PlatformGoogleAppEngine` INSTEAD
	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	AppEngine bool

	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// UnescapePathValues if true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool

	// RemoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	RemoteIPHeaders []string

	// TrustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	TrustedPlatform string

	// MaxMultipartMemory value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	// UseH2C enable h2c support.
	UseH2C bool

	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	ContextWithFallback bool

	HTMLRender render.HTMLRender
	FuncMap    template.FuncMap

	// Has unexported fields.
}
    Engine is the framework's instance, it contains the muxer, middleware and
    configuration settings. Create an instance of Engine, by using New() or
    Default()

func Default(opts ...OptionFunc) *Engine
    Default returns an Engine instance with the Logger and Recovery middleware
    already attached.

func New(opts ...OptionFunc) *Engine
    New returns a new blank Engine instance without any middleware
    attached. By default, the configuration is: - RedirectTrailingSlash:
    true - RedirectFixedPath: false - HandleMethodNotAllowed: false -
    ForwardedByClientIP: true - UseRawPath: false - UnescapePathValues: true

func (engine *Engine) Delims(left, right string) *Engine
    Delims sets template left and right delims and returns an Engine instance.

func (engine *Engine) HandleContext(c *Context)
    HandleContext re-enters a context that has been rewritten. This can be done
    by setting c.Request.URL.Path to your new target. Disclaimer: You can loop
    yourself to deal with this, use wisely.

func (engine *Engine) Handler() http.Handler

func (engine *Engine) LoadHTMLFiles(files ...string)
    LoadHTMLFiles loads a slice of HTML files and associates the result with
    HTML renderer.

func (engine *Engine) LoadHTMLGlob(pattern string)
    LoadHTMLGlob loads HTML files identified by glob pattern and associates the
    result with HTML renderer.

func (engine *Engine) NoMethod(handlers ...HandlerFunc)
    NoMethod sets the handlers called when Engine.HandleMethodNotAllowed = true.

func (engine *Engine) NoRoute(handlers ...HandlerFunc)
    NoRoute adds handlers for NoRoute. It returns a 404 code by default.

func (engine *Engine) Routes() (routes RoutesInfo)
    Routes returns a slice of registered routes, including some useful
    information, such as: the http method, path and the handler name.

func (engine *Engine) Run(addr ...string) (err error)
    Run attaches the router to a http.Server and starts listening and serving
    HTTP requests. It is a shortcut for http.ListenAndServe(addr, router) Note:
    this method will block the calling goroutine indefinitely unless an error
    happens.

func (engine *Engine) RunFd(fd int) (err error)
    RunFd attaches the router to a http.Server and starts listening and serving
    HTTP requests through the specified file descriptor. Note: this method will
    block the calling goroutine indefinitely unless an error happens.

func (engine *Engine) RunListener(listener net.Listener) (err error)
    RunListener attaches the router to a http.Server and starts listening and
    serving HTTP requests through the specified net.Listener

func (engine *Engine) RunTLS(addr, certFile, keyFile string) (err error)
    RunTLS attaches the router to a http.Server and starts listening and serving
    HTTPS (secure) requests. It is a shortcut for http.ListenAndServeTLS(addr,
    certFile, keyFile, router) Note: this method will block the calling
    goroutine indefinitely unless an error happens.

func (engine *Engine) RunUnix(file string) (err error)
    RunUnix attaches the router to a http.Server and starts listening and
    serving HTTP requests through the specified unix socket (i.e. a file). Note:
    this method will block the calling goroutine indefinitely unless an error
    happens.

func (engine *Engine) SecureJsonPrefix(prefix string) *Engine
    SecureJsonPrefix sets the secureJSONPrefix used in Context.SecureJSON.

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request)
    ServeHTTP conforms to the http.Handler interface.

func (engine *Engine) SetFuncMap(funcMap template.FuncMap)
    SetFuncMap sets the FuncMap used for template.FuncMap.

func (engine *Engine) SetHTMLTemplate(templ *template.Template)
    SetHTMLTemplate associate a template with HTML renderer.

func (engine *Engine) SetTrustedProxies(trustedProxies []string) error
    SetTrustedProxies set a list of network origins (IPv4 addresses, IPv4 CIDRs,
    IPv6 addresses or IPv6 CIDRs) from which to trust request's headers that
    contain alternative client IP when `(*gin.Engine).ForwardedByClientIP`
    is `true`. `TrustedProxies` feature is enabled by default, and it also
    trusts all proxies by default. If you want to disable this feature,
    use Engine.SetTrustedProxies(nil), then Context.ClientIP() will return the
    remote address directly.

func (engine *Engine) Use(middleware ...HandlerFunc) IRoutes
    Use attaches a global middleware to the router. i.e. the middleware attached
    through Use() will be included in the handlers chain for every single
    request. Even 404, 405, static files... For example, this is the right place
    for a logger or error management middleware.

func (engine *Engine) With(opts ...OptionFunc) *Engine
    With returns a new Engine instance with the provided options.

type Error struct {
	Err  error
	Type ErrorType
	Meta any
}
    Error represents a error's specification.

func (msg Error) Error() string
    Error implements the error interface.

func (msg *Error) IsType(flags ErrorType) bool
    IsType judges one error.

func (msg *Error) JSON() any
    JSON creates a properly formatted JSON

func (msg *Error) MarshalJSON() ([]byte, error)
    MarshalJSON implements the json.Marshaller interface.

func (msg *Error) SetMeta(data any) *Error
    SetMeta sets the error's meta data.

func (msg *Error) SetType(flags ErrorType) *Error
    SetType sets the error's type.

func (msg *Error) Unwrap() error
    Unwrap returns the wrapped error, to allow interoperability with
    errors.Is(), errors.As() and errors.Unwrap()

type ErrorType uint64
    ErrorType is an unsigned 64-bit error code as defined in the gin spec.

const (
	// ErrorTypeBind is used when Context.Bind() fails.
	ErrorTypeBind ErrorType = 1 << 63
	// ErrorTypeRender is used when Context.Render() fails.
	ErrorTypeRender ErrorType = 1 << 62
	// ErrorTypePrivate indicates a private error.
	ErrorTypePrivate ErrorType = 1 << 0
	// ErrorTypePublic indicates a public error.
	ErrorTypePublic ErrorType = 1 << 1
	// ErrorTypeAny indicates any other error.
	ErrorTypeAny ErrorType = 1<<64 - 1
	// ErrorTypeNu indicates any other error.
	ErrorTypeNu = 2
)
type H map[string]any
    H is a shortcut for map[string]any

func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error
    MarshalXML allows type H to be used with xml.Marshal.

type HandlerFunc func(*Context)
    HandlerFunc defines the handler used by gin middleware as return value.

func BasicAuth(accounts Accounts) HandlerFunc
    BasicAuth returns a Basic HTTP Authorization middleware. It takes as
    argument a map[string]string where the key is the user name and the value is
    the password.

func BasicAuthForProxy(accounts Accounts, realm string) HandlerFunc
    BasicAuthForProxy returns a Basic HTTP Proxy-Authorization middleware. If
    the realm is empty, "Proxy Authorization Required" will be used by default.

func BasicAuthForRealm(accounts Accounts, realm string) HandlerFunc
    BasicAuthForRealm returns a Basic HTTP Authorization middleware.
    It takes as arguments a map[string]string where the key is the user
    name and the value is the password, as well as the name of the Realm.
    If the realm is empty, "Authorization Required" will be used by default.
    (see http://tools.ietf.org/html/rfc2617#section-1.2)

func Bind(val any) HandlerFunc
    Bind is a helper function for given interface object and returns a Gin
    middleware.

func CustomRecovery(handle RecoveryFunc) HandlerFunc
    CustomRecovery returns a middleware that recovers from any panics and calls
    the provided handle func to handle it.

func CustomRecoveryWithWriter(out io.Writer, handle RecoveryFunc) HandlerFunc
    CustomRecoveryWithWriter returns a middleware for a given writer that
    recovers from any panics and calls the provided handle func to handle it.

func ErrorLogger() HandlerFunc
    ErrorLogger returns a HandlerFunc for any error type.

func ErrorLoggerT(typ ErrorType) HandlerFunc
    ErrorLoggerT returns a HandlerFunc for a given error type.

func Logger() HandlerFunc
    Logger instances a Logger middleware that will write the logs to
    gin.DefaultWriter. By default, gin.DefaultWriter = os.Stdout.

func LoggerWithConfig(conf LoggerConfig) HandlerFunc
    LoggerWithConfig instance a Logger middleware with config.

func LoggerWithFormatter(f LogFormatter) HandlerFunc
    LoggerWithFormatter instance a Logger middleware with the specified log
    format function.

func LoggerWithWriter(out io.Writer, notlogged ...string) HandlerFunc
    LoggerWithWriter instance a Logger middleware with the specified writer
    buffer. Example: os.Stdout, a file opened in write mode, a socket...

func Recovery() HandlerFunc
    Recovery returns a middleware that recovers from any panics and writes a 500
    if there was one.

func RecoveryWithWriter(out io.Writer, recovery ...RecoveryFunc) HandlerFunc
    RecoveryWithWriter returns a middleware for a given writer that recovers
    from any panics and writes a 500 if there was one.

func WrapF(f http.HandlerFunc) HandlerFunc
    WrapF is a helper function for wrapping http.HandlerFunc and returns a Gin
    middleware.

func WrapH(h http.Handler) HandlerFunc
    WrapH is a helper function for wrapping http.Handler and returns a Gin
    middleware.

type HandlersChain []HandlerFunc
    HandlersChain defines a HandlerFunc slice.

func (c HandlersChain) Last() HandlerFunc
    Last returns the last handler in the chain. i.e. the last handler is the
    main one.

type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}
    IRouter defines all router handle interface includes single and group
    router.

type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	Any(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PATCH(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	OPTIONS(string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes
	Match([]string, string, ...HandlerFunc) IRoutes

	StaticFile(string, string) IRoutes
	StaticFileFS(string, string, http.FileSystem) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, http.FileSystem) IRoutes
}
    IRoutes defines all router handle interface.

type LogFormatter func(params LogFormatterParams) string
    LogFormatter gives the signature of the formatter function passed to
    LoggerWithFormatter

type LogFormatterParams struct {
	Request *http.Request

	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP equals Context's ClientIP method.
	ClientIP string
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
	// ErrorMessage is set if error has occurred in processing the request.
	ErrorMessage string

	// BodySize is the size of the Response Body
	BodySize int
	// Keys are the keys set on the request's context.
	Keys map[string]any
	// Has unexported fields.
}
    LogFormatterParams is the structure any formatter will be handed when time
    to log comes

func (p *LogFormatterParams) IsOutputColor() bool
    IsOutputColor indicates whether can colors be outputted to the log.

func (p *LogFormatterParams) MethodColor() string
    MethodColor is the ANSI color for appropriately logging http method to a
    terminal.

func (p *LogFormatterParams) ResetColor() string
    ResetColor resets all escape attributes.

func (p *LogFormatterParams) StatusCodeColor() string
    StatusCodeColor is the ANSI color for appropriately logging http status code
    to a terminal.

type LoggerConfig struct {
	// Optional. Default value is gin.defaultLogFormatter
	Formatter LogFormatter

	// Output is a writer where logs are written.
	// Optional. Default value is gin.DefaultWriter.
	Output io.Writer

	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string

	// Skip is a Skipper that indicates which logs should not be written.
	// Optional.
	Skip Skipper
}
    LoggerConfig defines the config for Logger middleware.

type Negotiate struct {
	Offered  []string
	HTMLName string
	HTMLData any
	JSONData any
	XMLData  any
	YAMLData any
	Data     any
	TOMLData any
}
    Negotiate contains all negotiations data.

type OptionFunc func(*Engine)
    OptionFunc defines the function to change the default configuration

type Param struct {
	Key   string
	Value string
}
    Param is a single URL parameter, consisting of a key and a value.

type Params []Param
    Params is a Param-slice, as returned by the router. The slice is ordered,
    the first URL parameter is also the first slice value. It is therefore safe
    to read values by the index.

func (ps Params) ByName(name string) (va string)
    ByName returns the value of the first Param which key matches the given
    name. If no matching Param is found, an empty string is returned.

func (ps Params) Get(name string) (string, bool)
    Get returns the value of the first Param which key matches the given name
    and a boolean true. If no matching Param is found, an empty string is
    returned and a boolean false .

type RecoveryFunc func(c *Context, err any)
    RecoveryFunc defines the function passable to CustomRecovery.

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}
    ResponseWriter ...

type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandlerFunc
}
    RouteInfo represents a request route's specification which contains method
    and path and its handler.

type RouterGroup struct {
	Handlers HandlersChain

	// Has unexported fields.
}
    RouterGroup is used internally to configure router, a RouterGroup is
    associated with a prefix and an array of handlers (middleware).

func (group *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) IRoutes
    Any registers a route that matches all the HTTP methods. GET, POST, PUT,
    PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.

func (group *RouterGroup) BasePath() string
    BasePath returns the base path of router group. For example, if v :=
    router.Group("/rest/n/v1/api"), v.BasePath() is "/rest/n/v1/api".

func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes
    DELETE is a shortcut for router.Handle("DELETE", path, handlers).

func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes
    GET is a shortcut for router.Handle("GET", path, handlers).

func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup
    Group creates a new router group. You should add all the routes that have
    common middlewares or the same path prefix. For example, all the routes that
    use a common middleware for authorization could be grouped.

func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes
    HEAD is a shortcut for router.Handle("HEAD", path, handlers).

func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes
    Handle registers a new request handle and middleware with the given path
    and method. The last handler should be the real handler, the other ones
    should be middleware that can and should be shared among different routes.
    See the example code in GitHub.

    For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
    functions can be used.

    This function is intended for bulk loading and to allow the usage of less
    frequently used, non-standardized or custom methods (e.g. for internal
    communication with a proxy).

func (group *RouterGroup) Match(methods []string, relativePath string, handlers ...HandlerFunc) IRoutes
    Match registers a route that matches the specified methods that you
    declared.

func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes
    OPTIONS is a shortcut for router.Handle("OPTIONS", path, handlers).

func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes
    PATCH is a shortcut for router.Handle("PATCH", path, handlers).

func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes
    POST is a shortcut for router.Handle("POST", path, handlers).

func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) IRoutes
    PUT is a shortcut for router.Handle("PUT", path, handlers).

func (group *RouterGroup) Static(relativePath, root string) IRoutes
    Static serves files from the given file system root. Internally a
    http.FileServer is used, therefore http.NotFound is used instead of the
    Router's NotFound handler. To use the operating system's file system
    implementation, use :

        router.Static("/static", "/var/www")

func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes
    StaticFS works just like `Static()` but a custom `http.FileSystem` can be
    used instead. Gin by default uses: gin.Dir()

func (group *RouterGroup) StaticFile(relativePath, filepath string) IRoutes
    StaticFile registers a single route in order to serve a single
    file of the local filesystem. router.StaticFile("favicon.ico",
    "./resources/favicon.ico")

func (group *RouterGroup) StaticFileFS(relativePath, filepath string, fs http.FileSystem) IRoutes
    StaticFileFS works just like `StaticFile` but a custom `http.FileSystem`
    can be used instead.. router.StaticFileFS("favicon.ico",
    "./resources/favicon.ico", Dir{".", false}) Gin by default uses: gin.Dir()

func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes
    Use adds middleware to the group, see example code in GitHub.

type RoutesInfo []RouteInfo
    RoutesInfo defines a RouteInfo slice.

type Skipper func(c *Context) bool
    Skipper is a function to skip logs based on provided Context



======================================================================
 PACKAGE: github.com/gin-gonic/gin/binding
======================================================================

package binding // import "github.com/gin-gonic/gin/binding"


CONSTANTS

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMEYAML              = "application/x-yaml"
	MIMEYAML2             = "application/yaml"
	MIMETOML              = "application/toml"
)
    Content-Type MIME of the most common data formats.


VARIABLES

var (
	JSON          BindingBody = jsonBinding{}
	XML           BindingBody = xmlBinding{}
	Form          Binding     = formBinding{}
	Query         Binding     = queryBinding{}
	FormPost      Binding     = formPostBinding{}
	FormMultipart Binding     = formMultipartBinding{}
	ProtoBuf      BindingBody = protobufBinding{}
	MsgPack       BindingBody = msgpackBinding{}
	YAML          BindingBody = yamlBinding{}
	Uri           BindingUri  = uriBinding{}
	Header        Binding     = headerBinding{}
	TOML          BindingBody = tomlBinding{}
)
    These implement the Binding interface and can be used to bind the data
    present in the request to struct instances.

var (

	// ErrConvertMapStringSlice can not convert to map[string][]string
	ErrConvertMapStringSlice = errors.New("can not convert to map slices of strings")

	// ErrConvertToMapString can not convert to map[string]string
	ErrConvertToMapString = errors.New("can not convert to map of strings")
)
var (
	// ErrMultiFileHeader multipart.FileHeader invalid
	ErrMultiFileHeader = errors.New("unsupported field type for multipart.FileHeader")

	// ErrMultiFileHeaderLenInvalid array for []*multipart.FileHeader len invalid
	ErrMultiFileHeaderLenInvalid = errors.New("unsupported len of array for []*multipart.FileHeader")
)
var EnableDecoderDisallowUnknownFields = false
    EnableDecoderDisallowUnknownFields is used to call the DisallowUnknownFields
    method on the JSON Decoder instance. DisallowUnknownFields causes the
    Decoder to return an error when the destination is a struct and the input
    contains object keys which do not match any non-ignored, exported fields in
    the destination.

var EnableDecoderUseNumber = false
    EnableDecoderUseNumber is used to call the UseNumber method on the JSON
    Decoder instance. UseNumber causes the Decoder to unmarshal a number into an
    any as a Number instead of as a float64.


FUNCTIONS

func MapFormWithTag(ptr any, form map[string][]string, tag string) error

TYPES

type BindUnmarshaler interface {
	// UnmarshalParam decodes and assigns a value from an form or query param.
	UnmarshalParam(param string) error
}
    BindUnmarshaler is the interface used to wrap the UnmarshalParam method.

type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}
    Binding describes the interface which needs to be implemented for binding
    the data present in the request such as JSON request body, query parameters
    or the form POST.

func Default(method, contentType string) Binding
    Default returns the appropriate Binding instance based on the HTTP method
    and the content type.

type BindingBody interface {
	Binding
	BindBody([]byte, any) error
}
    BindingBody adds BindBody method to Binding. BindBody is similar with Bind,
    but it reads the body from supplied bytes instead of req.Body.

type BindingUri interface {
	Name() string
	BindUri(map[string][]string, any) error
}
    BindingUri adds BindUri method to Binding. BindUri is similar with Bind,
    but it reads the Params.

type SliceValidationError []error

func (err SliceValidationError) Error() string
    Error concatenates all error elements in SliceValidationError into a single
    string separated by \n.

type StructValidator interface {
	// ValidateStruct can receive any kind of type and it should never panic, even if the configuration is not right.
	// If the received type is a slice|array, the validation should be performed travel on every element.
	// If the received type is not a struct or slice|array, any validation should be skipped and nil must be returned.
	// If the received type is a struct or pointer to a struct, the validation should be performed.
	// If the struct is not valid or the validation itself fails, a descriptive error should be returned.
	// Otherwise nil must be returned.
	ValidateStruct(any) error

	// Engine returns the underlying validator engine which powers the
	// StructValidator implementation.
	Engine() any
}
    StructValidator is the minimal interface which needs to be implemented in
    order for it to be used as the validator engine for ensuring the correctness
    of the request. Gin provides a default implementation for this using
    https://github.com/go-playground/validator/tree/v10.6.1.

var Validator StructValidator = &defaultValidator{}
    Validator is the default validator which implements the StructValidator
    interface. It uses https://github.com/go-playground/validator/tree/v10.6.1
    under the hood.



======================================================================
 PACKAGE: github.com/gin-gonic/gin/ginS
======================================================================

package ginS // import "github.com/gin-gonic/gin/ginS"


FUNCTIONS

func Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    Any is a wrapper for Engine.Any.

func DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    DELETE is a shortcut for router.Handle("DELETE", path, handle)

func GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    GET is a shortcut for router.Handle("GET", path, handle)

func Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup
    Group creates a new router group. You should add all the routes that have
    common middlewares or the same path prefix. For example, all the routes that
    use a common middleware for authorization could be grouped.

func HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    HEAD is a shortcut for router.Handle("HEAD", path, handle)

func Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    Handle is a wrapper for Engine.Handle.

func LoadHTMLFiles(files ...string)
    LoadHTMLFiles is a wrapper for Engine.LoadHTMLFiles.

func LoadHTMLGlob(pattern string)
    LoadHTMLGlob is a wrapper for Engine.LoadHTMLGlob.

func NoMethod(handlers ...gin.HandlerFunc)
    NoMethod is a wrapper for Engine.NoMethod.

func NoRoute(handlers ...gin.HandlerFunc)
    NoRoute adds handlers for NoRoute. It returns a 404 code by default.

func OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)

func PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    PATCH is a shortcut for router.Handle("PATCH", path, handle)

func POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    POST is a shortcut for router.Handle("POST", path, handle)

func PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes
    PUT is a shortcut for router.Handle("PUT", path, handle)

func Routes() gin.RoutesInfo
    Routes returns a slice of registered routes.

func Run(addr ...string) (err error)
    Run attaches to a http.Server and starts listening and serving HTTP
    requests. It is a shortcut for http.ListenAndServe(addr, router) Note:
    this method will block the calling goroutine indefinitely unless an error
    happens.

func RunFd(fd int) (err error)
    RunFd attaches the router to a http.Server and starts listening and serving
    HTTP requests through the specified file descriptor. Note: the method will
    block the calling goroutine indefinitely unless on error happens.

func RunTLS(addr, certFile, keyFile string) (err error)
    RunTLS attaches to a http.Server and starts listening and serving HTTPS
    requests. It is a shortcut for http.ListenAndServeTLS(addr, certFile,
    keyFile, router) Note: this method will block the calling goroutine
    indefinitely unless an error happens.

func RunUnix(file string) (err error)
    RunUnix attaches to a http.Server and starts listening and serving HTTP
    requests through the specified unix socket (i.e. a file) Note: this method
    will block the calling goroutine indefinitely unless an error happens.

func SetHTMLTemplate(templ *template.Template)
    SetHTMLTemplate is a wrapper for Engine.SetHTMLTemplate.

func Static(relativePath, root string) gin.IRoutes
    Static serves files from the given file system root. Internally a
    http.FileServer is used, therefore http.NotFound is used instead of the
    Router's NotFound handler. To use the operating system's file system
    implementation, use :

        router.Static("/static", "/var/www")

func StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes
    StaticFS is a wrapper for Engine.StaticFS.

func StaticFile(relativePath, filepath string) gin.IRoutes
    StaticFile is a wrapper for Engine.StaticFile.

func Use(middlewares ...gin.HandlerFunc) gin.IRoutes
    Use attaches a global middleware to the router. i.e. the middlewares
    attached through Use() will be included in the handlers chain for every
    single request. Even 404, 405, static files... For example, this is the
    right place for a logger or error management middleware.



======================================================================
 PACKAGE: github.com/gin-gonic/gin/internal/bytesconv
======================================================================

package bytesconv // import "github.com/gin-gonic/gin/internal/bytesconv"


FUNCTIONS

func BytesToString(b []byte) string
    BytesToString converts byte slice to string
    without a memory allocation. For more details, see
    https://github.com/golang/go/issues/53003#issuecomment-1140276077.

func StringToBytes(s string) []byte
    StringToBytes converts string to byte slice
    without a memory allocation. For more details, see
    https://github.com/golang/go/issues/53003#issuecomment-1140276077.



======================================================================
 PACKAGE: github.com/gin-gonic/gin/internal/json
======================================================================

package json // import "github.com/gin-gonic/gin/internal/json"


VARIABLES

var (
	// Marshal is exported by gin/json package.
	Marshal = json.Marshal
	// Unmarshal is exported by gin/json package.
	Unmarshal = json.Unmarshal
	// MarshalIndent is exported by gin/json package.
	MarshalIndent = json.MarshalIndent
	// NewDecoder is exported by gin/json package.
	NewDecoder = json.NewDecoder
	// NewEncoder is exported by gin/json package.
	NewEncoder = json.NewEncoder
)


======================================================================
 PACKAGE: github.com/gin-gonic/gin/render
======================================================================

package render // import "github.com/gin-gonic/gin/render"


VARIABLES

var TOMLContentType = []string{"application/toml; charset=utf-8"}

FUNCTIONS

func WriteJSON(w http.ResponseWriter, obj any) error
    WriteJSON marshals the given interface object and writes it with custom
    ContentType.

func WriteMsgPack(w http.ResponseWriter, obj any) error
    WriteMsgPack writes MsgPack ContentType and encodes the given interface
    object.

func WriteString(w http.ResponseWriter, format string, data []any) (err error)
    WriteString writes data according to its format and write custom
    ContentType.


TYPES

type AsciiJSON struct {
	Data any
}
    AsciiJSON contains the given interface object.

func (r AsciiJSON) Render(w http.ResponseWriter) (err error)
    Render (AsciiJSON) marshals the given interface object and writes it with
    custom ContentType.

func (r AsciiJSON) WriteContentType(w http.ResponseWriter)
    WriteContentType (AsciiJSON) writes JSON ContentType.

type Data struct {
	ContentType string
	Data        []byte
}
    Data contains ContentType and bytes data.

func (r Data) Render(w http.ResponseWriter) (err error)
    Render (Data) writes data with custom ContentType.

func (r Data) WriteContentType(w http.ResponseWriter)
    WriteContentType (Data) writes custom ContentType.

type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}
    Delims represents a set of Left and Right delimiters for HTML template
    rendering.

type HTML struct {
	Template *template.Template
	Name     string
	Data     any
}
    HTML contains template reference and its name with given interface object.

func (r HTML) Render(w http.ResponseWriter) error
    Render (HTML) executes template and writes its result with custom
    ContentType for response.

func (r HTML) WriteContentType(w http.ResponseWriter)
    WriteContentType (HTML) writes HTML ContentType.

type HTMLDebug struct {
	Files   []string
	Glob    string
	Delims  Delims
	FuncMap template.FuncMap
}
    HTMLDebug contains template delims and pattern and function with file list.

func (r HTMLDebug) Instance(name string, data any) Render
    Instance (HTMLDebug) returns an HTML instance which it realizes Render
    interface.

type HTMLProduction struct {
	Template *template.Template
	Delims   Delims
}
    HTMLProduction contains template reference and its delims.

func (r HTMLProduction) Instance(name string, data any) Render
    Instance (HTMLProduction) returns an HTML instance which it realizes Render
    interface.

type HTMLRender interface {
	// Instance returns an HTML instance.
	Instance(string, any) Render
}
    HTMLRender interface is to be implemented by HTMLProduction and HTMLDebug.

type IndentedJSON struct {
	Data any
}
    IndentedJSON contains the given interface object.

func (r IndentedJSON) Render(w http.ResponseWriter) error
    Render (IndentedJSON) marshals the given interface object and writes it with
    custom ContentType.

func (r IndentedJSON) WriteContentType(w http.ResponseWriter)
    WriteContentType (IndentedJSON) writes JSON ContentType.

type JSON struct {
	Data any
}
    JSON contains the given interface object.

func (r JSON) Render(w http.ResponseWriter) error
    Render (JSON) writes data with custom ContentType.

func (r JSON) WriteContentType(w http.ResponseWriter)
    WriteContentType (JSON) writes JSON ContentType.

type JsonpJSON struct {
	Callback string
	Data     any
}
    JsonpJSON contains the given interface object its callback.

func (r JsonpJSON) Render(w http.ResponseWriter) (err error)
    Render (JsonpJSON) marshals the given interface object and writes it and its
    callback with custom ContentType.

func (r JsonpJSON) WriteContentType(w http.ResponseWriter)
    WriteContentType (JsonpJSON) writes Javascript ContentType.

type MsgPack struct {
	Data any
}
    MsgPack contains the given interface object.

func (r MsgPack) Render(w http.ResponseWriter) error
    Render (MsgPack) encodes the given interface object and writes data with
    custom ContentType.

func (r MsgPack) WriteContentType(w http.ResponseWriter)
    WriteContentType (MsgPack) writes MsgPack ContentType.

type ProtoBuf struct {
	Data any
}
    ProtoBuf contains the given interface object.

func (r ProtoBuf) Render(w http.ResponseWriter) error
    Render (ProtoBuf) marshals the given interface object and writes data with
    custom ContentType.

func (r ProtoBuf) WriteContentType(w http.ResponseWriter)
    WriteContentType (ProtoBuf) writes ProtoBuf ContentType.

type PureJSON struct {
	Data any
}
    PureJSON contains the given interface object.

func (r PureJSON) Render(w http.ResponseWriter) error
    Render (PureJSON) writes custom ContentType and encodes the given interface
    object.

func (r PureJSON) WriteContentType(w http.ResponseWriter)
    WriteContentType (PureJSON) writes custom ContentType.

type Reader struct {
	ContentType   string
	ContentLength int64
	Reader        io.Reader
	Headers       map[string]string
}
    Reader contains the IO reader and its length, and custom ContentType and
    other headers.

func (r Reader) Render(w http.ResponseWriter) (err error)
    Render (Reader) writes data with custom ContentType and headers.

func (r Reader) WriteContentType(w http.ResponseWriter)
    WriteContentType (Reader) writes custom ContentType.

type Redirect struct {
	Code     int
	Request  *http.Request
	Location string
}
    Redirect contains the http request reference and redirects status code and
    location.

func (r Redirect) Render(w http.ResponseWriter) error
    Render (Redirect) redirects the http request to new location and writes
    redirect response.

func (r Redirect) WriteContentType(http.ResponseWriter)
    WriteContentType (Redirect) don't write any ContentType.

type Render interface {
	// Render writes data with custom ContentType.
	Render(http.ResponseWriter) error
	// WriteContentType writes custom ContentType.
	WriteContentType(w http.ResponseWriter)
}
    Render interface is to be implemented by JSON, XML, HTML, YAML and so on.

type SecureJSON struct {
	Prefix string
	Data   any
}
    SecureJSON contains the given interface object and its prefix.

func (r SecureJSON) Render(w http.ResponseWriter) error
    Render (SecureJSON) marshals the given interface object and writes it with
    custom ContentType.

func (r SecureJSON) WriteContentType(w http.ResponseWriter)
    WriteContentType (SecureJSON) writes JSON ContentType.

type String struct {
	Format string
	Data   []any
}
    String contains the given interface object slice and its format.

func (r String) Render(w http.ResponseWriter) error
    Render (String) writes data with custom ContentType.

func (r String) WriteContentType(w http.ResponseWriter)
    WriteContentType (String) writes Plain ContentType.

type TOML struct {
	Data any
}
    TOML contains the given interface object.

func (r TOML) Render(w http.ResponseWriter) error
    Render (TOML) marshals the given interface object and writes data with
    custom ContentType.

func (r TOML) WriteContentType(w http.ResponseWriter)
    WriteContentType (TOML) writes TOML ContentType for response.

type XML struct {
	Data any
}
    XML contains the given interface object.

func (r XML) Render(w http.ResponseWriter) error
    Render (XML) encodes the given interface object and writes data with custom
    ContentType.

func (r XML) WriteContentType(w http.ResponseWriter)
    WriteContentType (XML) writes XML ContentType for response.

type YAML struct {
	Data any
}
    YAML contains the given interface object.

func (r YAML) Render(w http.ResponseWriter) error
    Render (YAML) marshals the given interface object and writes data with
    custom ContentType.

func (r YAML) WriteContentType(w http.ResponseWriter)
    WriteContentType (YAML) writes YAML ContentType for response.



