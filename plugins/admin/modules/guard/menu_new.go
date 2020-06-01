package guard

import (
	"fmt"
	"html/template"
	"strconv"

	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/auth"
	"github.com/HongJaison/go-admin3/modules/errors"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/form"
)

type MenuNewParam struct {
	Title    string
	Header   string
	ParentId int64
	Icon     string
	Uri      string
	Roles    []string
	Alert    template.HTML
}

func (e MenuNewParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

func (g *Guard) MenuNew(ctx *context.Context) {

	parentId := ctx.FormValue("parent_id")
	if parentId == "" {
		parentId = "0"
	}

	var (
		alert template.HTML
		token = ctx.FormValue(form.TokenKey)
	)

	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert = getAlert(errors.EditFailWrongToken)
	}

	if alert == "" {
		alert = checkEmpty(ctx, "title", "icon")
	}

	parentIdInt, _ := strconv.Atoi(parentId)

	ctx.SetUserValue(newMenuParamKey, &MenuNewParam{
		Title:    ctx.FormValue("title"),
		Header:   ctx.FormValue("header"),
		ParentId: int64(parentIdInt),
		Icon:     ctx.FormValue("icon"),
		Uri:      ctx.FormValue("uri"),
		Roles:    ctx.Request.Form["roles[]"],
		Alert:    alert,
	})
	ctx.Next()
}

func GetMenuNewParam(ctx *context.Context) *MenuNewParam {
	return ctx.UserValue[newMenuParamKey].(*MenuNewParam)
}

// added by jaison
type LoginLogParam struct {
	Username string
	Alert    template.HTML
}

// added by jaison
func (g *Guard) SetLoginLogParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetLoginLogParam`)

	username := ctx.FormValue("username")

	var (
		alert template.HTML
	// 	token = ctx.FormValue(form.TokenKey)
	)

	alert = template.HTML(``)
	// if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
	// 	alert = getAlert(errors.EditFailWrongToken)
	// }

	// if alert == "" {
	// 	alert = checkEmpty(ctx, "title", "icon")
	// }

	ctx.SetUserValue("LoginLogParam", &LoginLogParam{
		Username: username,
		Alert:    alert,
	})
	ctx.Next()
}

// added by jaison
func GetLoginLogParam(ctx *context.Context) *LoginLogParam {
	return ctx.UserValue["LoginLogParam"].(*LoginLogParam)
}

// added by jaison
func (e LoginLogParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
func (g *Guard) SetSearchPlayerParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetSearchPlayerParam`)

	username := ctx.FormValue("username")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("SearchPlayerParam", &LoginLogParam{
		Username: username,
		Alert:    alert,
	})
	ctx.Next()
}

// added by jaison
func GetSearchPlayerParam(ctx *context.Context) *LoginLogParam {
	return ctx.UserValue["SearchPlayerParam"].(*LoginLogParam)
}

// added by jaison
type SearchWinPlayersParam struct {
	DiffTime string
	Alert    template.HTML
}

// added by jaison
func (g *Guard) SetSearchWinPlayersParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetSearchWinPlayersParam`)

	difftime := ctx.FormValue("diffTime")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("SearchWinPlayersParam", &SearchWinPlayersParam{
		DiffTime: difftime,
		Alert:    alert,
	})
	ctx.Next()
}

// added by jaison
func GetSearchWinPlayersParam(ctx *context.Context) *SearchWinPlayersParam {
	return ctx.UserValue["SearchWinPlayersParam"].(*SearchWinPlayersParam)
}

// added by jaison
type ScoreLogParam struct {
	StartDateTime string
	EndDateTime   string
	Username      string
	Alert         template.HTML
}

// added by jaison
func (g *Guard) SetScoreLogParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetScoreLogParam`)

	username := ctx.FormValue("username")
	startdatetime := ctx.FormValue("startdate")
	enddatetime := ctx.FormValue("enddate")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("ScoreLogParam", &ScoreLogParam{
		StartDateTime: startdatetime,
		EndDateTime:   enddatetime,
		Username:      username,
		Alert:         alert,
	})
	ctx.Next()
}

// added by jaison
func GetScoreLogParam(ctx *context.Context) *ScoreLogParam {
	return ctx.UserValue["ScoreLogParam"].(*ScoreLogParam)
}

// added by jaison
func (e ScoreLogParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
type BonusLogParam struct {
	StartDate string
	Username  string
	Alert     template.HTML
}

// added by jaison
func (g *Guard) SetBonusLogParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetBonusLogParam`)

	username := ctx.FormValue("username")
	startdate := ctx.FormValue("startdate")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("BonusLogParam", &BonusLogParam{
		StartDate: startdate,
		Username:  username,
		Alert:     alert,
	})
	ctx.Next()
}

// added by jaison
func GetBonusLogParam(ctx *context.Context) *BonusLogParam {
	return ctx.UserValue["BonusLogParam"].(*BonusLogParam)
}

// added by jaison
func (e BonusLogParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
type ReportLogParam struct {
	StartDate string
	EndDate   string
	Alert     template.HTML
}

// added by jaison
func (g *Guard) SetReportLogParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetReportLogParam`)

	startdate := ctx.FormValue("startdate")
	enddate := ctx.FormValue("enddate")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("ReportLogParam", &ReportLogParam{
		StartDate: startdate,
		EndDate:   enddate,
		Alert:     alert,
	})
	ctx.Next()
}

// added by jaison
func GetReportLogParam(ctx *context.Context) *ReportLogParam {
	return ctx.UserValue["ReportLogParam"].(*ReportLogParam)
}

// added by jaison
func (e ReportLogParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
type UpdateConfigRequestModel struct {
	ProcessMode int
	Id          int
	Percent     float64
	CheckState  bool
	Alert       template.HTML
}

// added by jaison
func (g *Guard) SetConfigUpdateParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetConfigUpdateParam`)

	var (
		alert            template.HTML
		err1, err2, err3 error
		id               int
		percent          float64
		checkState       bool
	)

	alert = template.HTML(``)

	var requestModel UpdateConfigRequestModel

	requestType := ctx.FormValue("processType")

	if requestType == "1" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		percent, err2 = strconv.ParseFloat(ctx.FormValue("percent"), 64)

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 1,
			Id:          id,
			Percent:     percent,
			CheckState:  false,
			Alert:       alert,
		}
	}

	if requestType == "2" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		percent, err2 = strconv.ParseFloat(ctx.FormValue("percent"), 64)

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 2,
			Id:          id,
			Percent:     percent,
			CheckState:  false,
			Alert:       alert,
		}
	}

	if requestType == "3" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		percent, err2 = strconv.ParseFloat(ctx.FormValue("percent"), 64)
		checkState, err3 = strconv.ParseBool(ctx.FormValue("checkState"))

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 3,
			Id:          id,
			Percent:     percent,
			CheckState:  checkState,
			Alert:       alert,
		}
	}

	if requestType == "4" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		percent, err2 = strconv.ParseFloat(ctx.FormValue("percent"), 64)
		checkState, err3 = strconv.ParseBool(ctx.FormValue("checkState"))

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 4,
			Id:          id,
			Percent:     percent,
			CheckState:  checkState,
			Alert:       alert,
		}
	}

	if requestType == "5" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		percent, err2 = strconv.ParseFloat(ctx.FormValue("percent"), 64)
		checkState, err3 = strconv.ParseBool(ctx.FormValue("checkState"))

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 5,
			Id:          id,
			Percent:     percent,
			CheckState:  checkState,
			Alert:       alert,
		}
	}

	if requestType == "6" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		checkState, err3 = strconv.ParseBool(ctx.FormValue("checkState"))

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 6,
			Id:          id,
			Percent:     0,
			CheckState:  checkState,
			Alert:       alert,
		}
	}

	if requestType == "7" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		checkState, err3 = strconv.ParseBool(ctx.FormValue("checkState"))

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 7,
			Id:          id,
			Percent:     0,
			CheckState:  checkState,
			Alert:       alert,
		}
	}

	if requestType == "8" {
		id, err1 = strconv.Atoi(ctx.FormValue("id"))
		checkState = ctx.FormValue("state") == "1"

		requestModel = UpdateConfigRequestModel{
			ProcessMode: 8,
			Id:          id,
			Percent:     0,
			CheckState:  checkState,
			Alert:       alert,
		}
	}

	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Println(err1)
		fmt.Println(err2)
		fmt.Println(err3)
	}

	fmt.Println(requestModel)

	ctx.SetUserValue("UpdateConfigParam", &requestModel)
	ctx.Next()
}

// added by jaison
func GetConfigUpdateParam(ctx *context.Context) *UpdateConfigRequestModel {
	if ctx.UserValue["UpdateConfigParam"] == nil {
		return nil
	}

	return ctx.UserValue["UpdateConfigParam"].(*UpdateConfigRequestModel)
}

// added by jaison
func (e UpdateConfigRequestModel) HasAlert() bool {
	return e.Alert != template.HTML("")
}
