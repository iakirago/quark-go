package adminresource

import (
	"reflect"

	"github.com/quarkcms/quark-go/pkg/builder"
	"github.com/quarkcms/quark-go/pkg/component/admin/card"
	"github.com/quarkcms/quark-go/pkg/component/admin/tabs"
)

// 详情页标题
func (p *Template) DetailTitle(ctx *builder.Context) string {
	value := reflect.ValueOf(ctx.Template).Elem()
	title := value.FieldByName("Title").String()

	return title + "详情"
}

// 渲染详情页组件
func (p *Template) DetailComponentRender(ctx *builder.Context, data map[string]interface{}) interface{} {
	title := p.DetailTitle(ctx)
	formExtraActions := p.DetailExtraActions(ctx)
	fields := p.DetailFieldsWithinComponents(ctx, data)
	formActions := p.DetailActions(ctx)

	return p.DetailWithinCard(
		ctx,
		title,
		formExtraActions,
		fields,
		formActions,
		data,
	)
}

// 在卡片内的详情页组件
func (p *Template) DetailWithinCard(
	ctx *builder.Context,
	title string,
	extra interface{},
	fields interface{},
	actions []interface{},
	data map[string]interface{}) interface{} {

	return (&card.Component{}).
		Init().
		SetTitle(title).
		SetHeaderBordered(true).
		SetExtra(extra).
		SetBody(fields)
}

// 在标签页内的详情页组件
func (p *Template) DetailWithinTabs(
	ctx *builder.Context,
	title string,
	extra interface{},
	fields interface{},
	actions []interface{},
	data map[string]interface{}) interface{} {

	return (&tabs.Component{}).Init().SetTabPanes(fields).SetTabBarExtraContent(extra)
}

// 详情页页面显示前回调
func (p *Template) BeforeDetailShowing(ctx *builder.Context, data map[string]interface{}) map[string]interface{} {
	return data
}
