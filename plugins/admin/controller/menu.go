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
	"github.com/HongJaison/go-admin3/template/chartjs"
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
		h.showLoginLogQueryBox(ctx, errors2.New("Input the Username!"))
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
		h.showSearchPlayerQueryBox(ctx, errors2.New("Input the Username!"))
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

// Member Outstanding
// added by jaison
func (h *Handler) showMemberOutstandingQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showMemberOutstandingQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Member Outstanding</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="row col-md-6">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
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
			<h6 class="hidden-xs">
				Stock Management / Member Outstanding
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Stock Management / Member Outstanding</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showMemberOutstanding(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showMemberOutstanding`)
	user := auth.Auth(ctx)
	param := guard.GetScoreLogParam(ctx)

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		alert += aAlert().Warning(err.Error())
		h.HTML(ctx, user, types.Panel{
			Title: template2.HTML(template.HTML(`
				<h6 class="hidden-xs">
					Member Outstanding
				</h6>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Member Outstanding</li>
				</ol>`)),
			Description: template.HTML(``),
			Content:     alert,
		})
		return
	}

	panel := h.table("memberoutstanding", ctx)

	if param.Username != "" {
		panel.GetInfo().Where("username", "=", param.Username)
	}

	panel.GetInfo().Where("agentid", "=", agentDetail["id"])

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "memberoutstanding", params, panel, "/memberoutstanding/")

	if err != nil {
		h.showMemberOutstandingQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Member Outstanding</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="row col-md-6">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>
			<div class="row col-md-12" style="height:50px"/>
			<div class="row col-md-12">
				<label>`+user.UserName+`</label>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
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
			<h6 class="hidden-xs">
				Stock Management / Member Outstanding
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Stock Management / Member Outstanding</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) ShowMemberOutstanding(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowMemberOutstanding`)
	h.showMemberOutstandingQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) MemberOutstanding(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/MemberOutstanding`)
	param := guard.GetScoreLogParam(ctx)

	// if param.Username == "" {
	// 	h.showMemberOutstandingQueryBox(ctx, errors2.New("Input the Login name!"))
	// 	return
	// }

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showMemberOutstandingQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showMemberOutstanding(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/memberoutstanding"))
		return
	}

	h.showMemberOutstanding(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/memberoutstanding"))
}

// W/L Member
// added by jaison
func (h *Handler) showGameLogsQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showGameLogsQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">W/L Member</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="col-md-3">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">Start date:</label>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="datetimerange_start__goadmin col-md-8" placeholder="Input Start Date">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">End date:</label>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="datetimerange_end__goadmin col-md-8" placeholder="Input End Date">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>
			<div class="row col-md-12" style="height:30px;"></div>
			<div class="row col-md-12">
				<div class="form-group col-md-12">
					<div class="btn-group" data-toggle="buttons">
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="0" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Today
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="1" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Yesterday
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="2" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="3" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="4" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This month
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="5" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last month
						</label>
					</div>
				</div>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
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

				function getMonday(d) {
					d = new Date(d);
					var day = d.getDay(),
						diff = d.getDate() - day + (day == 0 ? -6:1); // adjust when day is sunday
					return new Date(d.setDate(diff));
				}

				function getFirstDayOfMonth(date)
				{
					return new Date(date.getFullYear(), date.getMonth(), 1);
				}

				function getDurationValue(radio) {
					// console.log('getDurationValue');

					var startDate = new Date();
					var endDate = new Date();

					switch(radio.value) {
						case "0":
							startDate.setHours(0, 0, 0);
							break;
						case "1":
							startDate.setDate(startDate.getDate() - 1);
							startDate.setHours(0, 0, 0);
							endDate.setHours(0, 0, 0);
							break;
						case "2":
							startDate = getMonday(startDate);
							startDate.setHours(0, 0, 0);
							break;
						case "3":
							startDate = getMonday(startDate);
							endDate = getMonday(endDate);
							startDate.setDate(startDate.getDate() - 7);
							break;
						case "4":
							startDate = getFirstDayOfMonth(startDate);
							break;
						case "5":
							endDate = getFirstDayOfMonth(endDate);
							startDate.setDate(endDate.getDate() - 5);
							startDate = getFirstDayOfMonth(startDate);
							break;
					}

					var dd = startDate.getDate();
					var mm = startDate.getMonth()+1; 
					var yyyy = startDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					startDate = yyyy + '-' + mm + '-' + dd;

					dd = endDate.getDate();
					mm = endDate.getMonth()+1; 
					yyyy = endDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					endDate = yyyy + '-' + mm + '-' + dd;

					var data = {
						username: '',
						startdate: startDate,
						enddate: endDate,
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h6 class="hidden-xs">
				Report / W/L Member
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Report / W/L Member</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showGameLogs(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showGameLogs`)
	user := auth.Auth(ctx)
	param := guard.GetGameLogsParam(ctx)

	panel := h.table("gamelogs", ctx)

	panel.GetInfo().
		Where("datetime", ">=", param.StartDate).
		Where("datetime", "<=", param.EndDate)

	if param.Username != "" {
		panel.GetInfo().Where("username", "=", param.Username)
	} else {
	}

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "gamelogs", params, panel, "/gamelogs/")

	if err != nil {
		h.showGameLogsQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">W/L Member</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="col-md-3">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">Start date:</label>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="datetimerange_start__goadmin col-md-8" placeholder="Input Start Date">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">End date:</label>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="datetimerange_end__goadmin col-md-8" placeholder="Input End Date">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>
			<div class="row col-md-12" style="height:30px;"></div>
			<div class="row col-md-12">
				<div class="form-group col-md-12">
					<div class="btn-group" data-toggle="buttons">
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="0" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Today
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="1" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Yesterday
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="2" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="3" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="4" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This month
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="5" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last month
						</label>
					</div>
				</div>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
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

				function getMonday(d) {
					d = new Date(d);
					var day = d.getDay(),
						diff = d.getDate() - day + (day == 0 ? -6:1); // adjust when day is sunday
					return new Date(d.setDate(diff));
				}

				function getFirstDayOfMonth(date)
				{
					return new Date(date.getFullYear(), date.getMonth(), 1);
				}

				function getDurationValue(radio) {
					// console.log('getDurationValue');

					var startDate = new Date();
					var endDate = new Date();

					switch(radio.value) {
						case "0":
							startDate.setHours(0, 0, 0);
							break;
						case "1":
							startDate.setDate(startDate.getDate() - 1);
							startDate.setHours(0, 0, 0);
							endDate.setHours(0, 0, 0);
							break;
						case "2":
							startDate = getMonday(startDate);
							startDate.setHours(0, 0, 0);
							break;
						case "3":
							startDate = getMonday(startDate);
							endDate = getMonday(endDate);
							startDate.setDate(startDate.getDate() - 7);
							break;
						case "4":
							startDate = getFirstDayOfMonth(startDate);
							break;
						case "5":
							endDate = getFirstDayOfMonth(endDate);
							startDate.setDate(endDate.getDate() - 5);
							startDate = getFirstDayOfMonth(startDate);
							break;
					}

					var dd = startDate.getDate();
					var mm = startDate.getMonth()+1; 
					var yyyy = startDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					startDate = yyyy + '-' + mm + '-' + dd;

					dd = endDate.getDate();
					mm = endDate.getMonth()+1; 
					yyyy = endDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					endDate = yyyy + '-' + mm + '-' + dd;

					var data = {
						username: '',
						startdate: startDate,
						enddate: endDate,
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
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
			<h6 class="hidden-xs">
				Report / W/L Member
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Report / W/L Member</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) ShowGameLogs(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowGameLogs`)
	h.showGameLogsQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) GameLogs(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/GameLogs`)
	param := guard.GetGameLogsParam(ctx)

	// if param.Username == "" {
	// 	h.showGameLogsQueryBox(ctx, errors2.New("Input the Login name!"))
	// 	return
	// }

	if param.StartDate == "" {
		h.showGameLogsQueryBox(ctx, errors2.New("Input Start Date!"))
		return
	}

	if param.EndDate == "" {
		h.showGameLogsQueryBox(ctx, errors2.New("Input End Date!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showGameLogsQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showGameLogs(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/gamelogs/searchgamelog"))
		return
	}

	h.showScoreLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/gamelogs/searchgamelog"))
}

// Agent Scores
// added by jaison
func (h *Handler) showAgentScoresQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showAgentScoresQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Deposit / Withdrawal</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="row col-md-6">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
						startdate: '',
						enddate: '',
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
			<h6 class="hidden-xs">
				Payment / Deposit / Withdrawal
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Payment / Deposit / Withdrawal</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showAgentScores(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showAgentScores`)
	user := auth.Auth(ctx)
	param := guard.GetScoreLogParam(ctx)

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		alert += aAlert().Warning(err.Error())
		h.HTML(ctx, user, types.Panel{
			Title: template2.HTML(template.HTML(`
				<h6 class="hidden-xs">
					Deposit / Withdrawal
				</h6>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Deposit / Withdrawal</li>
				</ol>`)),
			Description: template.HTML(``),
			Content:     alert,
		})
		return
	}

	panel := h.table("agentscores", ctx)

	if param.Username != "" {
		panel.GetInfo().Where("username", "=", param.Username)
	}

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "scorelogs", params, panel, "/agentscores/")

	if err != nil {
		h.showAgentScoresQueryBox(ctx, err)
		return
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Deposit / Withdrawal</h1>`)).
		SetBody(template.HTML(`
		<div class="row col-md-12">
			<div class="row col-md-6">
				<label class="text-blue col-md-5">Login name:</label>
				<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
			</div>
			<div class="col-md-1">
				<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
			</div>
		</div>
		<div class="row col-md-12" style="height:50px"/>
		<div class="row col-md-12">
			<label>Credit:</label>
			<label class="text-blue">`+ConvertInterface_A(agentDetail["score"])+`</label>
		</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
						startdate: '',
						enddate: '',
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

	// var scoreSum float64 = 0

	// for _, info := range panelInfo.InfoList {

	// 	// for k, v := range info[`setscore`] {
	// 	// 	fmt.Println(k)
	// 	// 	fmt.Println(v)
	// 	// }
	// 	value, _ := strconv.ParseFloat(info[`setscore`].Value, 64)
	// 	scoreSum += value
	// }

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		// added by jaison
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	paginator := panelInfo.Paginator
	paginator = paginator.SetHideEntriesInfo()

	dataTableDiv := aBox().
		SetTheme(`primary`).
		/*SetHeader(template.HTML(`
		<h3 class="box-title text-bold">
			<span id="d_tip_1" class="badge bg-yellow">set total：` + fmt.Sprintf("%.2f", scoreSum) + `</span>
		</h3>
		<div class="box-tools pull-right">
			<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
		</div>`)).*/
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		SetFooter(paginator.GetContent()).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title: template2.HTML(template.HTML(`
			<h6 class="hidden-xs">
				Payment / Deposit / Withdrawal
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Payment / Deposit / Withdrawal</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) ShowAgentScores(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowAgentScores`)
	h.showAgentScoresQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) AgentScores(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/AgentScores`)
	param := guard.GetScoreLogParam(ctx)

	// if param.Username == "" {
	// 	h.showAgentScoresQueryBox(ctx, errors2.New("Input the Login name!"))
	// 	return
	// }

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showAgentScoresQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showAgentScores(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/scorelog/agentscores"))
		return
	}

	h.showAgentScores(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/scorelog/agentscores"))
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
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Statement</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="col-md-3">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">Start date:</label>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="datetimerange_start__goadmin col-md-8" placeholder="Input Start Date">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">End date:</label>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="datetimerange_end__goadmin col-md-8" placeholder="Input End Date">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>
			<div class="row col-md-12" style="height:30px;"></div>
			<div class="row col-md-12">
				<div class="form-group col-md-12">
					<div class="btn-group" data-toggle="buttons">
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="0" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Today
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="1" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Yesterday
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="2" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="3" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="4" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This month
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="5" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last month
						</label>
					</div>
				</div>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
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

				function getMonday(d) {
					d = new Date(d);
					var day = d.getDay(),
						diff = d.getDate() - day + (day == 0 ? -6:1); // adjust when day is sunday
					return new Date(d.setDate(diff));
				}

				function getFirstDayOfMonth(date)
				{
					return new Date(date.getFullYear(), date.getMonth(), 1);
				}

				function getDurationValue(radio) {
					// console.log('getDurationValue');

					var startDate = new Date();
					var endDate = new Date();

					switch(radio.value) {
						case "0":
							startDate.setHours(0, 0, 0);
							break;
						case "1":
							startDate.setDate(startDate.getDate() - 1);
							startDate.setHours(0, 0, 0);
							endDate.setHours(0, 0, 0);
							break;
						case "2":
							startDate = getMonday(startDate);
							startDate.setHours(0, 0, 0);
							break;
						case "3":
							startDate = getMonday(startDate);
							endDate = getMonday(endDate);
							startDate.setDate(startDate.getDate() - 7);
							break;
						case "4":
							startDate = getFirstDayOfMonth(startDate);
							break;
						case "5":
							endDate = getFirstDayOfMonth(endDate);
							startDate.setDate(endDate.getDate() - 5);
							startDate = getFirstDayOfMonth(startDate);
							break;
					}

					var dd = startDate.getDate();
					var mm = startDate.getMonth()+1; 
					var yyyy = startDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					startDate = yyyy + '-' + mm + '-' + dd;

					dd = endDate.getDate();
					mm = endDate.getMonth()+1; 
					yyyy = endDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					endDate = yyyy + '-' + mm + '-' + dd;

					var data = {
						username: '',
						startdate: startDate,
						enddate: endDate,
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h6 class="hidden-xs">
				Payment / Statement
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Payment / Statement</li>
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
		Where("datetime", ">=", param.StartDate).
		Where("datetime", "<=", param.EndDate)

	if param.Username != "" {
		panel.GetInfo().Where("username", "=", param.Username)
	}

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
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Statement</h1>`)).
		SetBody(template.HTML(`
			<div class="row col-md-12">
				<div class="col-md-3">
					<label class="text-blue col-md-5">Login name:</label>
					<input type="text" class="col-md-7" id="txt_username" maxlength="12" autocomplete="off" value="">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">Start date:</label>
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="datetimerange_start__goadmin col-md-8" placeholder="Input Start Date">
				</div>
				<div class="col-md-3">
					<label class="text-blue col-md-4">End date:</label>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="datetimerange_end__goadmin col-md-8" placeholder="Input End Date">
				</div>
				<div class="col-md-1">
					<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
				</div>
			</div>
			<div class="row col-md-12" style="height:30px;"></div>
			<div class="row col-md-12">
				<div class="form-group col-md-12">
					<div class="btn-group" data-toggle="buttons">
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="0" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Today
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="1" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Yesterday
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="2" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="3" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last week
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="4" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> This month
						</label>
						<label class="btn btn-primary form-check-label waves-effect waves-light">
							<input value="5" name="daterange" onchange="getDurationValue(this)" class="form-check-input" type="radio" autocomplete="off"> Last month
						</label>
					</div>
				</div>
			</div>`)).
		// SetFooter(template.HTML(``)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					$('.datetimerange_start__goadmin').on("dp.change", function (e) {
						$('.datetimerange_end__goadmin').data("DateTimePicker").minDate(e.date);
					});
					$('.datetimerange_end__goadmin').on("dp.change", function (e) {
						$('.datetimerange_start__goadmin').data("DateTimePicker").maxDate(e.date);
					});
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
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

				function getMonday(d) {
					d = new Date(d);
					var day = d.getDay(),
						diff = d.getDate() - day + (day == 0 ? -6:1); // adjust when day is sunday
					return new Date(d.setDate(diff));
				}

				function getFirstDayOfMonth(date)
				{
					return new Date(date.getFullYear(), date.getMonth(), 1);
				}

				function getDurationValue(radio) {
					// console.log('getDurationValue');

					var startDate = new Date();
					var endDate = new Date();

					switch(radio.value) {
						case "0":
							startDate.setHours(0, 0, 0);
							break;
						case "1":
							startDate.setDate(startDate.getDate() - 1);
							startDate.setHours(0, 0, 0);
							endDate.setHours(0, 0, 0);
							break;
						case "2":
							startDate = getMonday(startDate);
							startDate.setHours(0, 0, 0);
							break;
						case "3":
							startDate = getMonday(startDate);
							endDate = getMonday(endDate);
							startDate.setDate(startDate.getDate() - 7);
							break;
						case "4":
							startDate = getFirstDayOfMonth(startDate);
							break;
						case "5":
							endDate = getFirstDayOfMonth(endDate);
							startDate.setDate(endDate.getDate() - 5);
							startDate = getFirstDayOfMonth(startDate);
							break;
					}

					var dd = startDate.getDate();
					var mm = startDate.getMonth()+1; 
					var yyyy = startDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					startDate = yyyy + '-' + mm + '-' + dd;

					dd = endDate.getDate();
					mm = endDate.getMonth()+1; 
					yyyy = endDate.getFullYear();

					if (dd < 10) {
						dd = '0' + dd;
					}
					if (mm < 10) {
						mm = '0' + mm;
					}
					endDate = yyyy + '-' + mm + '-' + dd;

					var data = {
						username: '',
						startdate: startDate,
						enddate: endDate,
					};
		
					$.pjax({
						type: 'POST',
						url: this.value,
						data: data,
						container: '#pjax-container'
					});
				}
			</script>`)

	/*var scoreSum float64 = 0

	for _, info := range panelInfo.InfoList {

		// for k, v := range info[`setscore`] {
		// 	fmt.Println(k)
		// 	fmt.Println(v)
		// }
		value, _ := strconv.ParseFloat(info[`setscore`].Value, 64)
		scoreSum += value
	}*/

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
		/*SetHeader(template.HTML(`
		<h3 class="box-title text-bold">
			<span id="d_tip_1" class="badge bg-yellow">set total：` + fmt.Sprintf("%.2f", scoreSum) + `</span>
		</h3>
		<div class="box-tools pull-right">
			<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
		</div>`)).*/
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
			<h6 class="hidden-xs">
				Payment / Statement
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Payment / Statement</li>
			</ol>`)),
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

	// if param.Username == "" {
	// 	h.showScoreLogQueryBox(ctx, errors2.New("Input the Login name!"))
	// 	return
	// }

	if param.StartDate == "" {
		h.showScoreLogQueryBox(ctx, errors2.New("Input Start Date!"))
		return
	}

	if param.EndDate == "" {
		h.showScoreLogQueryBox(ctx, errors2.New("Input End Date!"))
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
					<input type="text" id="DateRange_start__goadmin" name="DateRange_start__goadmin" value="" class="form-control DateRange_start__goadmin" placeholder="Input Start Date">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.DateRange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					// $('.DateRange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#DateRange_start__goadmin').val(),
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
					<input type="text" id="DateRange_start__goadmin" name="DateRange_start__goadmin" value="" class="form-control DateRange_start__goadmin" placeholder="Input Start Date Time">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.DateRange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					// $('.DateRange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#DateRange_start__goadmin').val(),
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
		h.showBonusLogQueryBox(ctx, errors2.New("Input the Username!"))
		return
	}

	if param.StartDate == "" {
		h.showBonusLogQueryBox(ctx, errors2.New("Input Start Date Time!"))
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
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="Input Start Date">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="Input End Date">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
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
type DailyPlayerWinningSumModel struct {
	TimeStamp    string
	TotalPlayers int
	WinPlayers   int
	WinAmount    float64
	LostPlayers  int
	LostAmount   float64
	Alert        template2.HTML
}

// added by jaison
func (h *Handler) showPlayerReportLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showPlayerReportLog`)

	const (
		layoutISO = "2006-01-02"
	)

	user := auth.Auth(ctx)
	param := guard.GetReportLogParam(ctx)

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
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="Input Start Date">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="Input End Date">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
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

	startTime, err1_0 := time.Parse(layoutISO, param.StartDate)
	endTime, err1_1 := time.Parse(layoutISO, param.EndDate)

	if err1_0 != nil || err1_1 != nil {
		h.showPlayerReportLogQueryBox(ctx, errors2.New(`Input Date value in correct date format!`))
		return
	}

	fmt.Println(startTime)
	fmt.Println(endTime)

	listWinningSumReport := make([]DailyPlayerWinningSumModel, 0)

	for date := startTime; int(endTime.Sub(date).Hours()) >= 0; date = date.AddDate(0, 0, 1) {
		segmentEnd := date.AddDate(0, 0, 1)

		fmt.Println(`----------------------`)
		fmt.Println(date)
		fmt.Println(segmentEnd)

		eachDayValue, err := db.WithDriver(h.conn).
			RawQuery(`SELECT username, SUM(bet) AS Bet, SUM(win) AS Win, SUM(bet)-SUM(win) AS Report FROM Reports WHERE datetime >= '` + date.Format("2006-01-02 15:04:05") + `' AND datetime < '` + segmentEnd.Format("2006-01-02 15:04:05") + `' GROUP BY username ORDER BY Report DESC`)

		fmt.Println(eachDayValue)

		var value DailyPlayerWinningSumModel = DailyPlayerWinningSumModel{
			TimeStamp:    date.Format(layoutISO),
			TotalPlayers: 0,
			WinPlayers:   0,
			WinAmount:    0,
			LostPlayers:  0,
			LostAmount:   0,
			Alert:        template.HTML(``),
		}

		if !db.CheckError(err, db.QUERY) {
			value.TotalPlayers = len(eachDayValue)

			for _, playerItem := range eachDayValue {

				report, _ := strconv.ParseFloat(ConvertInterface_A(playerItem[`Report`]), 64)

				if report >= 0 {
					value.LostPlayers++
					value.LostAmount += report
				} else {
					value.WinPlayers++
					value.WinAmount += report
				}
			}
		}

		listWinningSumReport = append(listWinningSumReport, value)
	}

	fmt.Println(listWinningSumReport)

	labels := []string{}
	wonAmountData := []float64{}
	loseAmountData := []float64{}
	totalPlayers := []float64{}
	winPlayers := []float64{}
	lostPlayers := []float64{}

	for _, winningSumReport := range listWinningSumReport {
		labels = append(labels, winningSumReport.TimeStamp)
		wonAmountData = append(wonAmountData, winningSumReport.WinAmount)
		loseAmountData = append(loseAmountData, winningSumReport.LostAmount)

		totalPlayers = append(totalPlayers, float64(winningSumReport.TotalPlayers))
		winPlayers = append(winPlayers, float64(winningSumReport.WinPlayers))
		lostPlayers = append(lostPlayers, float64(winningSumReport.LostPlayers))
	}

	line1 := chartjs.Line()
	lineChart1 := line1.
		SetID("Amount").
		SetHeight(200).
		SetTitle("Win,Lose Amount").
		SetLabels(labels).
		AddDataSet("Won Amounts").
		DSData(wonAmountData).
		DSBorderColor("rgb(219, 186, 70)").
		DSLineTension(0.1).
		AddDataSet("Lost Amounts").
		DSData(loseAmountData).
		DSBorderColor("rgb(0, 186, 70)").
		DSLineTension(0.1).
		GetContent()

	line2 := chartjs.Line()
	lineChart2 := line2.
		SetID("Players").
		SetHeight(200).
		SetTitle("Win,Lose Players").
		SetLabels(labels).
		AddDataSet("Win Players").
		DSData(winPlayers).
		DSBorderColor("rgb(219, 186, 70)").
		DSLineTension(0.1).
		AddDataSet("Lost Players").
		DSData(lostPlayers).
		DSBorderColor("rgb(0, 186, 70)").
		DSLineTension(0.1).
		AddDataSet("Total Players").
		DSData(totalPlayers).
		DSBorderColor("rgb(255, 0, 0)").
		DSLineTension(0.1).
		GetContent()

	dataTableDiv1_1 := aBox().
		SetTheme(`info`).
		WithHeadBorder().
		SetStyle("display: block;").
		SetHeader(template.HTML(`
			<h3 class="box-title">Win,Lose Amount</h3>
			<div class="box-tools pull-right">
				<button type="button" class="btn btn-box-tool" data-widget="collapse">
					<i class="fa fa-minus"></i>
				</button>
			</div>`)).
		SetBody(lineChart1).
		GetContent()

	dataTableDiv1_2 := aBox().
		SetTheme(`info`).
		WithHeadBorder().
		SetStyle("display: block;").
		SetHeader(template.HTML(`
			<h3 class="box-title">Win,Lose Players</h3>
			<div class="box-tools pull-right">
				<button type="button" class="btn btn-box-tool" data-widget="collapse">
					<i class="fa fa-minus"></i>
				</button>
			</div>`)).
		SetBody(lineChart2).
		GetContent()

	panel2 := h.table("topwinplayers", ctx)
	params2 := parameter.GetParam(ctx.Request.URL, panel2.GetInfo().DefaultPageSize, panel2.GetInfo().SortField, panel2.GetInfo().GetSort())
	panel2, panelInfo2, _, err2 := h.showTableDataWithRawQuery(ctx, "topwinplayers", `SELECT TOP 100 username, SUM(bet) as Bet, SUM(win) as Win, SUM(bet)-SUM(win) as Report FROM Reports WHERE datetime >='`+param.StartDate+`' and datetime < '`+param.EndDate+`' GROUP BY username ORDER BY Report ASC`, params2, panel2, "")

	if err2 != nil {
		h.showPlayerReportLogQueryBox(ctx, err2)
		return
	}

	dataTable2 := aDataTable().
		SetInfoList(panelInfo2.InfoList).
		// SetLayout(panel.GetInfo().TableLayout).
		SetLayout("auto").
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo2.Thead)

	dataTableDiv2 := aBox().
		SetTheme(`primary`).
		WithHeadBorder().
		SetStyle("display: block;").
		SetHeader(template.HTML(`
			<h3 class="box-title">TOP 100 Win Players</h3>
			<div class="box-tools pull-right">
				<button type="button" class="btn btn-box-tool" data-widget="collapse">
					<i class="fa fa-minus"></i>
				</button>
			</div>`)).
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable2.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	panel3 := h.table("toplostplayers", ctx)
	params3 := parameter.GetParam(ctx.Request.URL, panel3.GetInfo().DefaultPageSize, panel3.GetInfo().SortField, panel3.GetInfo().GetSort())
	panel3, panelInfo3, _, err3 := h.showTableDataWithRawQuery(ctx, "toplostplayers", `SELECT TOP 100 username, SUM(bet) as Bet, SUM(win) as Win, SUM(bet)-SUM(win) as Report FROM Reports WHERE datetime >='`+param.StartDate+`' and datetime < '`+param.EndDate+`' GROUP BY username ORDER BY Report DESC`, params3, panel3, "")

	if err3 != nil {
		h.showPlayerReportLogQueryBox(ctx, err3)
		return
	}

	dataTable3 := aDataTable().
		SetInfoList(panelInfo3.InfoList).
		// SetLayout(panel.GetInfo().TableLayout).
		SetLayout("auto").
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo3.Thead)

	dataTableDiv3 := aBox().
		SetTheme(`primary`).
		WithHeadBorder().
		SetStyle("display: block;").
		SetHeader(template.HTML(`
			<h3 class="box-title">TOP 100 Lost Players</h3>
			<div class="box-tools pull-right">
				<button type="button" class="btn btn-box-tool" data-widget="collapse">
					<i class="fa fa-minus"></i>
				</button>
			</div>`)).
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable3.GetContent() +
			template.HTML(`</div>`)).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv1_1 + dataTableDiv1_2 + dataTableDiv2 + dataTableDiv3,
		Description: "",
		Title:       "DailyPlayerReport(30DaysMax)",
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
		h.showPlayerReportLogQueryBox(ctx, errors2.New("Input Start Date!"))
		return
	}

	if param.EndDate == "" {
		h.showPlayerReportLogQueryBox(ctx, errors2.New("Input End Date!"))
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
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="Input Start Date">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="Input End Date">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
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
					<input type="text" id="datetimerange_start__goadmin" name="datetimerange_start__goadmin" value="" class="form-control datetimerange_start__goadmin" placeholder="Input Start Date">
					<span class="input-group-addon" style="border-left: 0; border-right: 0;">-</span>
					<input type="text" id="datetimerange_end__goadmin" name="datetimerange_end__goadmin" value="" class="form-control datetimerange_end__goadmin" placeholder="Input End Date">
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.datetimerange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en"});
					$('.datetimerange_end__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
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
		h.showAgentReportLogQueryBox(ctx, errors2.New("Input Start Date!"))
		return
	}

	if param.EndDate == "" {
		h.showAgentReportLogQueryBox(ctx, errors2.New("Input End Date!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showAgentReportLogQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showAgentReportLog(ctx, param.Alert)
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

// UpdateProfile
// added by jaison
func (h *Handler) showEditProfileQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showEditProfileQueryBox`)

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
			<input data-val="true" data-val-number="The field agentid must be a number." data-val-required="The agentid field is required." id="agentid" name="agentid" type="hidden"/>
			<div class="form-group">
				<label>Username</label>
				<input class="form-control" data-val="true" maxlength="50" value="`+user.UserName+`" disabled>
				<p id="p1"></p>
			</div>
			<div class="form-group">
				<label>Old password</label>
				<input class="form-control" data-val="true" data-val-required="The Current password field is required." id="txt_oldPassword" maxlength="50" name="OldPassword" type="password">
				<p id="p1"></p>
			</div>
			<div class="form-group">
				<label>New password</label>
				<input class="form-control" data-val="true" data-val-length="The New password must be at least 6 characters long." data-val-length-max="50" data-val-length-min="6" data-val-required="The New password field is required." id="txt_newPassword" maxlength="50" name="NewPassword" type="password">
				<p id="p2"></p>
			</div>
			<div class="form-group">
				<label>Confirm new password</label>
				<input class="form-control" data-val="true" data-val-equalto="'Confirm new password' and 'New password' do not match." data-val-equalto-other="*.NewPassword" id="txt_rePassword" maxlength="50" name="ConfirmPassword" type="password">
				<p id="p3"></p>
			</div>`)).
		SetFooter(template.HTML(`
			<button type="button" class="btn btn-primary" id="Button_OK">OK</button>
			<button type="button" class="btn btn-default" id="Cancel_button">Cancel</button>`)).
		GetContent() + template.HTML(`
			<script type="text/javascript">
				function checkPassWord(n) { return /^(?=.*?[0-9])(?=.*?[A-Z])(?=.*?[a-z])[0-9A-Za-z!)-_]{6,15}$/.test(n) }
				$(function () {
					$("#txt_oldPassword").blur(function () {
						//$.trim($(this).val()).length <= 0 ? ($(".box-body div:eq(0)").removeClass().addClass("form-group has-warning"),
						//$("#p1").removeClass().addClass("help-block").text("enter old password.")) : $(".box-body div:eq(0)").removeClass().addClass("form-group has-success")
					}),
					$("#txt_newPassword").focus(function () {
						$("#p2").removeClass().addClass("help-block").text("Password with minimum 6 characters, must with combination of numbers and alphabets. At least a capital letter and a small letter.")
					}),
					$("#txt_newPassword").blur(function () {
						//checkPassWord($.trim($(this).val())) ? $(this).val() == $.trim($("#txt_oldPassword").val()) ? ($(".box-body div:eq(1)").removeClass().addClass("form-group has-warning"), $("#p2").removeClass().addClass("help-block").text("confirm new password.")) : $(".box-body div:eq(1)").removeClass().addClass("form-group has-success") : ($(".box-body div:eq(1)").removeClass().addClass("form-group has-warning"), $("#p2").removeClass().addClass("help-block").text("Password with minimum 6 characters, must with combination of numbers and alphabets. At least a capital letter and a small letter."))
					}),
					$("#txt_rePassword").focus(function () {
						$("#p3").removeClass().addClass("smsg").text("confirm new password.")
					}),
					$("#txt_rePassword").blur(function () {
						//checkPassWord($.trim($(this).val())) && $.trim($(this).val()) == $.trim($("#txt_newPassword").val()) ? $(this).val() == $.trim($("#txt_oldPassword").val()) ? ($(".box-body div:eq(2)").removeClass().addClass("form-group has-warning"), $("#p3").removeClass().addClass("help-block").text("the new password cannot be the same as the old one.")) : $(".box-body div:eq(2)").removeClass().addClass("form-group has-success") : ($(".box-body div:eq(2)").removeClass().addClass("form-group has-warning"), $("#p3").removeClass().addClass("help-block").text("please confirm that the new password or the two input is not the same."))
					}),
					$("#Button_OK").click(function (e) {
						var a = $(".has-warning").length;

						if (a > 0) return false;

						$.pjax({
							type: 'POST',
							url: this.value,
							data: {
								oldPassWd: $("#txt_oldPassword").val(),
								newPassWd: $("#txt_newPassword").val(),
								rePassWd: $("#txt_rePassword").val()
							},
							container: '#pjax-container'
						});
						e.preventDefault();
						return true;
					})
				});
				$('#Cancel_button').click(function () {
					$('#txt_oldPassword').val("");
					$('#txt_newPassword').val("");
					$('#txt_rePassword').val("");
				})
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Change Password
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Change Password&nbsp;&nbsp;&nbsp;</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showEditProfile(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showEditProfile`)
	user := auth.Auth(ctx)

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(``)).
		SetBody(template.HTML(`
			<input data-val="true" data-val-number="The field agentid must be a number." data-val-required="The agentid field is required." id="agentid" name="agentid" type="hidden"/>
			<div class="form-group">
				<label>Username</label>
				<input class="form-control" data-val="true" maxlength="50" value="`+user.UserName+`" disabled>
				<p id="p1"></p>
			</div>
			<div class="form-group">
				<label>Old password</label>
				<input class="form-control" data-val="true" data-val-required="The Current password field is required." id="txt_oldPassword" maxlength="50" name="OldPassword" type="password">
				<p id="p1"></p>
			</div>
			<div class="form-group">
				<label>New password</label>
				<input class="form-control" data-val="true" data-val-length="The New password must be at least 6 characters long." data-val-length-max="50" data-val-length-min="6" data-val-required="The New password field is required." id="txt_newPassword" maxlength="50" name="NewPassword" type="password">
				<p id="p2"></p>
			</div>
			<div class="form-group">
				<label>Confirm new password</label>
				<input class="form-control" data-val="true" data-val-equalto="'Confirm new password' and 'New password' do not match." data-val-equalto-other="*.NewPassword" id="txt_rePassword" maxlength="50" name="ConfirmPassword" type="password">
				<p id="p3"></p>
			</div>`)).
		SetFooter(template.HTML(`
			<button type="button" class="btn btn-primary" id="Button_OK">OK</button>
			<button type="button" class="btn btn-default" id="Cancel_button">Cancel</button>`)).
		GetContent() + template.HTML(`
			<script type="text/javascript">
				function checkPassWord(n) { return /^(?=.*?[0-9])(?=.*?[A-Z])(?=.*?[a-z])[0-9A-Za-z!)-_]{6,15}$/.test(n) }
				$(function () {
					$("#txt_oldPassword").blur(function () {
						//$.trim($(this).val()).length <= 0 ? ($(".box-body div:eq(0)").removeClass().addClass("form-group has-warning"),
						//$("#p1").removeClass().addClass("help-block").text("enter old password.")) : $(".box-body div:eq(0)").removeClass().addClass("form-group has-success")
					}),
					$("#txt_newPassword").focus(function () {
						$("#p2").removeClass().addClass("help-block").text("Password with minimum 6 characters, must with combination of numbers and alphabets. At least a capital letter and a small letter.")
					}),
					$("#txt_newPassword").blur(function () {
						//checkPassWord($.trim($(this).val())) ? $(this).val() == $.trim($("#txt_oldPassword").val()) ? ($(".box-body div:eq(1)").removeClass().addClass("form-group has-warning"), $("#p2").removeClass().addClass("help-block").text("confirm new password.")) : $(".box-body div:eq(1)").removeClass().addClass("form-group has-success") : ($(".box-body div:eq(1)").removeClass().addClass("form-group has-warning"), $("#p2").removeClass().addClass("help-block").text("Password with minimum 6 characters, must with combination of numbers and alphabets. At least a capital letter and a small letter."))
					}),
					$("#txt_rePassword").focus(function () {
						$("#p3").removeClass().addClass("smsg").text("confirm new password.")
					}),
					$("#txt_rePassword").blur(function () {
						//checkPassWord($.trim($(this).val())) && $.trim($(this).val()) == $.trim($("#txt_newPassword").val()) ? $(this).val() == $.trim($("#txt_oldPassword").val()) ? ($(".box-body div:eq(2)").removeClass().addClass("form-group has-warning"), $("#p3").removeClass().addClass("help-block").text("the new password cannot be the same as the old one.")) : $(".box-body div:eq(2)").removeClass().addClass("form-group has-success") : ($(".box-body div:eq(2)").removeClass().addClass("form-group has-warning"), $("#p3").removeClass().addClass("help-block").text("please confirm that the new password or the two input is not the same."))
					}),
					$("#Button_OK").click(function (e) {
						var a = $(".has-warning").length;

						if (a > 0) return false;

						$.pjax({
							type: 'POST',
							url: this.value,
							data: {
								oldPassWd: $("#txt_oldPassword").val(),
								newPassWd: $("#txt_newPassword").val(),
								rePassWd: $("#txt_rePassword").val()
							},
							container: '#pjax-container'
						});
						e.preventDefault();
						return true;
					})
				});
				$('#Cancel_button').click(function () {
					$('#txt_oldPassword').val("");
					$('#txt_newPassword').val("");
					$('#txt_rePassword').val("");
				})
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h1 class="hidden-xs">
				Change Password
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Change Password&nbsp;&nbsp;&nbsp;</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) ShowEditProfile(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowEditProfile`)
	h.showEditProfileQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) EditProfile(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/EditProfile`)

	user := auth.Auth(ctx)
	param := guard.GetEditProfileParam(ctx)

	if param.PasswordOld == "" || param.PasswordNew == "" {
		h.showEditProfileQueryBox(ctx, errors2.New("Need to input old and new password"))
		return
	}

	if param.PasswordCon == "" {
		h.showEditProfileQueryBox(ctx, errors2.New("Need to confirm password, please input confirm password"))
		return
	}

	if param.PasswordNew != param.PasswordCon {
		h.showEditProfileQueryBox(ctx, errors2.New("please confirm correct password!"))
		return
	}

	if user.Password != param.PasswordOld {
		h.showEditProfileQueryBox(ctx, errors2.New("Old Password is wrong!"))
		return
	}

	_, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		h.showEditProfileQueryBox(ctx, err)
		return
	}

	_, updateUserErr := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		Update(dialect.H{
			"password": param.PasswordNew,
		})

	if db.CheckError(updateUserErr, db.UPDATE) {
		h.showEditProfileQueryBox(ctx, updateUserErr)
		return
	}

	if param.HasAlert() {
		h.showEditProfile(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/profile/edit"))
		return
	}

	h.showEditProfile(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/profile/edit"))
}

// RedPacketLog
// added by jaison
func (h *Handler) showRedPacketLogQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showRedPacketLogQueryBox`)

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
			<h3 class="box-title text-bold text-muted" id="d_tip_2"><span class="text-success text-sm">query player redpacket log.</span></h3>
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
					<input type="text" id="DateRange_start__goadmin" name="DateRange_start__goadmin" value="" class="form-control DateRange_start__goadmin" placeholder="Input Start Date">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.DateRange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					// $('.DateRange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#DateRange_start__goadmin').val(),
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
				RedPacket log
			</h1>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>RedPacket log</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showRedPacketLog(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showRedPacketLog`)
	user := auth.Auth(ctx)
	param := guard.GetRedPacketLogParam(ctx)

	checkExist, errExist := db.WithDriver(h.conn).Table("Players").
		Where("username", "=", param.Username).
		First()

	if db.CheckError(errExist, db.QUERY) {
		h.showRedPacketLogQueryBox(ctx, errExist)
		return
	}

	if checkExist == nil {
		checkExist, errExist = db.WithDriver(h.conn).Table("Agents").
			Where("username", "=", param.Username).
			First()

		if db.CheckError(errExist, db.QUERY) {
			h.showRedPacketLogQueryBox(ctx, errExist)
			return
		}
	} else {
	}

	panel := h.table("redpacketlogs", ctx)

	panel.GetInfo().
		Where("username", "=", param.Username).
		Where("datetime", ">=", param.StartDate)

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	panel, panelInfo, _, err := h.showTableData(ctx, "redpacketlogs", params, panel, "/redpacketlogs/")

	if err != nil {
		h.showRedPacketLogQueryBox(ctx, err)
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
					<input type="text" id="DateRange_start__goadmin" name="DateRange_start__goadmin" value="" class="form-control DateRange_start__goadmin" placeholder="Input Start Date Time">
					<span class="input-group-addon"><i class="fa fa-calendar fa-fw"></i></span>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary" id="Button_OK">OK</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(function () {
					$('.DateRange_start__goadmin').datetimepicker({"format":"YYYY-MM-DD","locale":"en","useCurrent":true});
					// $('.DateRange_start__goadmin').data("DateTimePicker").maxDate(e.date);
				});
				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_UserName').val(),
						startdate: $('#DateRange_start__goadmin').val(),
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
func (h *Handler) ShowRedPacketLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowRedPacketLog`)
	h.showRedPacketLogQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) RedPacketLog(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/RedPacketLog`)
	param := guard.GetRedPacketLogParam(ctx)

	if param.Username == "" {
		h.showRedPacketLogQueryBox(ctx, errors2.New("Input the Username!"))
		return
	}

	if param.StartDate == "" {
		h.showRedPacketLogQueryBox(ctx, errors2.New("Input Start Date!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showRedPacketLogQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showRedPacketLog(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/redpacketlog/searchredpacketlog"))
		return
	}

	h.showRedPacketLog(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/redpacketlog/searchredpacketlog"))
}

// ShareHolder
// Add Shareholder
// added by jaison
func (h *Handler) showAddShareHolderQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showAddShareHolderQueryBox`)

	user := auth.Auth(ctx)

	alert := template2.HTML(``)

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		alert += aAlert().Warning(err.Error())
		h.HTML(ctx, user, types.Panel{
			Title: template2.HTML(template.HTML(`
				<h6 class="hidden-xs">
					Shareholder / Add Shareholder
				</h6>
				<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
					<li>Shareholder / Add Shareholder</li>
				</ol>`)),
			Description: template.HTML(``),
			Content:     alert,
		})
		return
	}

	// scoreSum, _ := strconv.ParseFloat(ConvertInterface_A(agentDetail[`score`]), 64)

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Add Shareholder</h1>`)).
		SetBody(template.HTML(`
			<div class="box-body">
				<div class="row col-md-12">
					<div class="row col-md-12">
						<h3 style='display:block; float: left;'>Basic Info</h3>
					</div>
					<div class="row col-md-12">
						<div class="form-group col-md-6">
							<label class="asterisk control-label col-md-3">Username</label>
							<input class="col-md-9" type="text" id="txt_username" maxlength="12" required>
							<label class="control-label col-md-3"></label>
							<p id="pusername" class="col-md-9 hit">Enter only number (0-9) or letter (A-Z, a-z).</p>
						</div>
						<div class="form-group col-md-6">
							<label class="control-label col-md-3">Nickname</label>
							<input class="col-md-9" type="text" id="txt_nickname" maxlength="12">
							<label class="control-label col-md-3"></label>
							<p id="pnickname" class="col-md-9 hit">Enter only number (0-9) or letter (A-Z, a-z).</p>
						</div>
					</div>
					<div class="row col-md-12">
						<div class="form-group col-md-6">
							<label class="asterisk control-label col-md-3">Password</label>
							<input class="col-md-9" type="password" id="txt_password" maxlength="12" required>
							<label class="control-label col-md-3"></label>
							<p id="ppassword" class="col-md-9 hit">Enter combination of more than 6 numbers and alphabets. At least a capital letter and a small letter.</p>
						</div>
						<div class="form-group col-md-6">
							<label class="control-label col-md-3">Phone Number</label>
							<input class="col-md-9" type="text" id="txt_phonenum" maxlength="12">
							<label class="control-label col-md-3"></label>
							<p id="pphonenum" class="col-md-9 hit">Enter only number (0-9).</p>
						</div>
					</div>
				</div>
				<div class="row col-md-12">
					<div class="row col-md-6">
						<div class="row col-md-12">
							<h3 style='display:block; float: left;'>Cradit Settings</h3>
						</div>
						<div class="form-group col-md-12" id="currencySelect" style="display: block">
							<label class="asterisk control-label col-md-3">Currency</label>
							<select id="selectcurrency" class="col-md-9">
								<option value="THB">THB</option>
								<option value="CNY">CNY</option>
								<option value="MYR">MYR</option>
								<option value="USD">USD</option>
								<option value="JPY">JPY</option>
								<option value="IDR">IDR</option>
								<option value="HKD">HKD</option>
								<option value="KHR">KHR</option>
								<option value="PHP">PHP</option>
								<option value="LAK">LAK</option>
								<option value="VND">VND</option>
							</select>
						</div>
						<div class="form-group col-md-12">
							<label class="asterisk control-label col-md-3">Credit</label>
							<div class="col-md-9">
								<input type="text" class="form-control" id="txt_scorenum" maxlength="12">
								<span id="maxValue" class="badge bg-red" style="margin-left:8px;">MaxValue: `+ConvertInterface_A(agentDetail[`score`])+`</span>
								<p id="pscore" class="hit"></p>
							</div>
						</div>
						<div class="form-group col-md-12">
							<label class="asterisk control-label col-md-3">Shareholder type</label>
							<div class="btn-group col-md-9" data-toggle="buttons">
								<label class="btn btn-primary form-check-label waves-effect waves-light active">
									<input id="typeb2b" class="form-check-input" type="radio" checked="" autocomplete="off"> B2B
								</label>
								<label class="btn btn-primary form-check-label waves-effect waves-light">
									<input id="typeb2c" class="form-check-input" type="radio" autocomplete="off"> B2C
								</label>
							</div>
						</div>
						<div class="form-group col-md-12" id="ourPTSelect" style="display: block">
							<label class="asterisk control-label col-md-3">Our PT</label>
							<select id="selectourpt" class="col-md-9">
								<option value="100">100%</option>
								<option value="90">90%</option>
								<option value="80">80%</option>
								<option value="70">70%</option>
								<option value="60">60%</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="givenPTSelect" style="display: block">
							<label class="asterisk control-label col-md-3">Given PT</label>
							<select id="selectgivenpt" class="col-md-9">
								<option value="0">0%</option>
								<option value="10">10%</option>
								<option value="20">20%</option>
								<option value="30">30%</option>
								<option value="40">40%</option>
							</select>
						</div>
					</div>
					<div class="row col-md-6">
						<div class="row col-md-12">
							<h3 style='display:block; float: left;'>Commission Setting</h3>
						</div>
						<div class="form-group col-md-12" id="originalBac" style="display: block">
							<label class="asterisk control-label col-md-4">Original Baccarat</label>
							<select id="selectorgbac" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="super6Bac" style="display: block">
							<label class="asterisk control-label col-md-4">Super6 Baccarat</label>
							<select id="selectsupbac" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="superBac4" style="display: block">
							<label class="asterisk control-label col-md-4">Baccarat4 Point</label>
							<select id="selectbac4" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="cowCow" style="display: block">
							<label class="asterisk control-label col-md-4">Cow Cow</label>
							<select id="selectcowcow" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="dragonTiger" style="display: block">
							<label class="asterisk control-label col-md-4">Dragon Tiger</label>
							<select id="selectdragontiger" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="Roulette" style="display: block">
							<label class="asterisk control-label col-md-4">Roulette</label>
							<select id="selectroulette" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
						<div class="form-group col-md-12" id="sicbo" style="display: block">
							<label class="asterisk control-label col-md-4">Sicbo</label>
							<select id="selectsicbo" class="col-md-8">
								<option value="0">0.0</option>
								<option value="1">1.0</option>
								<option value="2">2.0</option>
								<option value="3">3.0</option>
								<option value="4">4.0</option>
								<option value="5">5.0</option>
								<option value="6">6.0</option>
								<option value="7">7.0</option>
								<option value="8">8.0</option>
								<option value="9">9.0</option>
								<option value="10">10.0</option>
							</select>
						</div>
					</div>
				</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary pull-right" id="Button_OK">Add Shareholder</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(document).ready(function () {
					$('#txt_username').on('focus focusout',function () {
						VaildAgentName();
						$('#txt_username').parent().addClass("has-success");
					});
					$('#txt_password').on('focus focusout',function () {
						VaildPassword();
						$('#txt_password').parent().addClass("has-success");
					});
					$('#txt_phonenum').on('focus focusout', function () {
						VaildPhoneNum();
						$('#txt_phonenum').parent().addClass("has-success");
					});
				})

				$("#Button_OK").on('click', function (e) {
					if (!VaildAgentName() || !VaildPassword() || !VaildPhoneNum()) {
						e.preventDefault();
					} else {
						var data = {
							username: $('#txt_username').val(),
							nickname: $('#txt_nickname').val(),
							password: $('#txt_password').val(),
							phonenum: $('#txt_phonenum').val(),

							currency: getSelectedValue('selectcurrency'),
							credit: $('#txt_scorenum').val(),
							shtype: getSelectedShareholderType(),
							ourpt: getSelectedValue('selectourpt'),
							givenpt: getSelectedValue('selectgivenpt'),

							commorgbac: getSelectedValue('selectorgbac'),
							commsupbac: getSelectedValue('selectsupbac'),
							commbac4: getSelectedValue('selectbac4'),
							commcowcow: getSelectedValue('selectcowcow'),
							commdragon: getSelectedValue('selectdragontiger'),
							commroulet: getSelectedValue('selectroulette'),
							commsicbo: getSelectedValue('selectsicbo'),
						};
			
						$.pjax({
							type: 'POST',
							url: this.value,
							data: data,
							container: '#pjax-container'
						});
						e.preventDefault();
					}
				})
				
				function getSelectedShareholderType() {
					var objSel = document.getElementById('typeb2b');
					if (objSel.checked) return 'b2b';

					objSel = document.getElementById('typeb2c');
					if (objSel.checked) return 'b2c';

					return '';
				}
				
				function getSelectedValue(elementId) {
					var objSel = document.getElementById(elementId);
					var optcnt = objSel.options.length;
					for (i = 0 ; i < optcnt; i++) {
						if (objSel.options[i].selected == true) {
							var selectedValue = objSel.options[i].value;
							return selectedValue;
						}
					}
			
					return "0";
				}

				function checkUserName(n) { return /^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{6,16})+$/.test(n) }
				function checkPassWord(n) { return /^(?=.*?[0-9])(?=.*?[A-Z])(?=.*?[a-z])[0-9A-Za-z!)-_]{6,15}$/.test(n) }
				function checkPhonenum(n) { return /^([0-9]{6,16})+$/.test(n) }

				function VaildAgentName() {
					if (!checkUserName($('#txt_username').val()))
					{
						AgentNameWarning();
						return false;
					}
					$('#txt_username').parent().removeClass("has-warning");
					$('#txt_username').parent().addClass("has-success");
			
					return true;
				}
				function AgentNameWarning() {
					$('#pusername').parent().addClass("has-warning");
					return false;
				}

				function VaildPassword() {
					if (!checkPassWord($('#txt_password').val())) {
						TipPassword();
						return false;
					}
					
					$('#ppassword').parent().removeClass("has-warning");
					$('#ppassword').parent().addClass("has-success");
					return true;
				}
				function TipPassword() {
					// $('#ppassword').text("Password with minimum 6 characters, must with combination of numbers and alphabets. At least a capital letter and a small letter.");
					$('#ppassword').parent().addClass("has-warning");
					return false;
				}

				function VaildPhoneNum() {
					if (!checkPhonenum($('#txt_phonenum').val()))
					{
						PhoneNumWarning();
						return false;
					}
					$('#txt_phonenum').parent().removeClass("has-warning");
					$('#txt_phonenum').parent().addClass("has-success");
			
					return true;
				}
				function PhoneNumWarning() {
					$('#txt_phonenum').parent().addClass("has-warning");
					return false;
				}
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h6 class="hidden-xs">
				Shareholder / Add Shareholder
			</h3>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Shareholder / Add Shareholder</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showAddShareHolder(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showAddShareHolder`)
	user := auth.Auth(ctx)
	param := guard.GetAddShareholderParam(ctx)

	checkExist, errExist := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", param.Username).
		First()

	if !db.CheckError(errExist, db.QUERY) && checkExist != nil {
		h.showAddShareHolderQueryBox(ctx, errors2.New("Username already exists as a shareholder!"))
		return
	}

	checkExist, errExist = db.WithDriver(h.conn).Table("SubAccounts").
		Where("accountname", "=", param.Username).
		First()

	if !db.CheckError(errExist, db.QUERY) && checkExist != nil {
		h.showAddShareHolderQueryBox(ctx, errors2.New("Username already exists as a subaccount!"))
		return
	}

	creditValue, creditError := strconv.ParseFloat(ConvertInterface_A(param.Credit), 64)
	if creditError != nil {
		h.showAddShareHolderQueryBox(ctx, errors2.New("Input correct value of Credit!"))
		return
	}

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		// alert += aAlert().Warning(err.Error())
		h.showAddShareHolderQueryBox(ctx, err)
		return
	}

	if agentDetail == nil {
		h.showAddShareHolderQueryBox(ctx, errors2.New(`Exception caused while get self information!`))
		return
	}

	parentLevel, _ := agentDetail[`level`].(int)
	scoreValue, _ := strconv.ParseFloat(ConvertInterface_A(agentDetail[`score`]), 64)

	// if I am a super user, no need to compare score
	if parentLevel > 0 {
		if scoreValue < 0 {
			h.showAddShareHolderQueryBox(ctx, errors2.New("Don't have enough Credit!"))
			return
		}
		if scoreValue < creditValue {
			h.showAddShareHolderQueryBox(ctx, errors2.New("Don't have enough Credit!"))
			return
		}

		scoreValue = scoreValue - creditValue

		_, updateError := db.WithDriver(h.conn).Table("Agents").
			Where("username", "=", user.UserName).
			Update(dialect.H{
				"score": scoreValue,
			})

		if !db.CheckError(updateError, db.QUERY) {
			updateError = nil
		}

		if updateError != nil {
			h.showAddShareHolderQueryBox(ctx, updateError)
			return
		}
	}

	// id,username,password,score,country,name,tel,description,parentid,level,noticereaddate
	description := "N/A"
	sadescription := "N/A"
	name := param.Nickname
	tel := param.Phonenum
	if param.Nickname == "" {
		name = "N/A"
	}
	if param.Phonenum == "" {
		tel = "N/A"
	}
	state := 1
	level := parentLevel + 1
	agentids := agentDetail[`agentids`]
	if agentids == nil {
		agentids = agentDetail[`id`].(int64)
	} else {
		agentids = ConvertInterface_A(agentids) + "," + agentDetail[`id`].(string)
	}

	passwordchangedate := time.Now().UTC().Format("2006-01-02 15:04:05")
	turnover := 0

	_, insertError := db.WithDriver(h.conn).Table("Agents").
		WithDriver(h.conn).
		Insert(dialect.H{
			"username":           param.Username,
			"password":           param.Password,
			"score":              creditValue,
			"name":               name,
			"tel":                tel,
			"description":        description,
			"state":              state,
			"parentid":           agentDetail[`id`],
			"level":              level,
			"agentids":           agentids,
			"passwordchangedate": passwordchangedate,
			"turnover":           turnover,
			"sadescription":      sadescription,
			"noticereaddate":     passwordchangedate,
		})

	if !db.CheckError(insertError, db.QUERY) {
		insertError = nil
	}

	if insertError != nil {
		h.showAddShareHolderQueryBox(ctx, insertError)
		return
	}

	beforescore := 0
	setscore := creditValue
	afterscore := creditValue
	account := user.UserName
	username := param.Username
	ip := ctx.LocalIP()
	datetime := time.Now()

	_, insertError = db.WithDriver(h.conn).Table("ScoreLogs").
		WithDriver(h.conn).
		Insert(dialect.H{
			"account":     account,
			"username":    username,
			"setscore":    setscore,
			"beforescore": beforescore,
			"afterscore":  afterscore,
			"ip":          ip,
			"datetime":    datetime,
		})

	if !db.CheckError(insertError, db.QUERY) {
		insertError = nil
	}

	if insertError != nil {
		h.showAddShareHolderQueryBox(ctx, insertError)
		return
	}

	_, insertError = db.WithDriver(h.conn).Table("SAOperationLogs").
		WithDriver(h.conn).
		Insert(dialect.H{
			"action":      "add new user",
			"account":     user.UserName,
			"username":    username,
			"ip":          ip,
			"datetime":    datetime,
			"description": user.UserName + ` --- 新增帐号：` + username + ` 并设定分数成功，分数为：` + param.Credit + ` 帐号类型：100`,
		})

	if !db.CheckError(insertError, db.QUERY) {
		insertError = nil
	}

	if insertError != nil {
		h.showAddShareHolderQueryBox(ctx, insertError)
		return
	}

	h.showAddShareHolderQueryBox(ctx, errors2.New("Successfully added new Shareholder"))
}

// added by jaison
func (h *Handler) ShowAddShareHolder(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowAddShareHolder`)
	h.showAddShareHolderQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) AddShareHolder(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/AddShareHolder`)
	param := guard.GetAddShareholderParam(ctx)

	if param.Username == "" {
		h.showAddShareHolderQueryBox(ctx, errors2.New("Input the Username!"))
		return
	}

	if param.Password == "" {
		h.showAddShareHolderQueryBox(ctx, errors2.New("Input the Password!"))
		return
	}
	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showAddShareHolderQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showAddShareHolder(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/management/addshareholder"))
		return
	}

	h.showAddShareHolder(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/management/addshareholder"))
}

// Shareholders
// added by jaison
type ShareHolderQueryParamInterface struct {
	Username string
	Level    string
}

func (h *Handler) showShareholdersQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showShareholdersQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Shareholder</h1>`)).
		SetBody(template.HTML(`
			<div class="box-body">
				<div class="row col-md-12">
					<div class="row col-md-4">
						<label class="text-blue col-md-3">Login name:</label>
						<input id="txt_username" class="col-md-9" type="text" maxlength="12" value="">
					</div>
					<div class="form-group col-md-4" id="levelSelect" style="display: block">
						<label class="text-blue col-md-3">Level:</label>
						<select id="selectLevel" class="col-md-9" name="select">
							<option value="0">All</option>
							<option value="1">1</option>
							<option value="2">2</option>
							<option value="3">3</option>
							<option value="4">4</option>
							<option value="5">5</option>
							<option value="6">6</option>
						</select>
					</div>
					<div class="row col-md-1">
						<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
					</div>
				</div>
			</div>`)).
		GetContent() + template.HTML(`
			<script>
				function getSelectedLevel() {
					var objSel = document.getElementById("selectLevel");
					var optcnt = objSel.options.length;
					for (i = 0 ; i < optcnt; i++) {
						if (objSel.options[i].selected == true) {
							var selectedValue = objSel.options[i].value;
							return selectedValue;
						}
					}
			
					return "0";
				}

				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
						level: getSelectedLevel(),
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
			<h6 class="hidden-xs">
				Member Management / Shareholder
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Member Management / Shareholder</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showShareholders(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showShareholders`)
	user := auth.Auth(ctx)
	param := guard.GetSearchShareholderParam(ctx)

	options := ``

	for i := 0; i < 6; i++ {
		selected := ``
		optionText := strconv.Itoa(i)

		if param.Level == strconv.Itoa(i) {
			selected = ` selected`
		}

		if i == 0 {
			optionText = `All`
		}

		options += `<option value="` + strconv.Itoa(i) + `"` + selected + `>` + optionText + `</option>`
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Shareholder</h1>`)).
		SetBody(template.HTML(`
			<div class="box-body">
				<div class="row col-md-12">
					<div class="row col-md-4">
						<label class="text-blue col-md-3">Login name:</label>
						<input id="txt_username" class="col-md-9" type="text" maxlength="12" value="`+param.Username+`">
					</div>
					<div class="form-group col-md-4" id="levelSelect" style="display: block">
						<label class="text-blue col-md-3">Level:</label>
						<select id="selectLevel" class="col-md-9" name="select">`+options+`</select>
					</div>
					<div class="row col-md-1">
						<button type="button" class="btn btn-primary" id="Button_OK">Search</button>
					</div>
				</div>
			</div>`)).
		GetContent() + template.HTML(`
			<script>
				function getSelectedLevel() {
					var objSel = document.getElementById("selectLevel");
					var optcnt = objSel.options.length;
					for (i = 0 ; i < optcnt; i++) {
						if (objSel.options[i].selected == true) {
							var selectedValue = objSel.options[i].value;
							return selectedValue;
						}
					}

					return "0";
				}

				$('#Button_OK').click(function (e) {
					var data = {
						username: $('#txt_username').val(),
						level: getSelectedLevel(),
					};

					let param = new Map();
					param.set('username', data.username);
					param.set('level', data.level);

					$.pjax({
						type: 'POST',
						url: addParameterToURL(param),
						data: data,
						container: '#pjax-container'
					});
					e.preventDefault();
				})
			</script>`)

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		// alert += aAlert().Warning(err.Error())
		h.showAddShareHolderQueryBox(ctx, err)
		return
	}

	if agentDetail == nil {
		h.showAddShareHolderQueryBox(ctx, errors2.New(`Exception caused while get super Agent information!`))
		return
	}

	panel := h.table("shareholderslist", ctx)

	if agentDetail["level"].(int64) > 0 {
		panel.GetInfo().Where("parentid", "=", agentDetail["parentid"])
	}

	if param.Username != "" {
		panel.GetInfo().Where("username", "=", param.Username)
	}
	if param.Level != "0" && param.Level != "" {
		panel.GetInfo().Where("level", "=", param.Level)
	}

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())

	// params.WithParameter("username", param.Username)
	// params.WithParameter("level", param.Level)

	// h.showTable(ctx, "shareholderslist", params, panel)
	panel, panelInfo, _, err := h.showTableData(ctx, "gameconfigs", params, panel, "/dailyagentreport/")
	paginator := panelInfo.Paginator
	paginator = paginator.SetHideEntriesInfo()
	url := paginator.GetUrl()
	// fmt.Println(url)
	// // url += "&" + "username=" + param.Username + "&level=" + param.Level

	// paginator.SetUrl(url)
	// paginator.SetNextUrl(url)
	// paginator.SetPreviousUrl(url)
	fmt.Println(url)
	// tableContent := h.getTableContent(ctx, "shareholderslist", params, panel)

	dataTable := aDataTable().
		SetInfoList(panelInfo.InfoList).
		SetLayout(panel.GetInfo().TableLayout).
		SetStyle(`hover table-bordered`).
		SetIsTab(true).
		SetHideThead(false).
		SetThead(panelInfo.Thead)

	dataTableDiv := aBox().
		SetTheme(`primary`).
		WithHeadBorder().
		SetStyle("display: block;").
		SetBody(template.HTML(`<div class="table-responsive">`) +
			dataTable.GetContent() +
			template.HTML(`</div>`)).
		SetFooter(paginator.GetContent()).
		GetContent()

	h.HTML(ctx, user, types.Panel{
		Content:     alert + queryBoxForm + dataTableDiv,
		Description: "",
		Title: template2.HTML(template.HTML(`
			<h6 class="hidden-xs">
				Member Management / Shareholder
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Member Management / Shareholder</li>
			</ol>`)),
	})
}

// added by jaison
func (h *Handler) ShowShareholders(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowShareholders`)
	// panel := h.table("shareholderslist", ctx)
	params := parameter.GetParam(ctx.Request.URL, 1)

	fmt.Println(ctx.Request.URL)
	fmt.Println(params)

	// if tempCache == nil {
	h.showShareholdersQueryBox(ctx, nil)
	// } else {
	// 	Username := tempCache.(ShareHolderQueryParamInterface).Username
	// 	Level := tempCache.(ShareHolderQueryParamInterface).Level
	// 	guard.SetSearchShareholderManualParam(ctx, Username, Level)

	// 	h.showShareholders(ctx, template2.HTML(``))
	// }
}

// added by jaison
func (h *Handler) Shareholders(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/Shareholders`)
	param := guard.GetSearchShareholderParam(ctx)

	// if param.Username == "" {
	// 	h.showShareholdersQueryBox(ctx, errors2.New("Input the Username!"))
	// 	return
	// }

	if param.Level == "" {
		h.showShareholdersQueryBox(ctx, errors2.New("Select Shareholder Level!"))
		return
	}

	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showShareholdersQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showShareholders(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/management/shareholders"))
		return
	}

	h.showShareholders(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/management/shareholders"))
}

// Sub Accounts
// Add Sub Account
// added by jaison
func (h *Handler) showAddSubAccountQueryBox(ctx *context.Context, err error) {
	fmt.Println(`plugins.admin.controller.menu.go/showAddSubAccountQueryBox`)

	user := auth.Auth(ctx)

	var alert template2.HTML

	if err != nil {
		alert = aAlert().Warning(err.Error())
	}

	queryBoxForm := aBox().
		SetTheme(`default`).
		SetStyle("display: block;").
		WithHeadBorder().
		SetHeader(template.HTML(`<h1 class="box-title text-bold text-muted" id="d_tip_2">Add Sub Account</h1>`)).
		SetBody(template.HTML(`
			<div class="box-body">
				<div class="row col-md-12">
					<div class="row col-md-12">
						<h3 style='display:block; float: left;'>Basic Info</h3>
					</div>
					<div class="row col-md-12">
						<div class="form-group col-md-6">
							<label class="asterisk control-label col-md-3">Username</label>
							<label class="control-label col-md-3">`+user.UserName+`@</label>
							<input class="col-md-6" type="text" id="txt_username" maxlength="12" required>
							<label class="control-label col-md-3"></label>
							<p id="pusername" class="col-md-9 hit">Enter only number (0-9) or letter (A-Z, a-z).</p>
						</div>
						<div class="form-group col-md-6">
							<label class="control-label col-md-3">Nickname</label>
							<input class="col-md-9" type="text" id="txt_nickname" maxlength="12">
							<label class="control-label col-md-3"></label>
							<p id="pnickname" class="col-md-9 hit">Enter only number (0-9) or letter (A-Z, a-z).</p>
						</div>
					</div>
					<div class="row col-md-12">
						<div class="form-group col-md-6">
							<label class="asterisk control-label col-md-3">Password</label>
							<input class="col-md-9" type="password" id="txt_password" maxlength="12" required>
							<label class="control-label col-md-3"></label>
							<p id="ppassword" class="col-md-9 hit">Enter combination of more than 6 numbers and alphabets. At least a capital letter and a small letter.</p>
						</div>
						<div class="form-group col-md-6">
							<label class="control-label col-md-3">Phone Number</label>
							<input class="col-md-9" type="text" id="txt_phonenum" maxlength="12">
							<label class="control-label col-md-3"></label>
							<p id="pphonenum" class="col-md-9 hit">Enter only number (0-9).</p>
						</div>
					</div>
				</div>
				<div class="row col-md-6">
					<h3 class="row" style='display:block; float: left;'>Permissions</h3>
					
					<div class="row col-md-12" style="display: block">
						<label class="control-label col-md-6">Account</label>
						<form class="row col-md-6">
							<input type="radio" id="accoff" name="account" cvalue="accoff">
							<label for="accoff">Off</label>
							<input type="radio" id="accview" name="account" value="accview" checked>
							<label for="accview">View</label>
							<input type="radio" id="accedit" name="account" value="accedit">
							<label for="accedit">Edit</label>
						</form>
					</div>
					
					<div class="row col-md-12" style="display: block">
						<label class="control-label col-md-6">Member Management</label>
						<form class="row col-md-6">
							<input type="radio" id="memoff" name="memman" value="memoff">
							<label for="memoff">Off</label>
							<input type="radio" id="memview" name="memman" value="memview" checked>
							<label for="memview">View</label>
							<input type="radio" id="memedit" name="memman" value="memedit">
							<label for="memedit">Edit</label>
						</form>
					</div>
					
					<div class="row col-md-12" style="display: block">
						<label class="control-label col-md-6">Stock Management</label>
						<form class="row col-md-6">
							<input type="radio" id="stockoff" name="stockman" value="stockoff">
							<label for="stockoff">Off</label>
							<input type="radio" id="stockview" name="stockman" value="stockview" checked>
							<label for="stockview">View</label>
							<input type="radio" id="stockedit" name="stockman" value="stockedit">
							<label for="stockedit">Edit</label>
						</form>
					</div>
					
					<div class="row col-md-12" style="display: block">
						<label class="control-label col-md-6">Report</label>
						<form class="row col-md-6">
							<input type="radio" id="reportoff" name="report" value="reportoff">
							<label for="reportoff">Off</label>
							<input type="radio" id="reportview" name="report" value="reportview" checked>
							<label for="reportview">View</label>
							<input type="radio" id="reportedit" name="report" value="reportedit">
							<label for="reportedit">Edit</label>
						</form>
					</div>
					
					<div class="row col-md-12" style="display: block">
						<label class="control-label col-md-6">Payment</label>
						<form class="row col-md-6">
							<input type="radio" id="paymentoff" name="payment" value="paymentoff">
							<label for="paymentoff">Off</label>
							<input type="radio" id="paymentview" name="payment" value="paymentview" checked>
							<label for="paymentview">View</label>
							<input type="radio" id="paymentedit" name="payment" value="paymentedit">
							<label for="paymentedit">Edit</label>
						</form>
					</div>
				</div>
			</div>`)).
		SetFooter(template.HTML(`<button type="button" class="btn btn-primary pull-right" id="Button_OK">Add Sub Account</button>`)).
		GetContent() + template.HTML(`
			<script>
				$(document).ready(function () {
					$('#txt_username').on('focus focusout',function () {
						ValidSubAccName();
						$('#txt_username').parent().addClass("has-success");
					});
					$('#txt_password').on('focus focusout',function () {
						VaildPassword();
						$('#txt_password').parent().addClass("has-success");
					});
					$('#txt_phonenum').on('focus focusout', function () {
						VaildPhoneNum();
						$('#txt_phonenum').parent().addClass("has-success");
					});
				})
				
				function getSelectedAccPerm() {
					var objSel = document.getElementById('accoff');
					if (objSel.checked) return '0';

					objSel = document.getElementById('accview');
					if (objSel.checked) return '1';

					return '2';
				}
				
				function getSelectedMemPerm() {
					var objSel = document.getElementById('memoff');
					if (objSel.checked) return '0';

					objSel = document.getElementById('memview');
					if (objSel.checked) return '1';

					return '2';
				}
				
				function getSelectedStockPerm() {
					var objSel = document.getElementById('stockoff');
					if (objSel.checked) return '0';

					objSel = document.getElementById('stockview');
					if (objSel.checked) return '1';

					return '2';
				}
				
				function getSelectedReportPerm() {
					var objSel = document.getElementById('reportoff');
					if (objSel.checked) return '0';

					objSel = document.getElementById('reportview');
					if (objSel.checked) return '1';

					return '2';
				}
				
				function getSelectedPaymentPerm() {
					var objSel = document.getElementById('paymentoff');
					if (objSel.checked) return '0';

					objSel = document.getElementById('paymentview');
					if (objSel.checked) return '1';

					return '2';
				}

				$("#Button_OK").on('click', function (e) {
					if (!ValidSubAccName() || !VaildPassword() || !VaildPhoneNum()) {
						e.preventDefault();
					} else {
						var data = {
							username: $('#txt_username').val(),
							nickname: $('#txt_nickname').val(),
							password: $('#txt_password').val(),
							phonenum: $('#txt_phonenum').val(),
							
							permaccount: getSelectedAccPerm(),
							permmem: getSelectedMemPerm(),
							permstock: getSelectedStockPerm(),
							permreport: getSelectedReportPerm(),
							permpayment: getSelectedPaymentPerm(),
						};
			
						$.pjax({
							type: 'POST',
							url: this.value,
							data: data,
							container: '#pjax-container'
						});
						e.preventDefault();
					}
				})

				function checkUserName(n) { return /^([a-zA-Z0-9]{1}[a-zA-Z0-9_-]{6,16})+$/.test(n) }
				function checkPassWord(n) { return /^(?=.*?[0-9])(?=.*?[A-Z])(?=.*?[a-z])[0-9A-Za-z!)-_]{6,15}$/.test(n) }
				function checkPhonenum(n) { return /^([0-9]{6,16})+$/.test(n) }

				function ValidSubAccName() {
					if (!checkUserName($('#txt_username').val()))
					{
						SubAccNameWarning();
						return false;
					}
					$('#txt_username').parent().removeClass("has-warning");
					$('#txt_username').parent().addClass("has-success");
			
					return true;
				}
				function SubAccNameWarning() {
					// $('#pusername').text("byte length: 7-16 byte.");
					$('#pusername').parent().addClass("has-warning");
					return false;
				}

				function VaildPassword() {
					if (!checkPassWord($('#txt_password').val())) {
						TipPassword();
						return false;
					}
					
					$('#ppassword').parent().removeClass("has-warning");
					$('#ppassword').parent().addClass("has-success");
					return true;
				}
				function TipPassword() {
					// $('#ppassword').text("Password with minimum 6 characters, must with combination of numbers and alphabets. At least a capital letter and a small letter.");
					$('#ppassword').parent().addClass("has-warning");
					return false;
				}

				function VaildPhoneNum() {
					if (!checkPhonenum($('#txt_phonenum').val()))
					{
						PhoneNumWarning();
						return false;
					}
					$('#txt_phonenum').parent().removeClass("has-warning");
					$('#txt_phonenum').parent().addClass("has-success");
			
					return true;
				}
				function PhoneNumWarning() {
					$('#txt_phonenum').parent().addClass("has-warning");
					return false;
				}
			</script>`)

	h.HTML(ctx, user, types.Panel{
		Title: template2.HTML(template.HTML(`
			<h6 class="hidden-xs">
				Sub Accounts / Add Sub Account
			</h6>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Sub Accounts / Add Sub Account</li>
			</ol>`)),
		Description: template.HTML(``),
		Content:     alert + queryBoxForm,
	})
}

// added by jaison
func (h *Handler) showAddSubAccount(ctx *context.Context, alert template2.HTML) {
	fmt.Println(`plugins.admin.controller.menu.go/showAddSubAccount`)
	user := auth.Auth(ctx)
	param := guard.GetAddSubAccountParam(ctx)

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		// alert += aAlert().Warning(err.Error())
		h.showAddSubAccountQueryBox(ctx, err)
		return
	}

	if agentDetail == nil {
		h.showAddSubAccountQueryBox(ctx, errors2.New(`Exception caused while get self information!`))
		return
	}

	checkExist, errExist := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", param.Username).
		First()

	if !db.CheckError(errExist, db.QUERY) && checkExist != nil {
		h.showAddSubAccountQueryBox(ctx, errors2.New("Username already exists as a shareholder!"))
		return
	}

	checkExist, errExist = db.WithDriver(h.conn).Table("SubAccounts").
		Where("accountname", "=", user.UserName+"@"+param.Username).
		First()

	if !db.CheckError(errExist, db.QUERY) && checkExist != nil {
		h.showAddSubAccountQueryBox(ctx, errors2.New("Username already exists as a subaccount!"))
		return
	}

	subAccDetails, errSubAccExist := db.WithDriver(h.conn).Table("SubAccounts").
		Where("agentid", "=", agentDetail["id"]).
		All()

	if db.CheckError(errSubAccExist, db.QUERY) {
		h.showAddSubAccountQueryBox(ctx, errSubAccExist)
		return
	}

	if subAccDetails == nil {
		h.showAddSubAccountQueryBox(ctx, errors2.New(`Exception caused while get sub account information!`))
		return
	}

	if len(subAccDetails) >= 5 {
		h.showAddSubAccountQueryBox(ctx, errors2.New(`Just can add upto 5 sub accounts!`))
		return
	}

	userName := user.UserName + "@" + param.Username
	password := param.Password
	nickname := param.Nickname
	tel := param.Phonenum
	description := "N/A"

	permission := []string{param.Account, param.MemberManagement, param.StockManagement, param.Report, param.Payment}
	binPermission := strings.Join(permission, "")
	dbPermission, permErr := strconv.ParseInt(binPermission, 3, 64)

	if permErr != nil {
		fmt.Println(permErr)
	}

	fmt.Println(dbPermission)

	insertDate := time.Now().UTC().Format("2006-01-02 15:04:05")

	_, insertError := db.WithDriver(h.conn).Table("SubAccounts").
		WithDriver(h.conn).
		Insert(dialect.H{
			"accountname": userName,
			"description": description,
			"password":    password,
			"level":       dbPermission,
			"agentid":     agentDetail["id"],
			"state":       1,
			"nickname":    nickname,
			"tel":         tel,
			"datetime":    insertDate,
		})

	if !db.CheckError(insertError, db.QUERY) {
		insertError = nil
	}

	if insertError != nil {
		h.showAddSubAccountQueryBox(ctx, insertError)
		return
	}

	h.showAddSubAccountQueryBox(ctx, errors2.New("Successfully added new Shareholder"))
}

// added by jaison
func (h *Handler) ShowAddSubAccount(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowAddSubAccount`)
	h.showAddSubAccountQueryBox(ctx, nil)
}

// added by jaison
func (h *Handler) AddSubAccount(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/AddSubAccount`)
	param := guard.GetAddSubAccountParam(ctx)

	if param.Username == "" {
		h.showAddSubAccountQueryBox(ctx, errors2.New("Input the Username!"))
		return
	}

	if param.Password == "" {
		h.showAddSubAccountQueryBox(ctx, errors2.New("Input the Password!"))
		return
	}
	// need to check again
	// if (param.Username == HttpContext.Session["superagentname"].ToString())
	// 	h.showAddSubAccountQueryBox(ctx, errors2.New("No Such User!"))

	if param.HasAlert() {
		h.showAddSubAccount(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/management/addsubaccount"))
		return
	}

	h.showAddSubAccount(ctx, template.HTML(``))
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePath("/management/addsubaccount"))
}

// Sub Accounts
// added by jaison
func (h *Handler) ShowSubAccounts(ctx *context.Context) {
	fmt.Println(`plugins.admin.controller.menu.go/ShowSubAccounts`)
	user := auth.Auth(ctx)

	agentDetail, err := db.WithDriver(h.conn).Table("Agents").
		Where("username", "=", user.UserName).
		First()

	if db.CheckError(err, db.QUERY) {
		// alert += aAlert().Warning(err.Error())
		h.showAddSubAccountQueryBox(ctx, err)
		return
	}

	if agentDetail == nil {
		h.showAddSubAccountQueryBox(ctx, errors2.New(`Exception caused while get self information!`))
		return
	}

	panel := h.table("subaccountslist", ctx)

	panel.GetInfo().Where("agentid", "=", agentDetail["id"])

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField, panel.GetInfo().GetSort())
	tableContent := h.getTableContent(ctx, "subaccountslist", params, panel)

	h.HTML(ctx, user, types.Panel{
		Content:     tableContent,
		Description: "",
		Title: template2.HTML(template.HTML(`
			<h3 class="hidden-xs">
				Sub Accounts
			</h3>
			<ol class="breadcrumb hidden-md hidden-lg hidden-sm">
				<li>Sub Accounts</li>
			</ol>`)),
	})
}
