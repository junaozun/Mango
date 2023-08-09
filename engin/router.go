package engin

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*treeNode  // 请求方式(get/post/put/delete) -> 前缀树
	handlers map[string]HandleFunc // key: method+"-"+pattern; value: 回调方法handler
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*treeNode),
		handlers: make(map[string]HandleFunc),
	}
}

func (r *router) handle(c *Context) {
	node, params := r.getRouter(c.R.Method, c.R.RequestURI)
	if node == nil {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND:%s\n", c.Path)
		})
	} else {
		c.Params = params
		key := c.R.Method + "-" + node.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	}
	// handler真正执行在next函数中
	c.Next()
}

func parsePattern(pattern string) []string {
	patternStrs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, part := range patternStrs {
		if part == "" {
			continue
		}
		parts = append(parts, part)
		// 当匹配到*不再往下匹配
		if part[0] == '*' {
			break
		}
	}
	return parts
}

func (r *router) addRouter(method string, pattern string, handler HandleFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &treeNode{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRouter(method, pattern string) (*treeNode, map[string]string) {
	parts := parsePattern(pattern)
	params := make(map[string]string)
	rootNode, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	trieNode := rootNode.search(parts, 0)
	if trieNode == nil {
		return nil, nil
	}
	trieNodeParts := parsePattern(trieNode.pattern)
	for index, part := range trieNodeParts {
		if part[0] == ':' {
			params[part[1:]] = parts[index]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(parts[index:], "/")
			break
		}
	}
	return trieNode, params
}
