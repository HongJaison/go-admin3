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
type AddShareholderParam struct {
	// basic info
	Username string
	Nickname string
	Password string
	Phonenum string

	// credit settings
	Currency string
	Credit   string
	SHType   string
	OurPT    string
	GivenPT  string

	// commission
	CommOrgBac string
	CommSupBac string
	CommBac4   string
	CommCowCow string
	CommDragon string
	CommRoulet string
	CommSicbo  string

	Alert template.HTML
}

// added by jaison
func (g *Guard) SetAddShareholderParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetAddShareholderParam`)

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	fmt.Println(ctx)

	// basic info
	Username := ctx.FormValue("username")
	Nickname := ctx.FormValue("nickname")
	Password := ctx.FormValue("password")
	Phonenum := ctx.FormValue("phonenum")

	// credit settings
	Currency := ctx.FormValue("currency")
	Credit := ctx.FormValue("credit")
	SHType := ctx.FormValue("shtype")
	OurPT := ctx.FormValue("ourpt")
	GivenPT := ctx.FormValue("givenpt")

	// commission
	CommOrgBac := ctx.FormValue("commorgbac")
	CommSupBac := ctx.FormValue("commsupbac")
	CommBac4 := ctx.FormValue("commbac4")
	CommCowCow := ctx.FormValue("commcowcow")
	CommDragon := ctx.FormValue("commdragon")
	CommRoulet := ctx.FormValue("commroulet")
	CommSicbo := ctx.FormValue("commsicbo")

	ctx.SetUserValue("AddShareholderParam", &AddShareholderParam{
		// basic info
		Username: Username,
		Nickname: Nickname,
		Password: Password,
		Phonenum: Phonenum,

		// credit settings
		Currency: Currency,
		Credit:   Credit,
		SHType:   SHType,
		OurPT:    OurPT,
		GivenPT:  GivenPT,

		// commission
		CommOrgBac: CommOrgBac,
		CommSupBac: CommSupBac,
		CommBac4:   CommBac4,
		CommCowCow: CommCowCow,
		CommDragon: CommDragon,
		CommRoulet: CommRoulet,
		CommSicbo:  CommSicbo,

		Alert: alert,
	})
	ctx.Next()
}

// added by jaison
func GetAddShareholderParam(ctx *context.Context) *AddShareholderParam {
	return ctx.UserValue["AddShareholderParam"].(*AddShareholderParam)
}

// added by jaison
func (e AddShareholderParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
type AddSubAccountParam struct {
	// basic info
	Username string
	Nickname string
	Password string
	Phonenum string

	// permission
	Account          string
	MemberManagement string
	StockManagement  string
	Report           string
	Payment          string

	Alert template.HTML
}

// added by jaison
func (g *Guard) SetAddSubAccountParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetAddSubAccountParam`)

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	// basic info
	Username := ctx.FormValue("username")
	Nickname := ctx.FormValue("nickname")
	Password := ctx.FormValue("password")
	Phonenum := ctx.FormValue("phonenum")

	// permission
	Account := ctx.FormValue("permaccount")
	MemberManagement := ctx.FormValue("permmem")
	StockManagement := ctx.FormValue("permstock")
	Report := ctx.FormValue("permreport")
	Payment := ctx.FormValue("permpayment")

	ctx.SetUserValue("AddSubAccountParam", &AddSubAccountParam{
		// basic info
		Username: Username,
		Nickname: Nickname,
		Password: Password,
		Phonenum: Phonenum,

		// permission
		Account:          Account,
		MemberManagement: MemberManagement,
		StockManagement:  StockManagement,
		Report:           Report,
		Payment:          Payment,

		Alert: alert,
	})
	ctx.Next()
}

// added by jaison
func GetAddSubAccountParam(ctx *context.Context) *AddSubAccountParam {
	return ctx.UserValue["AddSubAccountParam"].(*AddSubAccountParam)
}

// added by jaison
func (e AddSubAccountParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
type SearchShareholderParam struct {
	Username string
	Level    string

	Alert template.HTML
}

// added by jaison
func (g *Guard) SetSearchShareholderParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetSearchShareholderParam`)

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	// basic info
	Username := ctx.FormValue("username")
	Level := ctx.FormValue("level")

	ctx.SetUserValue("SearchShareholderParam", &SearchShareholderParam{
		Username: Username,
		Level:    Level,
		Alert:    alert,
	})

	ctx.Next()
}

// added by jaison
func SetSearchShareholderManualParam(ctx *context.Context, Username string, Level string) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetSearchShareholderManualParam`)

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("SearchShareholderParam", &SearchShareholderParam{
		Username: Username,
		Level:    Level,
		Alert:    alert,
	})
}

// added by jaison
func GetSearchShareholderParam(ctx *context.Context) *SearchShareholderParam {
	if ctx.UserValue["SearchShareholderParam"] == nil {
		return nil
	}

	return ctx.UserValue["SearchShareholderParam"].(*SearchShareholderParam)
}

// added by jaison
func (e SearchShareholderParam) HasAlert() bool {
	return e.Alert != template.HTML("")
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
func (g *Guard) SetOutstandingParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetOutstandingParam`)

	username := ctx.FormValue("username")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("MemberOutstandingParamm", &LoginLogParam{
		Username: username,
		Alert:    alert,
	})
	ctx.Next()
}

// added by jaison
func GetOutstandingParam(ctx *context.Context) *LoginLogParam {
	return ctx.UserValue["MemberOutstandingParamm"].(*LoginLogParam)
}

// added by jaison
func (g *Guard) SetGameLogsParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetGameLogsParam`)

	username := ctx.FormValue("username")
	startdate := ctx.FormValue("startdate")
	enddate := ctx.FormValue("enddate")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("GameLogsParam", &ScoreLogParam{
		StartDate: startdate,
		EndDate:   enddate,
		Username:  username,
		Alert:     alert,
	})
	ctx.Next()
}

// added by jaison
func GetGameLogsParam(ctx *context.Context) *ScoreLogParam {
	return ctx.UserValue["GameLogsParam"].(*ScoreLogParam)
}

// added by jaison
type ScoreLogParam struct {
	StartDate string
	EndDate   string
	Username  string
	Alert     template.HTML
}

// added by jaison
func (g *Guard) SetScoreLogParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetScoreLogParam`)

	username := ctx.FormValue("username")
	startdate := ctx.FormValue("startdate")
	enddate := ctx.FormValue("enddate")

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	ctx.SetUserValue("ScoreLogParam", &ScoreLogParam{
		StartDate: startdate,
		EndDate:   enddate,
		Username:  username,
		Alert:     alert,
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

// added by jaison
type UpdateProfileRequestModel struct {
	PasswordOld string
	PasswordNew string
	PasswordCon string
	Alert       template.HTML
}

// added by jaison
func (g *Guard) SetEditProfileParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetEditProfileParam`)

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	var requestModel UpdateProfileRequestModel

	requestModel = UpdateProfileRequestModel{
		PasswordOld: ctx.FormValue("oldPassWd"),
		PasswordNew: ctx.FormValue("newPassWd"),
		PasswordCon: ctx.FormValue("rePassWd"),
		Alert:       alert,
	}

	ctx.SetUserValue("UpdateProfileParam", &requestModel)
	ctx.Next()
}

// added by jaison
func GetEditProfileParam(ctx *context.Context) *UpdateProfileRequestModel {
	if ctx.UserValue["UpdateProfileParam"] == nil {
		return nil
	}

	return ctx.UserValue["UpdateProfileParam"].(*UpdateProfileRequestModel)
}

// added by jaison
func (e UpdateProfileRequestModel) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// added by jaison
func (g *Guard) SetRedPacketLogParam(ctx *context.Context) {
	fmt.Println(`plugins.admin.modules.guard.menu_new.go/SetRedPacketLogParam`)

	var (
		alert template.HTML
	)

	alert = template.HTML(``)

	username := ctx.FormValue("username")
	startdate := ctx.FormValue("startdate")

	ctx.SetUserValue("RedPacketLogParam", &BonusLogParam{
		StartDate: startdate,
		Username:  username,
		Alert:     alert,
	})
	ctx.Next()
}

// added by jaison
func GetRedPacketLogParam(ctx *context.Context) *BonusLogParam {
	return ctx.UserValue["RedPacketLogParam"].(*BonusLogParam)
}
