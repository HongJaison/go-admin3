package components

import (
	"fmt"
	"html/template"

	"github.com/HongJaison/go-admin3/template/types"
)

type BoxAttribute struct {
	// added by jaison
	ParentId  string
	OverlayId string
	Message   string

	Name              string
	Header            template.HTML
	Body              template.HTML
	Footer            template.HTML
	Title             template.HTML
	Theme             string
	HeadBorder        string
	Attr              template.HTMLAttr
	HeadColor         string
	SecondHeaderClass string
	SecondHeader      template.HTML
	SecondHeadBorder  string
	SecondHeadColor   string
	Style             template.HTMLAttr
	Padding           string
	types.Attribute
}

// added by jaison
func (compo *BoxAttribute) SetParentInput(value string) types.BoxAttribute {
	compo.ParentId = value
	return compo
}

// added by jaison
func (compo *BoxAttribute) SetOverlayLoad(value string) types.BoxAttribute {
	compo.OverlayId = value
	return compo
}

// added by jaison
func (compo *BoxAttribute) SetMessage(value string) types.BoxAttribute {
	compo.Message = value
	return compo
}

func (compo *BoxAttribute) SetTheme(value string) types.BoxAttribute {
	compo.Theme = value
	return compo
}

func (compo *BoxAttribute) SetHeader(value template.HTML) types.BoxAttribute {
	compo.Header = value
	return compo
}

func (compo *BoxAttribute) SetBody(value template.HTML) types.BoxAttribute {
	compo.Body = value
	return compo
}

func (compo *BoxAttribute) SetStyle(value template.HTMLAttr) types.BoxAttribute {
	compo.Style = value
	return compo
}

func (compo *BoxAttribute) SetAttr(attr template.HTMLAttr) types.BoxAttribute {
	compo.Attr = attr
	return compo
}

func (compo *BoxAttribute) SetIframeStyle(iframe bool) types.BoxAttribute {
	if iframe {
		compo.Attr = `style="border-radius: 0px;box-shadow:none;border-top:none;margin-bottom: 0px;"`
	}
	return compo
}

func (compo *BoxAttribute) SetFooter(value template.HTML) types.BoxAttribute {
	compo.Footer = value
	return compo
}

func (compo *BoxAttribute) SetTitle(value template.HTML) types.BoxAttribute {
	compo.Title = value
	return compo
}

func (compo *BoxAttribute) SetHeadColor(value string) types.BoxAttribute {
	compo.HeadColor = value
	return compo
}

func (compo *BoxAttribute) WithHeadBorder() types.BoxAttribute {
	compo.HeadBorder = "with-border"
	return compo
}

func (compo *BoxAttribute) SetSecondHeader(value template.HTML) types.BoxAttribute {
	compo.SecondHeader = value
	return compo
}

func (compo *BoxAttribute) SetSecondHeadColor(value string) types.BoxAttribute {
	compo.SecondHeadColor = value
	return compo
}

func (compo *BoxAttribute) SetSecondHeaderClass(value string) types.BoxAttribute {
	compo.SecondHeaderClass = value
	return compo
}

func (compo *BoxAttribute) SetNoPadding() types.BoxAttribute {
	compo.Padding = "padding:0;"
	return compo
}

func (compo *BoxAttribute) WithSecondHeadBorder() types.BoxAttribute {
	compo.SecondHeadBorder = "with-border"
	return compo
}

func (compo *BoxAttribute) GetContent() template.HTML {

	if compo.Style == "" {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="overflow-x: scroll;overflow-y: hidden;%s"`, compo.Padding))
	} else {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="%s"`, string(compo.Style)+compo.Padding))
	}

	return ComposeHtml(compo.TemplateList, *compo, "box")
}
