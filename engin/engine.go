package engin

import (
	"log"
	"net/http"
	"strings"
)

type HandleFunc func(ctx *Context)

type Engine struct {
	router       *router
	*RouterGroup                // 当前所处于的组
	groups       []*RouterGroup // 所有的组
}

func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = append(engine.groups, engine.RouterGroup)
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Recovery())
	return engine
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var middlewares []HandleFunc
	// 将所有当前url的组的中间件找出来，并赋值给ctx中的中间件
	for _, group := range e.groups {
		if strings.HasPrefix(r.RequestURI, group.name) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	ctx := NewContext(w, r)
	ctx.handlers = middlewares
	e.router.handle(ctx)
}

func (e *Engine) Run(addr string) error {
	log.Printf("服务已启动，监听地址%s", addr)
	return http.ListenAndServe(addr, e)
}
