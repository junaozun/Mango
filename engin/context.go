package engin

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

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		W:      w,
		R:      r,
		Params: make(map[string]string),
		Method: r.Method,
		Path:   r.URL.Path,
		index:  -1,
	}
}

func (c *Context) Next() {
	c.index++
	for i := c.index; i < len(c.handlers); i++ {
		c.handlers[i](c)
	}
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
	c.StatusCode = code
	c.W.WriteHeader(code)
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
