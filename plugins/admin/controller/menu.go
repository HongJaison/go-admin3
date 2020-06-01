package controller

import (
	"encoding/json"
	errors2 "errors"
	"fmt"
	template2 "html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/auth"
	"github.com/HongJaison/go-admin3/modules/db"
	"github.com/HongJaison/go-admin3/modules/db/dialect"

	"github.com/HongJaison/go-admin3/modules/errors"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/modules/menu"
	"github.com/HongJaison/go-admin3/plugins/admin/models"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/constant"
	form2 "github.com/HongJaison/go-admin3/plugins/admin/modules/form"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/guard"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/parameter"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/response"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/table"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/types"
)

// ShowMenu show menu info page.
func (h *Handler) ShowMenu(ctx *context.Context) {
	h.getMenuInfoPanel(ctx, "")
}

// ShowNewMenu show new menu page.
func (h *Handler) ShowNewMenu(ctx *context.Context) {
	h.showNewMenu(ctx, nil)
}

func (h *Handler) showNewMenu(ctx *context.Context, err error) {
	panel := h.table("menu", ctx)

	formInfo := panel.GetNewForm()

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	h.HTML(ctx, user, types.Panel{
		Content: alert + formContent(aForm().
			SetContent(formInfo.FieldList).
			SetTabContents(formInfo.GroupFieldList).
			SetTabHeaders(formInfo.GroupFieldHeaders).
			SetPrefix(h.config.PrefixFixSlash()).
			SetPrimaryKey(panel.GetPrimaryKey().Name).
			SetUrl(h.routePath("menu_edit")).
			SetHiddenFields(map[string]string{
				form2.TokenKey:    h.authSrv().AddToken(),
				form2.PreviousKey: h.routePath("menu"),
			}).
			SetOperationFooter(formFooter("new", false, false, false)),
			false, ctx.Query(constant.IframeKey) == "true", false, ""),
		Description: template2.HTML(panel.GetForm().Description),
		Title:       template2.HTML(panel.GetForm().Title),
	})
}

// ShowEditMenu show edit menu page.
func (h *Handler) ShowEditMenu(ctx *context.Context) {

	if ctx.Query("id") == "" {
		h.getMenuInfoPanel(ctx, template.Get(h.config.Theme).Alert().Warning(errors.WrongID))

		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
		return
	}

	model := h.table("menu", ctx)
	formInfo, err := model.GetDataWithId(parameter.BaseParam().WithPKs(ctx.Query("id")))

	user := auth.Auth(ctx)

	if err != nil {
		h.HTML(ctx, user, types.Panel{
			Content:     aAlert().Warning(err.Error()),
			Description: template2.HTML(model.GetForm().Description),
			Title:       template2.HTML(model.GetForm().Title),
		})
		return
	}

	h.showEditMenu(ctx, formInfo, nil)
}

func (h *Handler) showEditMenu(ctx *context.Context, formInfo table.FormInfo, err error) {

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	h.HTML(ctx, auth.Auth(ctx), types.Panel{
		Content: alert + formContent(aForm().
			SetContent(formInfo.FieldList).
			SetTabContents(formInfo.GroupFieldList).
			SetTabHeaders(formInfo.GroupFieldHeaders).
			SetPrefix(h.config.PrefixFixSlash()).
			SetPrimaryKey(h.table("menu", ctx).GetPrimaryKey().Name).
			SetUrl(h.routePath("menu_edit")).
			SetOperationFooter(formFooter("edit", false, false, false)).
			SetHiddenFields(map[string]string{
				form2.TokenKey:    h.authSrv().AddToken(),
				form2.PreviousKey: h.routePath("menu"),
			}), false, ctx.Query(constant.IframeKey) == "true", false, ""),
		Description: template2.HTML(formInfo.Description),
		Title:       template2.HTML(formInfo.Title),
	})
	return
}

// DeleteMenu delete the menu of given id.
func (h *Handler) DeleteMenu(ctx *context.Context) {
	models.MenuWithId(guard.GetMenuDeleteParam(ctx).Id).SetConn(h.conn).Delete()
	response.OkWithMsg(ctx, language.Get("delete succeed"))
}

// EditMenu edit the menu of given id.
func (h *Handler) EditMenu(ctx *context.Context) {

	param := guard.GetMenuEditParam(ctx)

	if param.HasAlert() {
		h.getMenuInfoPanel(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
		return
	}

	menuModel := models.MenuWithId(param.Id).SetConn(h.conn)

	// TODO: use transaction
	deleteRolesErr := menuModel.DeleteRoles()
	if db.CheckError(deleteRolesErr, db.DELETE) {
		formInfo, _ := h.table("menu", ctx).GetDataWithId(parameter.BaseParam().WithPKs(param.Id))
		h.showEditMenu(ctx, formInfo, deleteRolesErr)
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
		return
	}
	for _, roleId := range param.Roles {
		_, addRoleErr := menuModel.AddRole(roleId)
		if db.CheckError(addRoleErr, db.INSERT) {
			formInfo, _ := h.table("menu", ctx).GetDataWithId(parameter.BaseParam().WithPKs(param.Id))
			h.showEditMenu(ctx, formInfo, addRoleErr)
			ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
			return
		}
	}

	_, updateErr := menuModel.Update(param.Title, param.Icon, param.Uri, param.Header, param.ParentId)

	if db.CheckError(updateErr, db.UPDATE) {
		formInfo, _ := h.table("menu", ctx).GetDataWithId(parameter.BaseParam().WithPKs(param.Id))
		h.showEditMenu(ctx, formInfo, updateErr)
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
		return
	}

	h.getMenuInfoPanel(ctx, "")
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
}

// NewMenu create a new menu item.
func (h *Handler) NewMenu(ctx *context.Context) {

	param := guard.GetMenuNewParam(ctx)

	if param.HasAlert() {
		h.getMenuInfoPanel(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
		return
	}

	user := auth.Auth(ctx)

	// TODO: use transaction
	menuModel, createErr := models.Menu().SetConn(h.conn).
		New(param.Title, param.Icon, param.Uri, param.Header, param.ParentId, (menu.GetGlobalMenu(user, h.conn)).MaxOrder+1)

	if db.CheckError(createErr, db.INSERT) {
		h.showNewMenu(ctx, createErr)
		return
	}

	for _, roleId := range param.Roles {
		_, addRoleErr := menuModel.AddRole(roleId)
		if db.CheckError(addRoleErr, db.INSERT) {
			h.showNewMenu(ctx, addRoleErr)
			return
		}
	}

	menu.GetGlobalMenu(user, h.conn).AddMaxOrder()

	h.getMenuInfoPanel(ctx, "")
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("menu"))
}

// MenuOrder change the order of menu items.
func (h *Handler) MenuOrder(ctx *context.Context) {

	var data []map[string]interface{}
	_ = json.Unmarshal([]byte(ctx.FormValue("_order")), &data)

	models.Menu().SetConn(h.conn).ResetOrder([]byte(ctx.FormValue("_order")))

	response.Ok(ctx)
}

func (h *Handler) getMenuInfoPanel(ctx *context.Context, alert template2.HTML) {
	user := auth.Auth(ctx)

	tree := aTree().
		SetTree((menu.GetGlobalMenu(user, h.conn)).List).
		SetEditUrl(h.routePath("menu_edit_show")).
		SetUrlPrefix(h.config.Prefix()).
		SetDeleteUrl(h.routePath("menu_delete")).
		SetOrderUrl(h.routePath("menu_order")).
		GetContent()

	header := aTree().GetTreeHeader()
	box := aBox().SetHeader(header).SetBody(tree).GetContent()
	col1 := aCol().SetSize(types.SizeMD(6)).SetContent(box).GetContent()

	formInfo := h.table("menu", ctx).GetNewForm()

	newForm := menuFormContent(aForm().
		SetPrefix(h.config.PrefixFixSlash()).
		SetUrl(h.routePath("menu_new")).
		SetPrimaryKey(h.table("menu", ctx).GetPrimaryKey().Name).
		SetHiddenFields(map[string]string{
			form2.TokenKey:    h.authSrv().AddToken(),
			form2.PreviousKey: h.routePath("menu"),
		}).
		SetOperationFooter(formFooter("menu", false, false, false)).
		SetTitle("New").
		SetContent(formInfo.FieldList).
		SetTabContents(formInfo.GroupFieldList).
		SetTabHeaders(formInfo.GroupFieldHeaders))

	col2 := aCol().SetSize(types.SizeMD(6)).SetContent(newForm).GetContent()

	row := aRow().SetContent(col1 + col2).GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + row,
		Description: "Menus Manage",
		Title:       "Menus Manage",
	})
}

// LoginLog
// added by jaison
type LoginLog struct {
	Username string
	IP       string
	DateTime string
}

// added by jaison
type LoginLogs struct {
	List []LoginLog
}

// added by jaison
func (h *Handler) showLoginLogQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showLoginLogQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"><span class="text-success text-sm">Last 10.</span></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value=""></input>
				<p id="p0" class="help-block" style="display: none;">pls. enter account name.</p>
			</div>`)).
		SetFooter(template.HTML(`
			<button type="button" class="btn btn-primary" id="Button_OK">OK</button>
			<button type="button" class="btn btn-default" style=" margin-left:15px;margin-right:15px;" id="Button_Cancel">Cancel</button>`)).
		GetContent() + template.HTML(`
			<script>
				$('#Button_OK').click(function (e) {
					$.pjax({
						type: 'POST',
						url: this.value,
						data: {username: $('#txt_UserName').val()},
						container: '#pjax-container'
					});
					e.preventDefault();
				});
				$('#Button_Cancel').click(function (e) {
					e.preventDefault();
					$('#txt_UserName').val("");
					$('#p0').attr("style", "display: none");
				});
				$('#txt_UserName').click(function (e) {
					e.preventDefault();
					$('#p0').attr("style", "display: block");
				});
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm,
		Description: template.HTML(``),
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Player Login Logs
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Player Login Logs&nbsp;&nbsp;&nbsp;</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) showLoginLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showLoginLog`)
	user := auth.Auth(ctx)
	param := guard.GetLoginLogParam(ctx)

	panel := h.table("loginlogs", ctx)

	panel.GetInfo().
		Where("username", "=", param.Username)

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "loginlogs", params, panel, "/loginlog/")

	if err != nil {
		h.showLoginLogQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"><span class="text-success text-sm">Last 10.</span></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value=""></input>
				<p id="p0" class="help-block" style="display: none;">pls. enter account name.</p>
			</div>`)).
		SetFooter(template.HTML(`
			<button type="button" class="btn btn-primary" id="Button_OK">OK</button>
			<button type="button" class="btn btn-default" style=" margin-left:15px;margin-right:15px;" id="Button_Cancel">Cancel</button>`)).
		GetContent() + template.HTML(`
			<script>
				$('#Button_OK').click(function (e) {
					$.pjax({
						type: 'POST',
						url: this.value,
						data: {username: $('#txt_UserName').val()},
						container: '#pjax-container'
					});
					e.preventDefault();
				});
				$('#Button_Cancel').click(function (e) {
					e.preventDefault();
					$('#txt_UserName').val("");
					$('#p0').attr("style", "display: none");
				});
				$('#txt_UserName').click(function (e) {
					e.preventDefault();
					$('#p0').attr("style", "display: block");
				});
			</script>`)

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold">Data List</h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title:       "Player Login Logs",
	})
}

// added by jaison
func (h *Handler) ShowLoginLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowLoginLog`)
	h.showLoginLogQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) LoginLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/LoginLog`)
	param := guard.GetLoginLogParam(ctx)

	if param.Username == "" {
		h.showLoginLogQueryBox(ctx, errors2.New("Enter the User!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showLoginLogQueryBox(ctx, errors2.New("No Such User!"))

	checkExist, err := db.WithDriver(h.conn).Table("Players").
		Where("username", "=", param.Username).
		First()

	if db.CheckError(err, db.QUERY) {
		h.showLoginLogQueryBox(ctx, err)
		return
	}

	if checkExist == nil {
		h.showLoginLogQueryBox(ctx, errors2.New("No Such User!"))
		return
	}

	if param.HasAlert() {
		h.showLoginLog(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/loginlog/searchloginlog"))
		return
	}

	h.showLoginLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/loginlog/searchloginlog"))
}

// SearchPlayer
// added by jaison
func interfaces(arr []string) []interface{} {
	var iarr = make([]interface{}, len(arr))

	for key, v := range arr {
		iarr[key] = v
	}

	return iarr
}

// added by jaison
func (h *Handler) showSearchPlayerQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showSearchPlayerQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		SetBody(template.HTML(`
			<div class="input-group input-group-lg">
				<input type="text" id="txt_UserName" maxlength="17" class="form-control text-bold text-blue ui-autocomplete-input" autocomplete="off" value="">
				<span class="input-group-btn">
					<button type="button" id="Button_OK" class="btn btn-info btn-flat">Go</button>
				</span>
			</div>
			<input type="hidden" value="@ViewBag.type" id="type" />`)).
		GetContent() + template.HTML(`
			<script>
				$(document).ready(function () {
					var type = parseInt($('#type').val());
					if (type == 1) {
						AppendTable();
					}
				})
				$('#txt_UserName').keypress(function (e) {
					var key = e.which;
					if (key == 13) {
						e.preventDefault();
						AppendTable();
					}
				});
	
				$('#Button_OK').click(function (e) {
					e.preventDefault();
					AppendTable();
				});
				function AppendTable() {
					$.pjax({
						type: 'POST',
						url: this.value,
						data: {username: $('#txt_UserName').val()},
						container: '#pjax-container'
					});
				}
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm,
		Description: template.HTML(``),
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Search User
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Search User</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) showSearchPlayer(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showSearchPlayer`)

	var (
		upperAgentIds                         string
		agentList, playerList, upperAgentList []map[string]interface{}
		err                                   error

		panel0, panel1, panel2             table.Table
		params0, params1, params2          parameter.Parameters
		panelInfo0, panelInfo1, panelInfo2 table.PanelInfo
		dataTable0, dataTable1, dataTable2 template2.HTML = ``, ``, ``
	)

	user := auth.Auth(ctx)
	param := guard.GetSearchPlayerParam(ctx)

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		SetBody(template.HTML(`
			<div class="input-group input-group-lg">
				<input type="text" id="txt_UserName" maxlength="17" class="form-control text-bold text-blue ui-autocomplete-input" autocomplete="off" value="">
				<span class="input-group-btn">
					<button type="button" id="Button_OK" class="btn btn-info btn-flat">Go</button>
				</span>
			</div>
			<input type="hidden" value="@ViewBag.type" id="type" />`)).
		GetContent() + template.HTML(`
			<script>
				$(document).ready(function () {
					var type = parseInt($('#type').val());
					if (type == 1) {
						AppendTable();
					}
				})
				$('#txt_UserName').keypress(function (e) {
					var key = e.which;
					if (key == 13) {
						e.preventDefault();
						AppendTable();
					}
				});
	
				$('#Button_OK').click(function (e) {
					e.preventDefault();
					AppendTable();
				});
				function AppendTable() {
					$.pjax({
						type: 'POST',
						url: this.value,
						data: {username: $('#txt_UserName').val()},
						container: '#pjax-container'
					});
				}
			</script>`)

	playerList, err = db.WithDriver(h.conn).Table("Players").
		Where("username", "=", param.Username).
		All()

	if db.CheckError(err, db.QUERY) {
		h.showSearchPlayerQueryBox(ctx, err)
		return
	}

	if err != nil {
		h.showSearchPlayerQueryBox(ctx, err)
		return
	}

	if playerList == nil {
		h.showSearchPlayerQueryBox(ctx, errors2.New("No Such User!"))
		return
	}

	if len(playerList) == 1 {
		upperAgentIds = playerList[0][`agentids`].(string)
		arrUpperAgentIds := interfaces(strings.Split(upperAgentIds, `,`))
		upperAgentList, err = db.WithDriver(h.conn).Table("Agents").
			WhereIn("id", arrUpperAgentIds).
			All()

		if db.CheckError(err, db.QUERY) {
			h.showSearchPlayerQueryBox(ctx, err)
			return
		}

		if upperAgentList == nil {
			h.showSearchPlayerQueryBox(ctx, errors2.New("No Upper Agents!"))
			return
		}

		if len(upperAgentList) > 0 {
			panel0 = h.table("playeragentlist", ctx)

			strArrUpperAgents := strings.Split(upperAgentIds, `,`)

			for _, v := range strArrUpperAgents {
				panel0.GetInfo().WhereOr("id", "=", v)
			}

			params0 = parameter.GetParam(ctx.Request.URL, panel0.GetInfo().DefaultPageSize, panel0.GetInfo().SortField, panel0.GetInfo().GetSort())
			panel0, panelInfo0, _, err = h.showTableData(ctx, "playeragentlist", params0, panel0, "/searchplayer/")

			dataTable := aDataTable().
				SetInfoList(panelInfo0.InfoList).
				SetLayout(panel0.GetInfo().TableLayout).
				SetStyle(`hover table-bordered`).
				SetIsTab(true).
				SetHideThead(false).
				SetThead(panelInfo0.Thead)

			dataTable0 = aBox().
				SetTheme(`primary`).
				SetHeader(template.HTML(`
					<h3 class="box-title text-bold" id="d_tip_0">Higher Level AgentList</h3><i class="fa fa-angle-decimal-right"></i>
					<div class="box-tools pull-right">
						<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
					</div>`)).
				WithHeadBorder().
				SetStyle("display: block;").
				SetBody(template.HTML(`<div class="table-responsive">`) +
					dataTable.GetContent() +
					template.HTML(`</div>`)).
				GetContent()
		} else {
			dataTable0 = ``
		}

		if len(playerList) > 0 {
			panel1 = h.table("playerlist", ctx)
			panel1.GetInfo().Where(`username`, `=`, param.Username)

			params1 = parameter.GetParam(ctx.Request.URL, panel1.GetInfo().DefaultPageSize, panel1.GetInfo().SortField, panel1.GetInfo().GetSort())
			panel1, panelInfo1, _, err = h.showTableData(ctx, "playerlist", params1, panel1, "/searchplayer/")

			dataTable := aDataTable().
				SetInfoList(panelInfo1.InfoList).
				SetLayout(panel1.GetInfo().TableLayout).
				SetStyle(`hover table-bordered`).
				SetIsTab(true).
				SetHideThead(false).
				SetThead(panelInfo1.Thead)

			dataTable1 = aBox().
				SetTheme(`primary`).
				SetHeader(panel1.GetInfo().HeaderHtml).
				WithHeadBorder().
				SetStyle("display: block;").
				SetBody(template.HTML(`<div class="table-responsive">`) +
					dataTable.GetContent() +
					template.HTML(`</div>`)).
				GetContent()
		} else {
			dataTable1 = ``
		}

		h.HTML(ctx, user, types.Panel{
			Content:     alert + queryBoxForm + dataTable0 + dataTable1 + dataTable2,
			Description: "",
			Title: template2.HTML(template.HTML(`
				<h1 class="hidden-xs">
					Search User
				</h1>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Search User</li>
				</ol>`)),
		})
		return
	}

	agentList, err = db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", param.Username).
		All()

	if db.CheckError(err, db.QUERY) {
		h.showSearchPlayerQueryBox(ctx, err)
		return
	}

	if err != nil {
		h.showSearchPlayerQueryBox(ctx, err)
		return
	}

	if agentList == nil {
		h.showSearchPlayerQueryBox(ctx, errors2.New("No Such User!"))
		return
	}

	if len(agentList) == 1 {
		upperAgentIds = agentList[0][`agentids`].(string)
		arrUpperAgentIds := interfaces(strings.Split(upperAgentIds, `,`))
		upperAgentList, err = db.WithDriver(h.conn).Table("Agents").
			WhereIn("id", arrUpperAgentIds).
			All()

		if db.CheckError(err, db.QUERY) {
			h.showSearchPlayerQueryBox(ctx, err)
			return
		}

		if upperAgentList == nil {
			h.showSearchPlayerQueryBox(ctx, errors2.New("No Upper Agents!"))
			return
		}

		if len(upperAgentList) > 0 {
			panel0 = h.table("playeragentlist", ctx)

			strArrUpperAgents := strings.Split(upperAgentIds, `,`)

			for _, v := range strArrUpperAgents {
				panel0.GetInfo().WhereOr("id", "=", v)
			}

			params0 = parameter.GetParam(ctx.Request.URL, panel0.GetInfo().DefaultPageSize, panel0.GetInfo().SortField, panel0.GetInfo().GetSort())
			panel0, panelInfo0, _, err = h.showTableData(ctx, "playeragentlist", params0, panel0, "/searchplayer/")

			dataTable := aDataTable().
				SetInfoList(panelInfo0.InfoList).
				SetLayout(panel0.GetInfo().TableLayout).
				SetStyle(`hover table-bordered`).
				SetIsTab(true).
				SetHideThead(false).
				SetThead(panelInfo0.Thead)

			dataTable0 = aBox().
				SetTheme(`primary`).
				SetHeader(template.HTML(`
					<h3 class="box-title text-bold" id="d_tip_0">Higher Level AgentList</h3><i class="fa fa-angle-decimal-right"></i>
					<div class="box-tools pull-right">
						<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
					</div>`)).
				WithHeadBorder().
				SetStyle("display: block;").
				SetBody(template.HTML(`<div class="table-responsive">`) +
					dataTable.GetContent() +
					template.HTML(`</div>`)).
				GetContent()
		} else {
			dataTable0 = ``
		}

		if len(agentList) > 0 {
			panel2 = h.table("agentlist", ctx)
			panel2.GetInfo().
				Where(`username`, `=`, param.Username)

			params2 = parameter.GetParam(ctx.Request.URL, panel2.GetInfo().DefaultPageSize, panel2.GetInfo().SortField, panel2.GetInfo().GetSort())
			panel2, panelInfo2, _, err = h.showTableData(ctx, "agentlist", params2, panel2, "/searchplayer/")

			dataTable := aDataTable().
				SetInfoList(panelInfo2.InfoList).
				SetLayout(panel2.GetInfo().TableLayout).
				// added by jaison
				SetStyle(`hover table-bordered`).
				SetIsTab(true).
				SetHideThead(false).
				SetThead(panelInfo2.Thead)

			dataTable2 = aBox().
				SetTheme(`primary`).
				SetHeader(template.HTML(`
					<h3 class="box-title text-bold" id="d_tip_1">` + param.Username + ` Agent Information</h3><i class="fa fa-angle-decimal-right"></i>
					<div class="box-tools pull-right">
						<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
					</div>`)).
				WithHeadBorder().
				SetStyle("display: block;").
				SetBody(template.HTML(`<div class="table-responsive">`) +
					dataTable.GetContent() +
					template.HTML(`</div>`)).
				GetContent()
		} else {
			dataTable2 = ``
		}

		h.HTML(ctx, user, types.Panel{
			Content:     alert + queryBoxForm + dataTable0 + dataTable1 + dataTable2,
			Description: "",
			Title: template2.HTML(template.HTML(`
				<h1 class="hidden-xs">
					Search User
				</h1>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Search User</li>
				</ol>`)),
		})
		return
	}

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTable0 + dataTable1 + dataTable2,
		Description: "",
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Search User
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Search User</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) ShowSearchPlayer(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowSearchPlayer`)
	h.showSearchPlayerQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) SearchPlayer(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/SearchPlayer`)
	param := guard.GetSearchPlayerParam(ctx)

	if param.Username == "" {
		h.showSearchPlayerQueryBox(ctx, errors2.New("Enter the User!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showSearchPlayerQueryBox(ctx, errors2.New("No Such User!"))

	// check players table first
	playerList, err1 := db.WithDriver(h.conn).Table("Players").
		Where("username", "=", param.Username).
		All()

	if db.CheckError(err1, db.QUERY) {
		h.showSearchPlayerQueryBox(ctx, err1)
		return
	}

	if playerList != nil && len(playerList) > 0 {
		if param.HasAlert() {
			h.showSearchPlayer(ctx, param.Alert)
			ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
			ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/searchplayers"))
			return
		}

		h.showSearchPlayer(ctx, template.HTML(``))
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/searchplayers"))
		return
	}

	// check agents table
	agentList, err2 := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", param.Username).
		All()

	if db.CheckError(err2, db.QUERY) {
		h.showSearchPlayerQueryBox(ctx, err2)
		return
	}

	if agentList == nil || len(agentList) == 0 {
		h.showSearchPlayerQueryBox(ctx, errors2.New("No Such User!"))
		return
	}

	if param.HasAlert() {
		h.showSearchPlayer(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/searchplayers"))
		return
	}

	h.showSearchPlayer(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/searchplayers"))
}

// SearchWinPlayer
// added by jaison
func (h *Handler) showSearchWinPlayersQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showSearchWinPlayersQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`warning`).
		SetStyle("display: block;").
		SetBody(template.HTML(`
			<div class="form-group" id="selectGame">
				<label class="text-blue">Select Start Time</label>
				<select id="select_time" class="form-control">
						<option value="1">StartTime: Before 1 Hour(s)</option>
						<option value="2">StartTime: Before 2 Hour(s)</option>
						<option value="3">StartTime: Before 3 Hour(s)</option>
						<option value="4">StartTime: Before 4 Hour(s)</option>
						<option value="5">StartTime: Before 5 Hour(s)</option>
						<option value="6">StartTime: Before 6 Hour(s)</option>
						<option value="7">StartTime: Before 7 Hour(s)</option>
						<option value="8">StartTime: Before 8 Hour(s)</option>
						<option value="9">StartTime: Before 9 Hour(s)</option>
						<option value="10">StartTime: Before 10 Hour(s)</option>
						<option value="11">StartTime: Before 11 Hour(s)</option>
						<option value="12">StartTime: Before 12 Hour(s)</option>
						<option value="13">StartTime: Before 13 Hour(s)</option>
						<option value="14">StartTime: Before 14 Hour(s)</option>
						<option value="15">StartTime: Before 15 Hour(s)</option>
						<option value="16">StartTime: Before 16 Hour(s)</option>
						<option value="17">StartTime: Before 17 Hour(s)</option>
						<option value="18">StartTime: Before 18 Hour(s)</option>
						<option value="19">StartTime: Before 19 Hour(s)</option>
						<option value="20">StartTime: Before 20 Hour(s)</option>
						<option value="21">StartTime: Before 21 Hour(s)</option>
						<option value="22">StartTime: Before 22 Hour(s)</option>
						<option value="23">StartTime: Before 23 Hour(s)</option>
						<option value="24">StartTime: Before 24 Hour(s)</option>
				</select>
			</div>`)).
		GetContent() + template.HTML(`
			<script>
				$("#select_time").change(function () {
					var a = $(this).children("option:selected").val();
					if (a > 0)
						getPlayerList(a);
				})
				function getPlayerList(t) {
					$.pjax({
						type: 'POST',
						url: this.value,
						data: {diffTime: t},
						container: '#pjax-container'
					});
				}
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm,
		Description: template.HTML(``),
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Winning Online User List
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Winning Online User List</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) showSearchWinPlayers(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showSearchWinPlayers`)

	user := auth.Auth(ctx)
	param := guard.GetSearchWinPlayersParam(ctx)

	queryBoxForm := aBox().
		SetTheme(`warning`).
		SetStyle("display: block;").
		SetBody(template.HTML(`
			<div class="form-group" id="selectGame">
				<label class="text-blue">Select Start Time</label>
				<select id="select_time" class="form-control">
						<option value="1">StartTime: Before 1 Hour(s)</option>
						<option value="2">StartTime: Before 2 Hour(s)</option>
						<option value="3">StartTime: Before 3 Hour(s)</option>
						<option value="4">StartTime: Before 4 Hour(s)</option>
						<option value="5">StartTime: Before 5 Hour(s)</option>
						<option value="6">StartTime: Before 6 Hour(s)</option>
						<option value="7">StartTime: Before 7 Hour(s)</option>
						<option value="8">StartTime: Before 8 Hour(s)</option>
						<option value="9">StartTime: Before 9 Hour(s)</option>
						<option value="10">StartTime: Before 10 Hour(s)</option>
						<option value="11">StartTime: Before 11 Hour(s)</option>
						<option value="12">StartTime: Before 12 Hour(s)</option>
						<option value="13">StartTime: Before 13 Hour(s)</option>
						<option value="14">StartTime: Before 14 Hour(s)</option>
						<option value="15">StartTime: Before 15 Hour(s)</option>
						<option value="16">StartTime: Before 16 Hour(s)</option>
						<option value="17">StartTime: Before 17 Hour(s)</option>
						<option value="18">StartTime: Before 18 Hour(s)</option>
						<option value="19">StartTime: Before 19 Hour(s)</option>
						<option value="20">StartTime: Before 20 Hour(s)</option>
						<option value="21">StartTime: Before 21 Hour(s)</option>
						<option value="22">StartTime: Before 22 Hour(s)</option>
						<option value="23">StartTime: Before 23 Hour(s)</option>
						<option value="24">StartTime: Before 24 Hour(s)</option>
				</select>
			</div>`)).
		GetContent() + template.HTML(`
			<script>
				$("#select_time").change(function () {
					var a = $(this).children("option:selected").val();
					if (a > 0)
						getPlayerList(a);
				})
				function getPlayerList(t) {
					$.pjax({
						type: 'POST',
						url: this.value,
						data: {diffTime: t},
						container: '#pjax-container'
					});
				}
			</script>`)

	nowTime := time.Now().UTC()

	diffTime, _ := strconv.Atoi(param.DiffTime)
	startTime := nowTime.Add(-time.Duration(diffTime) * time.Hour)

	panel := h.table("winningusers", ctx)
	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	// panel, panelInfo, _, err := h.showTableDataWithRawQuery(ctx, "winningusers", "SELECT r.username username, p.isonline isonline, SUM(r.bet) as BetField, SUM(r.win) as WinField FROM Reports r LEFT JOIN Players p ON r.username = p.username WHERE datetime >= '"+startTime.Format("2006-01-02 15:04:05")+"' group by p.isonline, r.username", params, panel, "")
	panel, panelInfo, _, err := h.showTableDataWithRawQuery(ctx, "winningusers", "SELECT r.username username, p.isonline isonline, SUM(r.bet) as BetField, SUM(r.win) as WinField FROM Reports r LEFT JOIN Players p ON r.username = p.username WHERE datetime >= '"+startTime.Format("2006-01-02 15:04:05")+"' group by p.isonline, r.username", params, panel, "")

	if err != nil {
		h.showSearchWinPlayersQueryBox(ctx, err)
		return
	}

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		// SetLayout(panel.GetInfo().TableLayout).
		SetLayout("auto").
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		SetHeader(panel.GetInfo().HeaderHtml).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		GetContent()
	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Winning Online User List</small>
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Winning Online User List</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) ShowSearchWinPlayers(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowSearchWinPlayers`)
	h.showSearchWinPlayersQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) SearchWinPlayers(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/SearchWinPlayers`)
	param := guard.GetSearchWinPlayersParam(ctx)

	if param.DiffTime == "" {
		h.showSearchWinPlayersQueryBox(ctx, errors2.New("Select Time offset!"))
		return
	}

	h.showSearchWinPlayers(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/winusers"))
}

// ScoreLogs
// added by jaison
func ConvertInterface_A(input interface{}) string {

	var strRet string = ""
	object := reflect.ValueOf(input)

	// Make a slice of objects to iterate through and populate the string slice
	var items []interface{}
	for i := 0; i < object.Len(); i++ {
		items = append(items, object.Index(i).Interface())
	}

	// Populate the rest of the items into <records>
	for _, v := range items {
		// item := reflect.ValueOf(v)
		var aCaracter uint8 = v.(uint8)

		strRet += string(aCaracter)
	}

	return strRet
}

// added by jaison
func (h *Handler) showScoreLogQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showScoreLogQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value="">
			</div>
			<div class="form-group">
				<label class="text-blue">Start Date</label>
				<div class="input-group">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="input End Date Time">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#datetimerange_start__goadmin').val(),
						enddate: $('#datetimerange_end__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Set score log
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Set score log</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showScoreLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showScoreLog`)
	user := auth.Auth(ctx)
	param := guard.GetScoreLogParam(ctx)

	panel := h.table("scorelogs", ctx)

	panel.GetInfo().
		Where("username", "=", param.Username).
		Where("datetime", ">=", param.StartDateTime).
		Where("datetime", "<=", param.EndDateTime)

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "scorelogs", params, panel, "/scorelogs/")

	if err != nil {
		h.showScoreLogQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value="">
			</div>
			<div class="form-group">
				<label class="text-blue">Start Date</label>
				<div class="input-group">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="input End Date Time">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#datetimerange_start__goadmin').val(),
						enddate: $('#datetimerange_end__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	var scoreSum float64 = 0

	for _, info := range panelInfo.InfoList {

		// for k, v := range info[`setscore`] {
		// 	fmt.Println(k)
		// 	fmt.Println(v)
		// }
		value, _ := strconv.ParseFloat(info[`setscore`].Value, 64)
		scoreSum += value
	}

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		SetHeader(template.HTML(`
				<h3 class="box-title text-bold">
					<span id="d_tip_1" class="badge bg-yellow">set total：` + fmt.Sprintf("%.2f", scoreSum) + `</span>
				</h3>
				<div class="box-tools pull-right">
					<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
				</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title:       "Score log",
	})
}

// added by jaison
func (h *Handler) ShowScoreLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowScoreLog`)
	h.showScoreLogQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) ScoreLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ScoreLog`)
	param := guard.GetScoreLogParam(ctx)

	if param.Username == "" {
		h.showScoreLogQueryBox(ctx, errors2.New("Enter the User!"))
		return
	}

	if param.StartDateTime == "" {
		h.showScoreLogQueryBox(ctx, errors2.New("Enter Start Date Time!"))
		return
	}

	if param.EndDateTime == "" {
		h.showScoreLogQueryBox(ctx, errors2.New("Enter End Date Time!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showScoreLogQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showScoreLog(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/scorelog/searchscorelog"))
		return
	}

	h.showScoreLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/scorelog/searchscorelog"))
}

// BonusLog
// added by jaison
func (h *Handler) showBonusLogQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showBonusLogQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"><span class="text-success text-sm">query player win bonus log.</span></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value="">
			</div>
			<div class="form-group">
				<label class="text-blue">Start Date</label>
				<div class="input-group">
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					// $('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#datetimerange_start__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Bonus log
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Bonus log</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showBonusLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showBonusLog`)
	user := auth.Auth(ctx)
	param := guard.GetBonusLogParam(ctx)

	panel := h.table("bonuslogs", ctx)

	panel.GetInfo().
		Where("username", "=", param.Username).
		Where("datetime", ">=", param.StartDate)

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "bonuslogs", params, panel, "/bonuslogs/")

	if err != nil {
		h.showBonusLogQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"><span class="text-success text-sm">query player win bonus log.</span></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value="">
			</div>
			<div class="form-group">
				<label class="text-blue">Start Date</label>
				<div class="input-group">
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					// $('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#datetimerange_start__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold">
				<span id="td_currMoney" class="badge bg-yellow"></span>
				<span id="s_tip1" class="text-sm text-success" style=""></span>
			</h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title:       "Bonus log",
	})
}

// added by jaison
func (h *Handler) ShowBonusLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowBonusLog`)
	h.showBonusLogQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) BonusLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/BonusLog`)
	param := guard.GetBonusLogParam(ctx)

	if param.Username == "" {
		h.showBonusLogQueryBox(ctx, errors2.New("Enter the User!"))
		return
	}

	if param.StartDate == "" {
		h.showBonusLogQueryBox(ctx, errors2.New("Enter Start Date Time!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showBonusLogQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showBonusLog(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/bonuslog/searchbonuslog"))
		return
	}

	h.showBonusLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/bonuslog/searchbonuslog"))
}

// DailyPlayerReport
// added by jaison
func (h *Handler) showPlayerReportLogQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showPlayerReportLogQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(``)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">Date Range</label>
				<div class="input-group">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="input End Date Time">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						startdate: $('#datetimerange_start__goadmin').val(),
						enddate: $('#datetimerange_end__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				<span>DailyPlayerReport(30DaysMax)</span>
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				DailyPlayerReport(30DaysMax)
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showPlayerReportLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showPlayerReportLog`)
	user := auth.Auth(ctx)
	param := guard.GetReportLogParam(ctx)

	panel := h.table("playerreportlog", ctx)

	panel.GetInfo().
		Where("datetime", ">=", param.StartDate)

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "bonuslogs", params, panel, "/bonuslogs/")

	if err != nil {
		h.showBonusLogQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold text-muted" id="d_tip_2"><span class="text-success text-sm">query player win bonus log.</span></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">User name</label>
				<input type="text" class="form-control ui-autocomplete-input" id="txt_UserName" maxlength="17" autocomplete="off" value="">
			</div>
			<div class="form-group">
				<label class="text-blue">Start Date</label>
				<div class="input-group">
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					// $('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#datetimerange_start__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold">
				<span id="td_currMoney" class="badge bg-yellow"></span>
				<span id="s_tip1" class="text-sm text-success" style=""></span>
			</h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title:       "Bonus log",
	})
}

// added by jaison
func (h *Handler) ShowPlayerReportLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowPlayerReportLog`)
	h.showPlayerReportLogQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) PlayerReportLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/PlayerReportLog`)
	param := guard.GetReportLogParam(ctx)

	if param.StartDate == "" {
		h.showPlayerReportLogQueryBox(ctx, errors2.New("Enter Start Date Time!"))
		return
	}

	if param.EndDate == "" {
		h.showPlayerReportLogQueryBox(ctx, errors2.New("Enter End Date Time!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showPlayerReportLogQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showPlayerReportLog(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/report/dailyplayerreport"))
		return
	}

	h.showPlayerReportLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/report/dailyplayerreport"))
}

// DailyPlayerReport
// added by jaison
func (h *Handler) showAgentReportLogQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showAgentReportLogQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(``)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">Date Range</label>
				<div class="input-group">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="input End Date Time">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						startdate: $('#datetimerange_start__goadmin').val(),
						enddate: $('#datetimerange_end__goadmin').val(),
					};

					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				<span>DailyAgentReport(30DaysMax)</span>
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				DailyAgentReport(30DaysMax)
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showAgentReportLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showAgentReportLog`)
	user := auth.Auth(ctx)
	param := guard.GetReportLogParam(ctx)

	fmt.Println(param)

	if param.StartDate == "" || param.EndDate == "" {
		h.showAgentReportLogQueryBox(ctx, errors2.New("Select Date Time Range!"))
		return
	}

	// startTime, _ := time.Parse(time.Stamp, param.StartDate)
	// endTime, _ := time.Parse(time.Stamp, param.EndDate)
	// after30Days := startTime.AddDate(0, 0, 30)

	// fmt.Println(startTime)
	// fmt.Println(endTime)
	// fmt.Println(after30Days)

	// if endTime.After(after30Days) {
	// 	h.showAgentReportLogQueryBox(ctx, errors2.New("Date Time Range have to be in 30 days!"))
	// 	return
	// }

	panel := h.table("agentreportlogs", ctx)
	params := parameter.GetParam(ctx.Request.URL, 30, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableDataWithRawQuery(ctx, "agentreportlogs", "SELECT a.username as Username, SUM(r.bet) as Bet, SUM(r.win) as Win, SUM(r.bet)-SUM(r.win) as Report FROM Reports r LEFT JOIN Agents a ON r.agentid = a.id WHERE datetime >= '"+param.StartDate+"' and datetime <= '"+param.EndDate+"' group by r.agentid, a.username order by Report asc", params, panel, "/dailyagentreport/")

	if err != nil {
		h.showAgentReportLogQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(``)).
		SetBody(template.HTML(`
			<div class="form-group">
				<label class="text-blue">Date Range</label>
				<div class="input-group">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="input Start Date Time">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="input End Date Time">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD HH:mm:ss","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						startdate: $('#datetimerange_start__goadmin').val(),
						enddate: $('#datetimerange_end__goadmin').val(),
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		SetHeader(template.HTML(`
			<h3 class="box-title text-bold"><span id="td_totalreport" class="badge bg-yellow"></span></h3>
			<h3 class="box-title text-bold"><span id="td_totalscorecount" class="badge bg-green"></span></h3>
			<h3 class="box-title text-bold"><span id="td_totalscoreamount" class="badge bg-blue"></span></h3>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`)).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title:       "DailyAgentReport(30DaysMax)",
	})
}

// added by jaison
func (h *Handler) ShowAgentReportLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowAgentReportLog`)
	h.showAgentReportLogQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) AgentReportLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/AgentReportLog`)
	param := guard.GetReportLogParam(ctx)

	if param.StartDate == "" {
		h.showAgentReportLogQueryBox(ctx, errors2.New("Enter Start Date Time!"))
		return
	}

	if param.EndDate == "" {
		h.showAgentReportLogQueryBox(ctx, errors2.New("Enter End Date Time!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showAgentReportLogQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showPlayerReportLog(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/report/dailyagentreport"))
		return
	}

	h.showAgentReportLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/report/dailyagentreport"))
}

// GameConfig
// Contains tells whether a contains x.
func IsTableGame(gameId int) bool {
	tableGameIDs := []int{174, 167, 168, 38, 32, 69, 101, 29, 57, 141, 171, 173, 82, 26, 112, 81, 25, 111, 143, 110, 23, 17, 24, 58, 6, 169, 170, 172, 175, 991, 990}

	for _, tableGameID := range tableGameIDs {
		if tableGameID == gameId {
			return true
		}
	}

	return false
}

// added by jaison
func (h *Handler) ResetPayout(requestModel *guard.UpdateConfigRequestModel) (int, error) {
	if requestModel.Id == -1 {
		configList, err := db.Table(`Configs`).
			WithDriver(h.conn).
			All()

		if !db.CheckError(err, db.QUERY) {
			err = nil
		}

		if err != nil {
			return 0, err
		}

		if configList == nil {
			return 0, errors2.New("Can't get game configs")
		}

		for _, item := range configList {
			gameId, _ := strconv.Atoi(item[`id`].(string))

			if IsTableGame(gameId) {
			} else {
				_, err := db.Table(`PayoutResets`).
					WithDriver(h.conn).
					Insert(dialect.H{
						"gameid":    gameId,
						"percent":   requestModel.Percent,
						"timestamp": time.Now().UTC().Format("2006-01-02 15:04:05"),
					})

				if err != nil {
					return 0, err
				}
			}
		}
	} else {
		_, err := db.Table(`PayoutResets`).
			WithDriver(h.conn).
			Insert(dialect.H{
				"gameid": requestModel.Id,
				// "percent":   requestModel.Percent,
				"timestamp": time.Now().UTC().Format("2006-01-02 15:04:05"),
			})

		if !db.CheckError(err, db.QUERY) {
			err = nil
		}

		if err != nil {
			return 0, err
		}
	}

	return 1, nil
}

// added by jaison
func (h *Handler) UpdatePercent(requestModel *guard.UpdateConfigRequestModel) (string, error) {
	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"winchance":  requestModel.Percent,
			"changedate": time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%f", requestModel.Percent), nil
}

// added by jaison
func (h *Handler) UpdateEventRate(requestModel *guard.UpdateConfigRequestModel) (string, error) {
	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"hasevent":   requestModel.CheckState,
			"eventrate":  requestModel.Percent,
			"changedate": time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%f", requestModel.Percent), nil
}

// added by jaison
func (h *Handler) UpdateFreeSpinWinRate(requestModel *guard.UpdateConfigRequestModel) (string, error) {
	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"hasfreespinwinrate": requestModel.CheckState,
			"freespinwinrate":    requestModel.Percent,
			"changedate":         time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%f", requestModel.Percent), nil
}

// added by jaison
func (h *Handler) UpdateRandomBonusLimit(requestModel *guard.UpdateConfigRequestModel) (string, error) {
	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"hasrandombonuslimit": requestModel.CheckState,
			"randombonuslimit":    requestModel.Percent,
			"changedate":          time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%f", requestModel.Percent), nil
}

// added by jaison
func (h *Handler) UpdateDynamicTableInfo(requestModel *guard.UpdateConfigRequestModel) (int, error) {
	newTableSet := 0
	if requestModel.CheckState {
		newTableSet = 1
	}

	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"tableset":   newTableSet,
			"changedate": time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return -1, err
	}

	return 1, nil
}

// added by jaison
func (h *Handler) UpdateCanCloseInfo(requestModel *guard.UpdateConfigRequestModel) (int, error) {
	newCanClose := 0
	if requestModel.CheckState {
		newCanClose = 1
	}

	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"canclose":   newCanClose,
			"changedate": time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return -1, err
	}

	return 1, nil
}

// added by jaison
func (h *Handler) UpdateOpenCloseState(requestModel *guard.UpdateConfigRequestModel) (int, error) {
	newCheckStatus := 1

	if requestModel.CheckState {
		newCheckStatus = 0
	}

	_, err := db.Table(`Configs`).
		WithDriver(h.conn).
		Where("id", "=", requestModel.Id).
		Update(dialect.H{
			"openclose":  newCheckStatus,
			"changedate": time.Now().UTC().Format("2006-01-02 15:04:05"),
		})

	if !db.CheckError(err, db.QUERY) {
		err = nil
	}

	if err != nil {
		return -1, err
	}

	return newCheckStatus, nil
}

// added by jaison
func (h *Handler) ProcessUpdateConfig(ctx *context.Context, requestModel *guard.UpdateConfigRequestModel) (interface{}, error) {
	var (
		returnValue interface{}
		err         error
	)

	strMessage := ``

	switch requestModel.ProcessMode {
	case 1:
		returnValue, err = h.ResetPayout(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else if returnValue.(int) == 1 {
			strMessage = `Operation success!`
		} else {
			strMessage = `Operation failed!`
		}
		break
	case 2:
		returnValue, err = h.UpdatePercent(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else {
			strMessage = `Operation success!`
		}
		break
	case 3:
		returnValue, err = h.UpdateEventRate(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else {
			strMessage = `Operation success!`
		}
		break
	case 4:
		returnValue, err = h.UpdateFreeSpinWinRate(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else {
			strMessage = `Operation success!`
		}
		break
	case 5:
		returnValue, err = h.UpdateRandomBonusLimit(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else {
			strMessage = `Operation success!`
		}
		break
	case 6:
		returnValue, err = h.UpdateDynamicTableInfo(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else if returnValue.(int) == -1 {
			strMessage = `Operation failed!`
		} else {
			strMessage = `Operation success!`
		}
		break
	case 7:
		returnValue, err = h.UpdateCanCloseInfo(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else if returnValue.(int) == -1 {
			strMessage = `Operation failed!`
		} else {
			strMessage = `Operation success!`
		}
		break
	case 8:
		returnValue, err = h.UpdateOpenCloseState(requestModel)

		if err != nil {
			strMessage = err.Error()
		} else if returnValue.(int) == -1 {
			strMessage = `Operation failed!`
		} else {
			strMessage = `Operation success!`
		}
		break
	default:
		returnValue = nil
		strMessage = `Failed to identify method`
		break
	}

	alert := template.HTML(`<div class="alert alert-warning alert-dismissible">
			<button type="button" class="close" data-dismiss="alert" aria-hidden="true">×</button>
			<h4><i class="icon fa fa-info-circle"></i>&nbsp;&nbsp;Response</h4>` +
		strMessage +
		`</div>`)

	requestModel.Alert = alert

	ctx.SetUserValue("UpdateConfigParam", requestModel)

	return returnValue, err
}

// added by jaison
func (h *Handler) RefreshGameConfigs(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/RefreshGameConfigs`)
	user := auth.Auth(ctx)

	param := guard.GetConfigUpdateParam(ctx)

	_, err := h.ProcessUpdateConfig(ctx, param)

	if err != nil {
		h.HTML(ctx, user, types.Panel{
			Content:     aAlert().Warning(err.Error()),
			Description: "",
			Title: template.HTML(`
				<h1 class="hidden-xs">
					Config Setting<small class="logQuery"><a name="scoreLog" href="/gameconfig/create" style="cursor:pointer"><i class="fa fa-fw fa-plus"></i> Add Game</a></small>
				</h1>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Config Setting&nbsp;&nbsp;&nbsp;<a name="scoreLog" href="/gameconfig/create" class="logQuery" style="cursor:pointer"><i class="fa fa-fw fa-plus"></i> Add Game</a></li>
				</ol>`),
		})
		return
	}

	h.ShowGameConfigs(ctx)
}

// added by jaison
func (h *Handler) ShowGameConfigs(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowGameConfigs`)
	user := auth.Auth(ctx)
	alert := template.HTML(``)
	param := guard.GetConfigUpdateParam(ctx)

	if param != nil && param.HasAlert() {
		alert = param.Alert
	}

	ctx.SetUserValue("UpdateConfigParam", nil)

	panel := h.table("gameconfigs", ctx)
	params := parameter.GetParam(ctx.Request.URL, 20, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "gameconfigs", params, panel, "/dailyagentreport/")

	if err != nil {
		h.HTML(ctx, user, types.Panel{
			Content:     aAlert().Warning(err.Error()),
			Description: "",
			Title: template.HTML(`
				<h1 class="hidden-xs">
					Config Setting<small class="logQuery"><a name="scoreLog" href="/gameconfig/create" style="cursor:pointer"><i class="fa fa-fw fa-plus"></i> Add Game</a></small>
				</h1>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Config Setting&nbsp;&nbsp;&nbsp;<a name="scoreLog" href="/gameconfig/create" class="logQuery" style="cursor:pointer"><i class="fa fa-fw fa-plus"></i> Add Game</a></li>
				</ol>`),
		})
		return
	}

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		// SetStyle(`striped datatable table table-bordered`).
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	paginator := panelInfo.Paginator
	paginator = paginator.SetHideEntriesInfo()

	dataTableDiv := aBox().
		SetTheme(`primary`).
		WithHeadBorder().
		SetStyle("display: block;").
		SetHeader(template.HTML(`<h3 class="box-title" style="margin-top: 10px">List Games</h3>`)).
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		SetFooter(paginator.GetContent()).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content: alert + dataTableDiv + template.HTML(`
			<script>
				function isFloat(val) {
					var floatRegex = /^-?\d+(?:[.,]\d*?)?$/;
					if (!floatRegex.test(val))
						return false;
			
					val = parseFloat(val);
					if (isNaN(val))
						return false;
					return true;
				}

				$(document).ready(function () {
					bindButtonEvent();
				});

				function unbindButtonEvent() {
					$("button[name='openclose']").unbind("click");
				}
			
				function bindButtonEvent() {
					unbindButtonEvent();
			
					$("input[name='percent']").each(function (e) {
						if ($(this).val() == 0)
							$(this).val('');
					});
			
					$("input[name='eventrate']").each(function (e) {
						if ($(this).val() == 0)
							$(this).val('');
					});
			
					$("input[name='freespinwinrate']").each(function (e) {
						if ($(this).val() == 0)
							$(this).val('');
					});
			
					$("input[name='randombonuslimit']").each(function (e) {
						if ($(this).val() == 0)
							$(this).val('');
					});
			
					$("input[name='hasevent']").on("click", function () {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $("#hasevent" + gameid).prop("checked");
			
						changeEventRate(gameid, checked);
					});
			
					$("input[name='hasevent']").each(function (index, value) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $("#hasevent" + gameid).prop("checked");
			
						changeEventRate(gameid, checked, 1);
					});
			
					$("input[name='hasfreespinwinrate']").each(function (index, value) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $("#hasfreespinwinrate" + gameid).prop("checked");
			
						changeFreeSpinWinRate(gameid, checked, 1);
					});
			
					$("input[name='hasrandombonuslimit']").each(function (index, value) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $("#hasrandombonuslimit" + gameid).prop("checked");
			
						changeRandomBonusLimit(gameid, checked, 1);
					});
			
					$("button[name='payout_reset']").unbind("click").click(function (e) {
						e.preventDefault();
						var gameid = $(this).data("gameid");
						var resetpercent = $(this).parent().find("input[name='resetpercent']").val();
						if (VaildPercent(resetpercent)) {
							resetPayoutRate(gameid, resetpercent);
						}
					});
			
					$("button[name='game_update']").unbind("click").click(function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var game_percent = $(this).parent().parent().parent().find("input[name='percent']").val();
						if (VaildPercent(game_percent)) {
							PostPercent(gameid, game_percent);
						}
					});
					
					$("button[name='eventrate_update']").unbind("click").click(function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $(this).parent().parent().parent().find("input[name='hasevent']").prop("checked");
						var eventrate = $(this).parent().parent().parent().find("input[name='eventrate']").val();
						if (VaildPercent(eventrate)) {
							PostEventRate(gameid, checked ? 1 : 0, eventrate);
						}
					});
			
					$("button[name='freespinwinrate_update']").unbind("click").click(function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $(this).parent().parent().parent().find("input[name='hasfreespinwinrate']").prop("checked");
						var freespinwinrate = $(this).parent().parent().parent().find("input[name='freespinwinrate']").val();
						if (VaildPercent(freespinwinrate)) {
							PostFreeSpinWinRate(gameid, checked ? 1 : 0, freespinwinrate);
						}
					});
			
					$("button[name='randombonuslimit_update']").unbind("click").click(function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $(this).parent().parent().parent().find("input[name='hasrandombonuslimit']").prop("checked");
						var randombonuslimit = $(this).parent().parent().parent().find("input[name='randombonuslimit']").val();
						if (isFloat(randombonuslimit)) {
							PostRandomBonusLimit(gameid, checked ? 1 : 0, randombonuslimit);
						}
					});
			
					$("button[name='dynamictable_update']").unbind("click").click(function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $(this).parent().parent().parent().find("input[name='dynamictable']").prop("checked");
						
						PostDynamicTableInfo(gameid, checked ? 1 : 0);
					});
			
					$("button[name='canclose_update']").unbind("click").click(function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var checked = $(this).parent().parent().parent().find("input[name='cancloseset']").prop("checked");
						PostCanCloseInfo(gameid, checked ? 1 : 0);
					});
			
					$("button[name='openclose']").unbind("click").on("click", function (e) {
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
						var state = $(this).parent().parent().parent().find("button[name='openclose']").val();
			
						PostOpenCloseState(gameid, state);
					});
			
					$("button[name='del']").unbind("click").click(function (e) {
						e.preventDefault();
			
						var gameid = $(this).parent().parent().parent().find("input[name='game_id']").val();
			
						swal({
							title: "Really delete this game?",
							type: "info",
							showCancelButton: true,
							closeOnConfirm: false,
							showLoaderOnConfirm: true,
						}, function (isConfirm) {
							if (isConfirm) {
								deleteGame(gameid);
								swal("Operation Successful.");
							}
						});
					});
				}
			
				function changeEventRate(i, checked, load) {
					if (checked) {
						document.getElementById("eventrate" + i).removeAttribute("disabled");
						if(load != 1)
							document.getElementById("hasevent" + i).setAttribute("checked", "");
					} else {
						document.getElementById("eventrate" + i).setAttribute("disabled", "");
						if(load != 1)
							document.getElementById("hasevent" + i).setAttribute("checked", "checked");
					}
				}
			
				function changeFreeSpinWinRate(i, checked, load) {
					if (checked) {
						document.getElementById("freespinwinrate" + i).removeAttribute("disabled");
						if (load != 1)
							document.getElementById("hasfreespinwinrate" + i).setAttribute("checked", "");
					} else {
						document.getElementById("freespinwinrate" + i).setAttribute("disabled", "");
						if (load != 1)
							document.getElementById("hasfreespinwinrate" + i).setAttribute("checked", "checked");
					}
				}
			
				function changeRandomBonusLimit(i, checked, load) {
					if (checked) {
						document.getElementById("randombonuslimit" + i).removeAttribute("disabled");
						if (load != 1)
							document.getElementById("hasrandombonuslimit" + i).setAttribute("checked", "");
					} else {
						document.getElementById("randombonuslimit" + i).setAttribute("disabled", "");
						if (load != 1)
							document.getElementById("hasrandombonuslimit" + i).setAttribute("checked", "checked");
					}
				}
			
				function deleteGame(id){
					document.location.href = "/gameconfig/delete?id=" + id;
				}
			
				function VaildPercent(percent) {
					if (!isFloat(percent)) {
						swal("game setting value have to be Integer or decimal type!");
						return false;
					}

					if (parseFloat(percent) < 0 || parseFloat(percent) > 100) {
						swal("game setting value have to be 0-100", "", "warning");
						return false;
					}
					
					return true;
				}

				function resetPayoutRate(id, percent) {
					if (confirm('Are you sure to reset payoutrate?')) {
						// Save it!
						var data = {
							processType: '1',
							id: parseInt(id),
							percent: parseFloat(percent)
						};

						$.pjax({
							type: 'POST',
							url: this.value,
							data: data,
							container: '#pjax-container'
						});
					} else {
					}
				}
			
				function PostPercent(id, percent) {
					var data = {
						processType: '2',
						id: parseInt(id),
						percent: parseFloat(percent)
					};

					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			
				function PostEventRate(id, checked, percent) {
					var data = {
						processType: '3',
						id: parseInt(id),
						checkState: checked,
						percent: parseFloat(percent)
					};

					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			
				function PostFreeSpinWinRate(id, checked, percent) {
					var data = {
						processType: '4',
						id: parseInt(id),
						checkState: checked,
						percent: parseFloat(percent)
					};
					
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			
				function PostRandomBonusLimit(id, checked, percent) {
					var data = {
						processType: '5',
						id: parseInt(id),
						checkState: checked,
						percent: parseFloat(percent)
					};
					
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				} 

				function PostDynamicTableInfo(id, checked) {
					var data = {
						processType: '6',
						id: parseInt(id),
						checkState: checked,
					};
					
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}

				function PostCanCloseInfo(id, checked) {
					var data = {
						processType: '7',
						id: parseInt(id),
						checkState: checked,
					};
					
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}

				function PostOpenCloseState(id, state) {
					var data = {
						processType: '8',
						id: parseInt(id),
						state: state
					};
			
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			</script>`),
		Description: "",
		Title: template.HTML(`
			<h1 class="hidden-xs">
				Config Setting<small class="logQuery"><a name="scoreLog" href="/gameconfig/create" style="cursor:pointer"><i class="fa fa-fw fa-plus"></i> Add Game</a></small>
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Config Setting&nbsp;&nbsp;&nbsp;<a name="scoreLog" href="/gameconfig/create" class="logQuery" style="cursor:pointer"><i class="fa fa-fw fa-plus"></i> Add Game</a></li>
			</ol>`),
	})
}
