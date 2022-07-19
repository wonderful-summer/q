package q

import (
	"fmt"
	qt "github.com/Wonderful-Summer/q-type"
	"net/http"
	"strings"
)

type DefaultEngine interface {
	GetView() qt.DefaultView
	Get(key string) any
}

// Engine 框架的核心结构体，
// 框架的所有内容都会挂载到这个核心的结构体上，包括后面各种插件逻辑
type Engine struct {
	// 路由组, 其实RouterGroup结构体上的所有属性和方法都可以直接挂在Engine上，只是这样不够友好，显得Engine很臃肿
	*RouterGroup

	// 视图操作
	qt.DefaultView

	// 实际处理路由
	router qt.DefaultRouter

	// 所有分组的信息
	groups []*RouterGroup

	// 用户set各种插件
	plugins map[string]any
}

// New 框架实例化入口
func New() *Engine {
	// 如果这里有很多入参，
	// 可以使用默认参数来进行复制操作 https://time.geekbang.org/column/article/386238 选项模式
	engine := &Engine{
		router:  NewRouter(),
		plugins: make(map[string]any),
	}

	// 默认视图
	engine.DefaultView = NewView()

	// 实例化分组，
	engine.RouterGroup = &RouterGroup{engine: engine}

	// 这里相当于group里面放的第一个分组是根分组/ 后续所有分组都会在这里体现，包括分组的子分组，所有分组不过多深，在这里是平级关系
	engine.groups = []*RouterGroup{engine.RouterGroup}

	return engine
}

// SetRouter 使用第三方路由替换原生路由
func (engine *Engine) SetRouter(router qt.DefaultRouter) {
	engine.router = router
}

// SetView 设置第三方view
func (engine *Engine) SetView(view qt.DefaultView) {
	engine.DefaultView = view
}

// GetView 框架运行入库
func (engine *Engine) GetView() qt.DefaultView {
	return engine.DefaultView
}

// Set 设置plugin
func (engine *Engine) Set(name string, plugin any) {
	engine.plugins[name] = plugin
}

// Get 获取对应的plugin
func (engine *Engine) Get(key string) any {
	value, isOk := engine.plugins[key]
	if !isOk {
		fmt.Sprintf("%s plugin 不存在", key)
		return nil
	}
	return value
}

// Run 框架运行入库
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// ServeHTTP 实现 原生http Handler interface
// 也是整个框架的核心处理逻辑入口
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 用于拼接当前路由前面所有上级分组路由上绑定的各种中间件
	var middlewares []qt.HandlerFunc

	// 用于精准识别上级分组使用
	_prefix := ""

	// 准备把本次请求需要用到的所有中间件拼接起来
	// todo: 其实这个操作可以在整个框架启动的时候处理完成，相当于框架启动时把所有路由上该有的中间件全部先规划好，可以节省每次请求时还需要计算中间件的逻辑
	for _, group := range engine.groups {
		// tips: 这里必须要进行路由前缀prefix拼接，不然路由的中间件无法进行串联，会丢失中间件
		if strings.HasPrefix(req.URL.Path, _prefix+group.prefix) {
			_prefix += group.prefix
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	// 创建本次请求的上下文环境
	c := newContext(w, req)

	// 把计算好的所有需要执行的中间件指定给本次请求
	c.SetHandlers(middlewares)

	// 把框架核心结构体挂载到本次请求上，方便随时获取框架层面的配置
	c.SetEngin(engine)

	// 开始执行具体的路由逻辑
	engine.router.Handle(c)
}
