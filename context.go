package q

import (
	"encoding/json"
	qt "github.com/wonderful-summer/q-type"
	"net/http"
)

// Context 框架的核心结构体
// 用于挂载完整的请求和响应，同时会挂载当次请求的各种基础信息和响应的基础信息
type Context struct {
	writer     http.ResponseWriter
	req        *http.Request
	path       string
	method     string
	params     map[string]string
	StatusCode int

	// 指向整个框架，方便随时获取框架层面全局的一些配置
	engine DefaultEngine

	// 本次请求的所有中间件以及handler处理函数的数组
	handlers []qt.HandlerFunc
	// 用于计算执行到那个中间件
	index int
}

func newContext(w http.ResponseWriter, req *http.Request) qt.DefaultContext {
	return &Context{
		writer: w,
		req:    req,
		path:   req.URL.Path,
		method: req.Method,
		index:  -1,
	}
}

func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}
func (c *Context) Req() *http.Request {
	return c.req
}
func (c *Context) Method() string {
	return c.method
}
func (c *Context) Path() string {
	return c.path
}

// Get engin里面的Get函数
func (c *Context) Get(key string) any {
	return c.engine.Get(key)
}

func (c *Context) SetEngin(engin qt.DefaultEngine) {
	c.engine = engin
}
func (c *Context) SetParams(params map[string]string) {
	c.params = params
}
func (c *Context) SetHandler(handlers qt.HandlerFunc) {
	c.handlers = append(c.handlers, handlers)
}
func (c *Context) SetHandlers(handlers []qt.HandlerFunc) {
	c.handlers = append(c.handlers, handlers...)
}

func (c *Context) Next() {
	c.index++
	// tips: 如果直接 s := len(c.handlers)会只执行第一个全局的中间件
	// 这里循环处理，是为了所有的中间件和处理函数都可以被调用
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Body(key string) string {
	return c.req.FormValue(key)
}
func (c *Context) Query(key string) string {
	return c.req.URL.Query().Get(key)
}
func (c *Context) Param(key string) string {
	value, _ := c.params[key]
	return value
}
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Req().Cookie(name)
}
func (c *Context) SetCookie(name string, value string, maxAge int, path string, domain string, secure bool, httpOnly bool) {
	http.SetCookie(c.Writer(), &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Status(code int) qt.DefaultContext {
	c.StatusCode = code
	c.writer.WriteHeader(code)
	return c
}

func (c *Context) SetHeader(key string, value string) qt.DefaultContext {
	c.writer.Header().Set(key, value)
	return c
}

func (c *Context) Header(key string) string {
	return c.writer.Header().Get(key)
}

func (c *Context) End(text string) {
	ct := c.Header("Content-Type")
	if len(ct) == 0 {
		c.SetHeader("Content-Type", "text/html")
	}
	c.writer.Write([]byte(text))
}

func (c *Context) Json(obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	encoder := json.NewEncoder(c.writer)
	if err := encoder.Encode(obj); err != nil {
		c.Status(500).End(err.Error())
	}
}
func (c *Context) Fail(err string) {
	c.index = len(c.handlers)
	c.Json(qt.H{"message": err})
}

func (c *Context) Render(name string, data map[string]interface{}) {
	c.SetHeader("Content-Type", "text/html")
	str, err := c.engine.GetView().Render(name, data)
	if err != nil {
		c.Status(500).Fail(err.Error())
		return
	}
	c.End(str)
}
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.writer, c.req, location, code)
}
