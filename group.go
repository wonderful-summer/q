package q

import (
	qt "github.com/wonderful-summer/q-type"
	"net/http"
	"path"
)

type RouterGroup struct {
	prefix      string
	middlewares []qt.HandlerFunc

	// 用于记录上级分组
	parent *RouterGroup
	engine *Engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...qt.HandlerFunc) {
	middle := group.middlewares
	group.middlewares = append(middle, middlewares...)
}

// feat: 新增功能，递归获取上级的路由前缀
func (group *RouterGroup) getPrefix() string {
	prefix := group.prefix
	parent := group.parent
	if parent != nil && parent.prefix != "" {
		return parent.getPrefix() + prefix
	}
	return prefix
}
func (group *RouterGroup) addRoute(method string, comp string, handler qt.HandlerFunc) {
	// 解决不能嵌套group
	pattern := group.getPrefix() + comp

	group.engine.router.Add(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler qt.HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}
func (group *RouterGroup) POST(pattern string, handler qt.HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}
func (group *RouterGroup) PUT(pattern string, handler qt.HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}
func (group *RouterGroup) DELETE(pattern string, handler qt.HandlerFunc) {
	group.addRoute("DELETE", pattern, handler)
}
func (group *RouterGroup) OPTION(pattern string, handler qt.HandlerFunc) {
	group.addRoute("OPTION", pattern, handler)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) qt.HandlerFunc {
	abslutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(abslutePath, http.FileServer(fs))
	return func(c qt.DefaultContext) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer(), c.Req())
	}
}

// Static 配置静态服务路由
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	// todo: 这里有个bug，assets/index.html 访问是404，就算有文件，但是其他名称没问题
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}
