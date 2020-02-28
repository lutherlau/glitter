package glitter

import (
	"strings"
)

type HandlerChain []HandlerFunc

type node struct {
	pattern   string // fullpath
	part      string
	children  []*node
	handlers  HandlerChain           // for router group
	handler   map[string]HandlerFunc // handler for method
	hasMethod map[string]bool        //can handle this method?
	isWild    bool
}

// find first child
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// find children
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// insert node, if method is "" , mean than handlers for group
func (n *node) insert(method string, pattern string, parts []string, height int, handlers ...HandlerFunc) {
	if method != "" {
		n.hasMethod[method] = true
	}
	if len(parts) == height {
		// insert handler to node
		if method != "" {
			n.pattern = pattern
			n.handler[method] = handlers[0]
		} else {
			// insert handlers to group
			n.handlers = append(n.handlers, handlers...)
		}
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:      part,
			isWild:    part[0] == ':' || part[0] == '*',
			handlers:  make(HandlerChain, 0),
			handler:   make(map[string]HandlerFunc),
			hasMethod: make(map[string]bool),
		}
		n.children = append(n.children, child)
	}
	child.insert(method, pattern, parts, height+1, handlers...)
}

// search node
func (n *node) search(method string, parts []string, height int) (*node, HandlerChain) {
	if _, exists := n.hasMethod[method]; !exists {
		return nil, nil
	}

	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// this node is not endpoint
		if n.pattern == "" {
			return nil, nil
		}
		return n, n.handlers
	}
	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		result, handlers := child.search(method, parts, height+1)
		if result != nil {
			return result, append(n.handlers, handlers...)
		}
	}
	return nil, nil
}
