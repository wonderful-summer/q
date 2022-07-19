package q

import (
	"fmt"
	qt "github.com/Wonderful-Summer/q-type"
	"log"
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]qt.HandlerFunc
}

func NewRouter() qt.DefaultRouter {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]qt.HandlerFunc),
	}
}

// 只允许路由中有一个*
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			// todo: 这里的item可能是中文，需要特殊处理
			// 配置路由的时候是英文，但是页面传递的是中文
			parts = append(parts, item)
			// *是通配符，如果匹配到* 后面的路由没有意义直接按照通配符进行处理
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) Add(method string, pattern string, handler qt.HandlerFunc) {
	log.Printf("内部路由信息: %4s - %s", method, pattern)
	parts := parsePattern(pattern)
	key := method + "_" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	// 使用http 方法进行归类
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)

	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	if n == nil {
		return nil, nil
	}

	parts := parsePattern(n.pattern)
	for index, part := range parts {
		if part[0] == ':' {
			params[part[1:]] = searchParts[index]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(searchParts[index:], "/")
			break
		}
	}
	return n, params
}

func (r *router) Handle(c qt.DefaultContext) {
	n, params := r.getRoute(c.Method(), c.Path())
	if n == nil {
		c.SetHandler(func(c qt.DefaultContext) {
			c.Status(http.StatusNotFound).End(fmt.Sprintf("404 not found: %s", c.Path))
		})
	} else {
		c.SetParams(params)
		key := c.Method() + "_" + n.pattern
		c.SetHandler(r.handlers[key])
	}

	c.Next()
}
