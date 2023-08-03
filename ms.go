package mango

import (
	"fmt"
	"log"
	"net/http"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type routerGroup struct {
	name             string                // 组名
	handleFunMap     map[string]HandleFunc // 路由路径pattern->方法handler
	handlerMethodMap map[string][]string   // key为请求方式Get、post、put等，value为该请求方式下的路由pattern集合
}

func (r *routerGroup) Any(pattern string, handleFunc HandleFunc) {
	r.handleFunMap[pattern] = handleFunc
	r.handlerMethodMap["ANY"] = append(r.handlerMethodMap["ANY"], pattern)
}

func (r *routerGroup) Get(pattern string, handleFunc HandleFunc) {
	r.handleFunMap[pattern] = handleFunc
	r.handlerMethodMap[http.MethodGet] = append(r.handlerMethodMap[http.MethodGet], pattern)
}

func (r *routerGroup) Post(pattern string, handleFunc HandleFunc) {
	r.handleFunMap[pattern] = handleFunc
	r.handlerMethodMap[http.MethodPost] = append(r.handlerMethodMap[http.MethodPost], pattern)
}

func (r *routerGroup) Delete(pattern string, handleFunc HandleFunc) {
	r.handleFunMap[pattern] = handleFunc
	r.handlerMethodMap[http.MethodDelete] = append(r.handlerMethodMap[http.MethodDelete], pattern)
}

func (r *routerGroup) Put(pattern string, handleFunc HandleFunc) {
	r.handleFunMap[pattern] = handleFunc
	r.handlerMethodMap[http.MethodPut] = append(r.handlerMethodMap[http.MethodPut], pattern)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	rg := &routerGroup{
		name:             name,
		handleFunMap:     make(map[string]HandleFunc),
		handlerMethodMap: make(map[string][]string),
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

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	fmt.Println(method)
}

func (e *Engine) Run() {
	//for _, group := range e.routerGroups {
	//	for pattern, handler := range group.handleFunMap {
	//		http.HandleFunc("/"+group.name+pattern, handler)
	//	}
	//}
	err := http.ListenAndServe(":8989", e)
	if err != nil {
		log.Fatal(err)
	}
}
