package display

import (
	"github.com/HongJaison/go-admin3/template/types"
	"html/template"
)

type Loading struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("loading", new(Loading))
}

func (l *Loading) Get(args ...interface{}) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		param := args[0].([]string)

		for i := 0; i < len(param); i++ {
			if value.Value == param[i] {
				return template.HTML(`<i class="fa fa-refresh fa-spin text-primary"></i>`)
			}
		}

		return value.Value
	}
}
