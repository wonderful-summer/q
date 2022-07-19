package q

import (
	"bytes"
	"fmt"
	qt "github.com/Wonderful-Summer/q-type"
	"html/template"
)

type View struct {
	// 模板相关
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

func NewView() qt.DefaultView {
	return &View{}
}

func (v *View) Render(name string, data map[string]interface{}) (str string, err error) {
	var tpl bytes.Buffer

	if err = v.htmlTemplates.ExecuteTemplate(&tpl, name, data); err != nil {
		fmt.Println("模板引擎渲染模板失败")
	}
	str = tpl.String()

	return
}

// SetFuncMap 原生模板 实现自定义方法的功能 todo: 需要插件化
func (v *View) SetFuncMap(funcMap template.FuncMap) {
	// 设置模板的处理方法，支持自定义
	v.funcMap = funcMap
}

// LoadHtmlGlob 加载所有模板到内存  todo: 需要插件化
func (v *View) LoadHtmlGlob(pattern string) {
	// 用于把所有模板加载到内存
	v.htmlTemplates = template.Must(template.New("").Funcs(v.funcMap).ParseGlob(pattern))
}
