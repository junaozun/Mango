package engine

import (
	"net/http"
)

type RouterGroup struct {
	name        string       // 组名
	middlewares []HandleFunc // 中间件被定义为组的属性
	engine      *Engine
}

func (rg *RouterGroup) Group(name string) *RouterGroup {
	engine := rg.engine
	newGroup := &RouterGroup{
		name:   rg.name + name,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (rg *RouterGroup) Use(middlewares ...HandleFunc) {
	rg.middlewares = append(rg.middlewares, middlewares...)
}

func (rg *RouterGroup) addRoute(method string, part string, handler HandleFunc) {
	pattern := rg.name + part
	rg.engine.router.addRouter(method, pattern, handler)
}

func (r *RouterGroup) ANY(pattern string, handleFunc HandleFunc) {
	r.addRoute(http.MethodGet, pattern, handleFunc)
	r.addRoute(http.MethodPost, pattern, handleFunc)
	r.addRoute(http.MethodDelete, pattern, handleFunc)
	r.addRoute(http.MethodPut, pattern, handleFunc)
}

func (r *RouterGroup) GET(pattern string, handleFunc HandleFunc) {
	r.addRoute(http.MethodGet, pattern, handleFunc)
}

func (r *RouterGroup) POST(pattern string, handleFunc HandleFunc) {
	r.addRoute(http.MethodPost, pattern, handleFunc)
}

func (r *RouterGroup) DELETE(pattern string, handleFunc HandleFunc) {
	r.addRoute(http.MethodDelete, pattern, handleFunc)
}

func (r *RouterGroup) PUT(pattern string, handleFunc HandleFunc) {
	r.addRoute(http.MethodPut, pattern, handleFunc)
}
