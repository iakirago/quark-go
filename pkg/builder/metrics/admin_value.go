package metrics

import (
	"github.com/quarkcms/quark-go/pkg/component/admin/statistic"
	"gorm.io/gorm"
)

type AdminValue struct {
	AdminMetrics
	Precision int
}

// 记录条数
func (p *AdminValue) Count(DB *gorm.DB) *statistic.Component {
	var count int64
	DB.Count(&count)

	return p.Result(count)
}

// 包含组件的结果
func (p *AdminValue) Result(value int64) *statistic.Component {
	return (&statistic.Component{}).Init().SetTitle(p.Title).SetValue(value)
}
