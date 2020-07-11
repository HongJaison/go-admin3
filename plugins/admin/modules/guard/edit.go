package guard

import (
	"fmt"
	tmpl "html/template"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/HongJaison/go-admin3/template/types"

	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/auth"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/modules/db"
	"github.com/HongJaison/go-admin3/modules/errors"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/constant"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/form"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/parameter"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/response"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/table"
	"github.com/HongJaison/go-admin3/template"
)

type ShowFormParam struct {
	Panel  table.Table
	Id     string
	Prefix string
	Param  parameter.Parameters
}

func (g *Guard) ShowForm(ctx *context.Context) {
	fmt.Println(`plugins/admin/modules/guard/edit.go/ShowForm`)

	panel, prefix := g.table(ctx)

	if !panel.GetEditable() {
		fmt.Println(`not editable`)
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	if panel.GetOnlyInfo() {
		fmt.Println(`Only Info`)
		ctx.Redirect(config.Url("/info/" + prefix))
		ctx.Abort()
		return
	}

	if panel.GetOnlyDetail() {
		fmt.Println(`Only Detail`)
		ctx.Redirect(config.Url("/info/" + prefix + "/detail"))
		ctx.Abort()
		return
	}

	if panel.GetOnlyNewForm() {
		fmt.Println(`Only New Form`)
		ctx.Redirect(config.Url("/info/" + prefix + "/new"))
		ctx.Abort()
		return
	}

	id := ctx.Query(constant.EditPKKey)

	fmt.Println(`ID: ` + id)

	if id == "" && prefix != "site" {
		alert(ctx, panel, errors.WrongPK(panel.GetPrimaryKey().Name), g.conn, g.navBtns)
		ctx.Abort()
		return
	}
	if prefix == "site" {
		id = "1"
	}

	showFormParam := &ShowFormParam{
		Panel:  panel,
		Id:     id,
		Prefix: prefix,
		Param: parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField,
			panel.GetInfo().GetSort()).WithPKs(id),
	}

	fmt.Println(showFormParam)
	ctx.SetUserValue(showFormParamKey, showFormParam)
	ctx.Next()
}

func GetShowFormParam(ctx *context.Context) *ShowFormParam {
	return ctx.UserValue[showFormParamKey].(*ShowFormParam)
}

type EditFormParam struct {
	Panel        table.Table
	Id           string
	Prefix       string
	Param        parameter.Parameters
	Path         string
	MultiForm    *multipart.Form
	PreviousPath string
	Alert        tmpl.HTML
	FromList     bool
	IsIframe     bool
	IframeID     string
}

func (e EditFormParam) Value() form.Values {
	return e.MultiForm.Value
}

func (g *Guard) EditForm(ctx *context.Context) {
	previous := ctx.FormValue(form.PreviousKey)
	panel, prefix := g.table(ctx)

	if !panel.GetEditable() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}
	token := ctx.FormValue(form.TokenKey)

	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert(ctx, panel, errors.EditFailWrongToken, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	fromList := isInfoUrl(previous)

	param := parameter.GetParamFromURL(previous, panel.GetInfo().DefaultPageSize,
		panel.GetInfo().GetSort(), panel.GetPrimaryKey().Name)

	if fromList {
		previous = config.Url("/info/" + prefix + param.GetRouteParamStr())
	}

	multiForm := ctx.Request.MultipartForm

	id := multiForm.Value[panel.GetPrimaryKey().Name][0]

	values := ctx.Request.MultipartForm.Value

	ctx.SetUserValue(editFormParamKey, &EditFormParam{
		Panel:        panel,
		Id:           id,
		Prefix:       prefix,
		Param:        param.WithPKs(id),
		Path:         strings.Split(previous, "?")[0],
		MultiForm:    multiForm,
		IsIframe:     form.Values(values).Get(constant.IframeKey) == "true",
		IframeID:     form.Values(values).Get(constant.IframeIDKey),
		PreviousPath: previous,
		FromList:     fromList,
	})
	ctx.Next()
}

func isInfoUrl(s string) bool {
	reg, _ := regexp.Compile("(.*?)info/(.*?)$")
	sub := reg.FindStringSubmatch(s)
	return len(sub) > 2 && !strings.Contains(sub[2], "/")
}

func GetEditFormParam(ctx *context.Context) *EditFormParam {
	return ctx.UserValue[editFormParamKey].(*EditFormParam)
}

func alert(ctx *context.Context, panel table.Table, msg string, conn db.Connection, btns *types.Buttons) {
	if ctx.WantJSON() {
		response.BadRequest(ctx, msg)
	} else {
		response.Alert(ctx, panel.GetInfo().Description, panel.GetInfo().Title, msg, conn, btns)
	}
}

func alertWithTitleAndDesc(ctx *context.Context, title, desc, msg string, conn db.Connection, btns *types.Buttons) {
	response.Alert(ctx, desc, title, msg, conn, btns)
}

func getAlert(msg string) tmpl.HTML {
	return template.Get(config.GetTheme()).Alert().Warning(msg)
}
