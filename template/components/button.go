package components

import (
	"fmt"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/template/icon"
	"github.com/HongJaison/go-admin3/template/types"
	"html/template"
)

type ButtonAttribute struct {
	Name        string
	Content     template.HTML
	Orientation string
	LoadingText template.HTML
	Theme       string
	Type        string
	Size        string
	Href        string
	Style       template.HTMLAttr
	MarginLeft  int
	MarginRight int
	types.Attribute
}

func (compo *ButtonAttribute) SetContent(value template.HTML) types.ButtonAttribute {
	compo.Content = value
	return compo
}

func (compo *ButtonAttribute) SetOrientationRight() types.ButtonAttribute {
	compo.Orientation = "pull-right"
	return compo
}

func (compo *ButtonAttribute) SetOrientationLeft() types.ButtonAttribute {
	compo.Orientation = "pull-left"
	return compo
}

func (compo *ButtonAttribute) SetMarginLeft(px int) types.ButtonAttribute {
	compo.MarginLeft = px
	return compo
}

func (compo *ButtonAttribute) SetSmallSize() types.ButtonAttribute {
	compo.Size = "btn-sm"
	return compo
}

func (compo *ButtonAttribute) SetMiddleSize() types.ButtonAttribute {
	compo.Size = "btn-md"
	return compo
}

func (compo *ButtonAttribute) SetMarginRight(px int) types.ButtonAttribute {
	compo.MarginRight = px
	return compo
}

func (compo *ButtonAttribute) SetLoadingText(value template.HTML) types.ButtonAttribute {
	compo.LoadingText = value
	return compo
}

func (compo *ButtonAttribute) SetThemePrimary() types.ButtonAttribute {
	compo.Theme = "primary"
	return compo
}

func (compo *ButtonAttribute) SetThemeDefault() types.ButtonAttribute {
	compo.Theme = "default"
	return compo
}

func (compo *ButtonAttribute) SetThemeWarning() types.ButtonAttribute {
	compo.Theme = "warning"
	return compo
}

func (compo *ButtonAttribute) SetHref(href string) types.ButtonAttribute {
	compo.Href = href
	return compo
}

func (compo *ButtonAttribute) SetTheme(value string) types.ButtonAttribute {
	compo.Theme = value
	return compo
}

func (compo *ButtonAttribute) SetType(value string) types.ButtonAttribute {
	compo.Type = value
	return compo
}

func (compo *ButtonAttribute) GetContent() template.HTML {

	if compo.MarginLeft != 0 {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="margin-left:%dpx;"`, compo.MarginLeft))
	}

	if compo.MarginRight != 0 {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="margin-right:%dpx;"`, compo.MarginRight))
	}

	if compo.LoadingText == "" {
		compo.LoadingText = icon.Icon(icon.Spinner, 1) + language.GetFromHtml(`Save`)
	}

	return ComposeHtml(compo.TemplateList, *compo, "button")
}
