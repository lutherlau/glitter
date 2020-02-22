package glitter

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSON map[string]interface{}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	Path   string
	Method string
	Params map[string]string

	StatusCode int

	handlers []HandlerFunc
	// index of middleware
	index  int
	engine *Engine
}

// switch to next handler
func (ctx *Context) Next() {
	ctx.index++
	for s := len(ctx.handlers); ctx.index < s; ctx.index++ {
		ctx.handlers[ctx.index](ctx)
	}
}

func (ctx *Context) Param(key string) string {
	value, _ := ctx.Params[key]
	return value
}

// reset for reuse
func (ctx *Context) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.Writer = w
	ctx.Request = r
	ctx.Path = ctx.Request.URL.Path
	ctx.Method = ctx.Request.Method
	ctx.index = -1
	ctx.Params = make(map[string]string)
	ctx.StatusCode = 200
	ctx.handlers = ctx.handlers[0:0]
}

func (ctx *Context) PostForm(key string) string {
	return ctx.Request.FormValue(key)
}

func (ctx *Context) Query(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

func (ctx *Context) Status(code int) {
	ctx.StatusCode = code
	ctx.Writer.WriteHeader(code)
}

func (ctx *Context) SetHeader(key string, value string) {
	ctx.Writer.Header().Set(key, value)
}

func (ctx *Context) String(code int, format string, values ...interface{}) {
	ctx.SetHeader("Content-Type", "text/plain")
	ctx.Status(code)
	ctx.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (ctx *Context) JSON(code int, obj interface{}) {
	ctx.SetHeader("Content-Type", "application/json")
	jsont, err := json.MarshalIndent(&obj, "", "\t\t")
	if err != nil {
		http.Error(ctx.Writer, err.Error(), 500)
	} else {
		ctx.Status(code)
		ctx.Writer.Write(jsont)
	}
}

func (ctx *Context) Data(code int, data []byte) {
	ctx.Status(code)
	ctx.Writer.Write(data)
}

func (ctx *Context) Fail(code int, err string) {
	ctx.index = len(ctx.handlers)
	ctx.JSON(code, JSON{"message": err})
}

func (ctx *Context) HTML(code int, name string, data interface{}) {
	ctx.SetHeader("Content-Type", "text/html")
	ctx.Status(code)
	if err := ctx.engine.htmlTemplates.ExecuteTemplate(ctx.Writer, name, data); err != nil {
		ctx.Fail(500, err.Error())
	}
}
