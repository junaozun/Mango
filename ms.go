package mango

import (
	"log"
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type routerGroup struct {
	name         string                // 组名
	handleFunMap map[string]HandleFunc // 路由路径pattern->方法handler
}

func (r *routerGroup) Add(pattern string, handleFunc HandleFunc) {
	r.handleFunMap[pattern] = handleFunc
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	rg := &routerGroup{
		name:         name,
		handleFunMap: make(map[string]HandleFunc),
	}
	r.routerGroups = append(r.routerGroups, rg)
	return rg
}

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{
		router: router{},
	}
}

func (e *Engine) Run() {
	for _, group := range e.routerGroups {
		for pattern, handler := range group.handleFunMap {
			http.HandleFunc("/"+group.name+pattern, handler)
		}
	}
	err := http.ListenAndServe(":8989", nil)
	if err != nil {
		log.Fatal(err)
	}
}
