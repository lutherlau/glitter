package glitter

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"sync"
	"time"
)

type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix string
	parent *RouterGroup
	engine *Engine
}

type Engine struct {
	*RouterGroup
	router        *router
	htmlTemplates *template.Template
	funcMap       template.FuncMap
	pool          sync.Pool
}

// middleware for log
func Logger() HandlerFunc {
	return func(ctx *Context) {
		t := time.Now()
		ctx.Next()
		log.Printf("[%d] %s %s in %v", ctx.StatusCode, ctx.Method, ctx.Request.RequestURI, time.Since(t))
	}
}

func (engine *Engine) newContext() *Context {
	return &Context{
		engine: engine,
	}
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.pool.New = func() interface{} {
		return engine.newContext()
	}
	return engine
}

// default use logger and recover
func DefaultEngine() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(ctx *Context) {
		file := ctx.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	return newGroup
}

func (group *RouterGroup) addRouter(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Add Route %8s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRouter("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRouter("POST", pattern, handler)
}
func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRouter("PUT", pattern, handler)
}
func (group *RouterGroup) PATCH(pattern string, handler HandlerFunc) {
	group.addRouter("PATCH", pattern, handler)
}
func (group *RouterGroup) HEAD(pattern string, handler HandlerFunc) {
	group.addRouter("HEAD", pattern, handler)
}
func (group *RouterGroup) OPTIONS(pattern string, handler HandlerFunc) {
	group.addRouter("OPTIONS", pattern, handler)
}
func (group *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	group.addRouter("DELETE", pattern, handler)
}

func (group *RouterGroup) CONNECT(pattern string, handler HandlerFunc) {
	group.addRouter("CONNECT", pattern, handler)
}

func (group *RouterGroup) TRACE(pattern string, handler HandlerFunc) {
	group.addRouter("TRACE", pattern, handler)
}

func (group *RouterGroup) Any(pattern string, handler HandlerFunc) {
	group.addRouter("GET", pattern, handler)
	group.addRouter("POST", pattern, handler)
	group.addRouter("PUT", pattern, handler)
	group.addRouter("PATCH", pattern, handler)
	group.addRouter("HEAD", pattern, handler)
	group.addRouter("TRACE", pattern, handler)
	group.addRouter("CONNECT", pattern, handler)
	group.addRouter("OPTIONS", pattern, handler)
	group.addRouter("DELETE", pattern, handler)
}

func (group *RouterGroup) Use(middleWares ...HandlerFunc) {
	var pattern string
	if group.prefix == "" {
		pattern = "engine"
	} else {
		pattern = group.prefix
	}
	log.Printf("Use %d middleWars on %s", len(middleWares), pattern)
	group.engine.router.useMiddleWars(pattern, middleWares...)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := engine.pool.Get().(*Context)
	ctx.Reset(w, r)
	engine.router.handle(ctx)
	engine.pool.Put(ctx)
}
