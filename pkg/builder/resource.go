package builder

import (
	"errors"
	"reflect"
	"strings"
)

type Route struct {
	Path        string
	HandlerName string
}

type AdminLayout struct {
	Title         string                   // layout 的左上角 的 title
	Logo          interface{}              // layout 的左上角 的 logo
	HeaderActions []map[string]interface{} // layout 的头部行为
	Layout        string                   // layout 的菜单模式,side：右侧导航，top：顶部导航，mix：混合模式
	SplitMenus    bool                     // layout 的菜单模式为mix时，是否自动分割菜单
	HeaderTheme   string                   // layout 的菜单模式为mix时，顶部主题 'dark' | 'light'
	ContentWidth  string                   // layout 的内容模式,Fluid：定宽 1200px，Fixed：自适应
	NavTheme      string                   // 导航的主题，'light' | 'dark'
	PrimaryColor  string                   // 主题色,"#1890ff"
	FixedHeader   bool                     // 是否固定 header 到顶部
	FixSiderbar   bool                     // 是否固定导航
	IconfontUrl   string                   // 使用 IconFont 的图标配置
	Locale        string                   // 当前 layout 的语言设置，'zh-CN' | 'zh-TW' | 'en-US'
	SiderWidth    int                      // 侧边菜单宽度
	Copyright     string                   // 网站版权 time.Now().Format("2006") + " QuarkGo"
	Links         []map[string]interface{} // 友情链接
}

type Resource struct {
	Providers        []interface{}                  // 服务列表
	Request          *Request                       // 请求数据
	TemplateInstance interface{}                    // 资源模板实例
	UseHandlers      []func(request *Request) error // 中间件方法
}

// 解析UseHandler方法
func (p *Resource) UseHandlerParser() error {
	var err error

	// 执行本资源的方法
	for _, Handler := range p.UseHandlers {
		err = Handler(p.Request)
		if err != nil {
			return err
		}
	}

	return err
}

// 解析路由方法
func (p *Resource) RouteParser() (interface{}, error) {
	var (
		result           interface{}
		err              error
		templateInstance interface{}
	)

	// 获取模板实例
	templateInstance, err = p.GetTemplateInstance()
	if err != nil {
		return nil, err
	}

	// 设置模板实例
	p.SetTemplateInstance(templateInstance)

	// 执行挂载的方法
	templateInstanceRoutes := templateInstance.(interface {
		GetRoutes() []*Route
	}).GetRoutes()

	for _, v := range templateInstanceRoutes {
		if v.Path == p.Request.FullPath() {
			handlerResult := reflect.
				ValueOf(templateInstance).
				MethodByName(v.HandlerName).
				Call([]reflect.Value{
					reflect.ValueOf(p.Request),
					reflect.ValueOf(p),
					reflect.ValueOf(templateInstance),
				})

			if len(handlerResult) == 1 {
				result = handlerResult[0].Interface()
			}
		}
	}

	return result, err
}

// 替换路由中的资源参数
//
//	url := p.RouteToResourceUrl("/api/admin/login/:resource/captchaId") // url = "/api/admin/login/index/captchaId"
func (p *Resource) RouteToResourceUrl(route string) string {
	resourceName := p.Request.ResourceName()

	return strings.ReplaceAll(route, ":resource", resourceName)
}

// 根据路由判断是否为当前加载实例
func (p *Resource) IsCurrentTemplateInstance(provider interface{}) bool {
	providerName := reflect.TypeOf(provider).String()
	getNames := strings.Split(providerName, ".")
	structName := getNames[len(getNames)-1]
	resourceName := p.Request.ResourceName()

	// fmt.Println(providerName)
	// fmt.Println(resourceName)

	return strings.EqualFold(strings.ToLower(structName), strings.ToLower(resourceName))
}

// 获取当前模板实例
func (p *Resource) GetTemplateInstance() (interface{}, error) {
	var templateInstance interface{}

	for _, provider := range p.Providers {

		// 初始化
		getTemplateInstance := provider.(interface {
			Init() interface{}
		}).Init()

		// 获取模板定义的路由
		templateInstanceRoutes := getTemplateInstance.(interface {
			GetRoutes() []*Route
		}).GetRoutes()

		for _, v := range templateInstanceRoutes {
			if v.Path == p.Request.FullPath() {
				if p.IsCurrentTemplateInstance(provider) {
					// 设置实例
					templateInstance = getTemplateInstance
				}
			}
		}
	}

	if templateInstance == nil {
		return nil, errors.New("未获取到实例")
	}

	return templateInstance, nil
}

// 设置当前模板实例
func (p *Resource) SetTemplateInstance(templateInstance interface{}) {
	// 设置实例
	p.TemplateInstance = templateInstance
}

// 处理执行
func (p *Resource) Run() (interface{}, error) {

	// 解析UseHandler方法
	err := p.UseHandlerParser()
	if err != nil {
		return nil, err
	}

	// 解析路由
	return p.RouteParser()
}