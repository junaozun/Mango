package engine

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	W          http.ResponseWriter
	R          *http.Request
	Path       string
	Params     map[string]string
	Method     string
	StatusCode int
	handlers   []HandleFunc // middleware
	index      int          // 中间件的索引
}

func NewContext() *Context {
	return &Context{
		Params: make(map[string]string),
		index:  -1,
	}
}

func (c *Context) flush() {
	c.handlers = nil
	c.index = -1
}

func (c *Context) Next() {
	c.index++
	c.handlers[c.index](c)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.W.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.W)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.W, err.Error(), 500)
	}
}

func (c *Context) Status(code int) {
	c.W.WriteHeader(code)
	c.StatusCode = code
}

func (c *Context) SetHeader(key string, value string) {
	c.W.Header().Set(key, value)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.W.Write([]byte(html))
}

func (c *Context) Query(key string) string {
	return c.R.URL.Query().Get(key)
}

func (c *Context) PostForm(key string) string {
	return c.R.FormValue(key)
}

func (c *Context) HtmlTemplateGlob(name string, data any, pattern string) error {
	c.SetHeader("Content-Type", "text/html")
	t := template.New(name)
	e, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	return e.Execute(c.W, data)
}

func (c *Context) File(filename string) {
	http.ServeFile(c.W, c.R, filename)
}

func (c *Context) Redirect(status int, url string) {
	http.Redirect(c.W, c.R, url, status)
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
