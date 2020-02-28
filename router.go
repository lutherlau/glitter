package glitter

import (
	"net/http"
	"strings"
)

type router struct {
	// roots map[string]*node
	root *node
	// handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		root: &node{
			handlers:  make(HandlerChain, 0),
			handler:   make(map[string]HandlerFunc),
			hasMethod: make(map[string]bool),
		},
		// roots:    make(map[string]*node),
		// handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) useMiddleWars(pattern string, handlers ...HandlerFunc) {
	parts := parsePattern(pattern)
	r.root.insert("", pattern, parts, 0, handlers...)
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	r.root.insert(method, pattern, parts, 0, handler)

}

func (r *router) getRoute(method string, path string) (*node, HandlerChain, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	// TODO : when find node, get params directly, moit this scan step
	n, handlers := r.root.search(method, searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			} else {
				if part[0] == '*' && len(part) > 1 {
					params[part[1:]] = strings.Join(searchParts[index:], "/")
				}
			}
		}
		return n, handlers, params
	}
	return nil, nil, nil
}

func (r *router) handle(ctx *Context) {
	n, handlers, params := r.getRoute(ctx.Method, ctx.Path)
	if n != nil {
		ctx.Params = params
		ctx.handlers = append(handlers, n.handler[ctx.Method])
	} else {
		ctx.handlers = append(ctx.handlers, func(ctx *Context) {
			ctx.String(http.StatusNotFound, "404 NOT FOUND: %s\n", ctx.Path)
		})
	}
	// handlers start work
	ctx.Next()
}
