package admin

import (
	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/auth"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/response"
	"github.com/HongJaison/go-admin3/template"
)

// initRouter initialize the router and return the context.
func (admin *Admin) initRouter() *Admin {
	app := context.NewApp()

	route := app.Group(config.Prefix(), admin.globalErrorHandler)

	// auth
	route.GET(config.GetLoginUrl(), admin.handler.ShowLogin)
	route.POST("/signin", admin.handler.Auth)

	// auto install
	route.GET("/install", admin.handler.ShowInstall)
	route.POST("/install/database/check", admin.handler.CheckDatabase)

	for _, path := range template.Get(config.GetTheme()).GetAssetList() {
		route.GET("/assets"+path, admin.handler.Assets)
	}

	for _, path := range template.GetComponentAsset() {
		route.GET("/assets"+path, admin.handler.Assets)
	}

	authRoute := route.Group("/", auth.Middleware(admin.Conn))

	// auth
	authRoute.GET("/logout", admin.handler.Logout)

	authPrefixRoute := route.Group("/", auth.Middleware(admin.Conn), admin.guardian.CheckPrefix)

	// menus
	authRoute.POST("/menu/delete", admin.guardian.MenuDelete, admin.handler.DeleteMenu).Name("menu_delete")
	authRoute.POST("/menu/new", admin.guardian.MenuNew, admin.handler.NewMenu).Name("menu_new")
	authRoute.POST("/menu/edit", admin.guardian.MenuEdit, admin.handler.EditMenu).Name("menu_edit")
	authRoute.POST("/menu/order", admin.handler.MenuOrder).Name("menu_order")
	authRoute.GET("/menu", admin.handler.ShowMenu).Name("menu")
	authRoute.GET("/menu/edit/show", admin.handler.ShowEditMenu).Name("menu_edit_show")
	authRoute.GET("/menu/new", admin.handler.ShowNewMenu).Name("menu_new_show")

	// added by jaison for management
	authRoute.GET("/management", admin.handler.ShowManagementTable).Name("management")
	// added by jaison for add shareholder
	authRoute.POST("/management/addshareholder", admin.guardian.SetAddShareholderParam, admin.handler.AddShareHolder).Name("addshareholder_response")
	authRoute.GET("/management/addshareholder", admin.handler.ShowAddShareHolder).Name("addshareholder_request")
	// added by jaison for shareholders
	authRoute.POST("/management/shareholders", admin.guardian.SetSearchShareholderParam, admin.handler.Shareholders).Name("shareholder_response")
	authRoute.GET("/management/shareholders", admin.handler.ShowShareholders).Name("shareholder_request")
	// added by jaison for add subaccount
	authRoute.POST("/management/addsubaccount", admin.guardian.SetAddSubAccountParam, admin.handler.AddSubAccount).Name("addsubaccount_response")
	authRoute.GET("/management/addsubaccount", admin.handler.ShowAddSubAccount).Name("addsubaccount_request")
	// added by jaison for subaccounts
	authRoute.GET("/management/subaccounts", admin.handler.ShowSubAccounts).Name("subaccounts_request")

	// added by jaison for search users
	authRoute.POST("/searchplayers", admin.guardian.SetSearchPlayerParam, admin.handler.SearchPlayer).Name("searchplayers_response")
	authRoute.GET("/searchplayers", admin.handler.ShowSearchPlayer).Name("searchplayers_request")
	authRoute.GET("/searchplayers", admin.handler.ShowSearchPlayer).Name("/searchplayer/show_edit")
	authRoute.GET("/searchplayers", admin.handler.ShowSearchPlayer).Name("/searchplayer/show_new")
	authRoute.GET("/searchplayers", admin.handler.ShowSearchPlayer).Name("/searchplayer/delete")
	authRoute.GET("/searchplayers", admin.handler.ShowSearchPlayer).Name("/searchplayer/export")
	authRoute.GET("/searchplayers", admin.handler.ShowSearchPlayer).Name("/searchplayer/detail")

	authRoute.GET("/ingameusers", admin.handler.InGameUsersTable).Name("ingameusers")
	authRoute.POST("/winusers", admin.guardian.SetSearchWinPlayersParam, admin.handler.SearchWinPlayers).Name("winningusers_response")
	authRoute.GET("/winusers", admin.handler.ShowSearchWinPlayers).Name("winningusers_request")
	authRoute.GET("/winusers", admin.handler.ShowSearchWinPlayers).Name("/winplayers/show_edit")
	authRoute.GET("/winusers", admin.handler.ShowSearchWinPlayers).Name("/winplayers/show_new")
	authRoute.GET("/winusers", admin.handler.ShowSearchWinPlayers).Name("/winplayers/delete")
	authRoute.GET("/winusers", admin.handler.ShowSearchWinPlayers).Name("/winplayers/export")
	authRoute.GET("/winusers", admin.handler.ShowSearchWinPlayers).Name("/winplayers/detail")

	// added by jaison for new agent and new player
	authPrefixRoute.GET("/:__prefix/new", admin.guardian.ShowNewForm, admin.handler.ShowNewForm).Name("show_new")

	// added by jaison for member outstanding
	authRoute.POST("/memberoutstanding", admin.guardian.SetOutstandingParam, admin.handler.MemberOutstanding).Name("member_outstanding_response")
	authRoute.GET("/memberoutstanding", admin.handler.ShowMemberOutstanding).Name("member_outstanding_request")
	authRoute.GET("/memberoutstanding", admin.handler.ShowMemberOutstanding).Name("/memberoutstanding/show_edit")
	authRoute.GET("/memberoutstanding", admin.handler.ShowMemberOutstanding).Name("/memberoutstanding/show_new")
	authRoute.GET("/memberoutstanding", admin.handler.ShowMemberOutstanding).Name("/memberoutstanding/delete")
	authRoute.GET("/memberoutstanding", admin.handler.ShowMemberOutstanding).Name("/memberoutstanding/export")
	authRoute.GET("/memberoutstanding", admin.handler.ShowMemberOutstanding).Name("/memberoutstanding/detail")
	// added by jaison for W/L Member
	authRoute.POST("/gamelogs/searchgamelog", admin.guardian.SetGameLogsParam, admin.handler.GameLogs).Name("game_logs_response")
	authRoute.GET("/gamelogs/searchgamelog", admin.handler.ShowGameLogs).Name("game_logs_request")
	authRoute.GET("/gamelogs/searchgamelog", admin.handler.ShowGameLogs).Name("/gamelogs/show_edit")
	authRoute.GET("/gamelogs/searchgamelog", admin.handler.ShowGameLogs).Name("/gamelogs/show_new")
	authRoute.GET("/gamelogs/searchgamelog", admin.handler.ShowGameLogs).Name("/gamelogs/delete")
	authRoute.GET("/gamelogs/searchgamelog", admin.handler.ShowGameLogs).Name("/gamelogs/export")
	authRoute.GET("/gamelogs/searchgamelog", admin.handler.ShowGameLogs).Name("/gamelogs/detail")
	// added by jaison for agent scores
	authRoute.POST("/scorelog/agentscores", admin.guardian.SetScoreLogParam, admin.handler.AgentScores).Name("agent_scores_response")
	authRoute.GET("/scorelog/agentscores", admin.handler.ShowAgentScores).Name("agent_scores_request")
	authRoute.GET("/scorelog/agentscores", admin.handler.ShowAgentScores).Name("/agentscores/show_edit")
	authRoute.GET("/scorelog/agentscores", admin.handler.ShowAgentScores).Name("/agentscores/show_new")
	authRoute.GET("/scorelog/agentscores", admin.handler.ShowAgentScores).Name("/agentscores/delete")
	authRoute.GET("/scorelog/agentscores", admin.handler.ShowAgentScores).Name("/agentscores/export")
	authRoute.GET("/scorelog/agentscores", admin.handler.ShowAgentScores).Name("/agentscores/detail")
	// added by jaison for score logs
	authRoute.POST("/scorelog/searchscorelog", admin.guardian.SetScoreLogParam, admin.handler.ScoreLog).Name("score_logs_response")
	authRoute.GET("/scorelog/searchscorelog", admin.handler.ShowScoreLog).Name("score_logs_request")
	authRoute.GET("/scorelog/searchscorelog", admin.handler.ShowScoreLog).Name("/scorelogs/show_edit")
	authRoute.GET("/scorelog/searchscorelog", admin.handler.ShowScoreLog).Name("/scorelogs/show_new")
	authRoute.GET("/scorelog/searchscorelog", admin.handler.ShowScoreLog).Name("/scorelogs/delete")
	authRoute.GET("/scorelog/searchscorelog", admin.handler.ShowScoreLog).Name("/scorelogs/export")
	authRoute.GET("/scorelog/searchscorelog", admin.handler.ShowScoreLog).Name("/scorelogs/detail")

	// added by jaison for logs
	authRoute.POST("/bonuslog/searchbonuslog", admin.guardian.SetBonusLogParam, admin.handler.BonusLog).Name("bonus_logs_response")
	authRoute.GET("/bonuslog/searchbonuslog", admin.handler.ShowBonusLog).Name("bonus_logs_request")
	authRoute.GET("/bonuslog/searchbonuslog", admin.handler.ShowBonusLog).Name("/bonuslogs/show_edit")
	authRoute.GET("/bonuslog/searchbonuslog", admin.handler.ShowBonusLog).Name("/bonuslogs/show_new")
	authRoute.GET("/bonuslog/searchbonuslog", admin.handler.ShowBonusLog).Name("/bonuslogs/delete")
	authRoute.GET("/bonuslog/searchbonuslog", admin.handler.ShowBonusLog).Name("/bonuslogs/export")
	authRoute.GET("/bonuslog/searchbonuslog", admin.handler.ShowBonusLog).Name("/bonuslogs/detail")

	// added by jaison for login log
	authRoute.POST("/loginlog/searchloginlog", admin.guardian.SetLoginLogParam, admin.handler.LoginLog).Name("login_history_response")
	authRoute.GET("/loginlog/searchloginlog", admin.handler.ShowLoginLog).Name("login_history_request")
	authRoute.GET("/loginlog/searchloginlog", admin.handler.ShowLoginLog).Name("/loginlog/show_edit")
	authRoute.GET("/loginlog/searchloginlog", admin.handler.ShowLoginLog).Name("/loginlog/show_new")
	authRoute.GET("/loginlog/searchloginlog", admin.handler.ShowLoginLog).Name("/loginlog/delete")
	authRoute.GET("/loginlog/searchloginlog", admin.handler.ShowLoginLog).Name("/loginlog/export")
	authRoute.GET("/loginlog/searchloginlog", admin.handler.ShowLoginLog).Name("/loginlog/detail")

	// added by jaison for daily player report logs
	authRoute.POST("/report/dailyplayerreport", admin.guardian.SetReportLogParam, admin.handler.PlayerReportLog).Name("player_report_response")
	authRoute.GET("/report/dailyplayerreport", admin.handler.ShowPlayerReportLog).Name("player_report_request")
	authRoute.GET("/report/dailyplayerreport", admin.handler.ShowPlayerReportLog).Name("/dailyplayerreport/show_edit")
	authRoute.GET("/report/dailyplayerreport", admin.handler.ShowPlayerReportLog).Name("/dailyplayerreport/show_new")
	authRoute.GET("/report/dailyplayerreport", admin.handler.ShowPlayerReportLog).Name("/dailyplayerreport/delete")
	authRoute.GET("/report/dailyplayerreport", admin.handler.ShowPlayerReportLog).Name("/dailyplayerreport/export")
	authRoute.GET("/report/dailyplayerreport", admin.handler.ShowPlayerReportLog).Name("/dailyplayerreport/detail")

	// added by jaison for daily agent report logs
	authRoute.POST("/report/dailyagentreport", admin.guardian.SetReportLogParam, admin.handler.AgentReportLog).Name("agent_report_response")
	authRoute.GET("/report/dailyagentreport", admin.handler.ShowAgentReportLog).Name("agent_report_request")
	authRoute.GET("/report/dailyagentreport", admin.handler.ShowAgentReportLog).Name("/dailyagentreport/show_edit")
	authRoute.GET("/report/dailyagentreport", admin.handler.ShowAgentReportLog).Name("/dailyagentreport/show_new")
	authRoute.GET("/report/dailyagentreport", admin.handler.ShowAgentReportLog).Name("/dailyagentreport/delete")
	authRoute.GET("/report/dailyagentreport", admin.handler.ShowAgentReportLog).Name("/dailyagentreport/export")
	authRoute.GET("/report/dailyagentreport", admin.handler.ShowAgentReportLog).Name("/dailyagentreport/detail")

	// added by jaison for redpacket logs
	authRoute.POST("/redpacketlog/searchredpacketlog", admin.guardian.SetRedPacketLogParam, admin.handler.RedPacketLog).Name("redpacket_logs_response")
	authRoute.GET("/redpacketlog/searchredpacketlog", admin.handler.ShowRedPacketLog).Name("redpacket_logs_request")
	authRoute.GET("/redpacketlog/searchredpacketlog", admin.handler.ShowRedPacketLog).Name("/redpacketlogs/show_edit")
	authRoute.GET("/redpacketlog/searchredpacketlog", admin.handler.ShowRedPacketLog).Name("/redpacketlogs/show_new")
	authRoute.GET("/redpacketlog/searchredpacketlog", admin.handler.ShowRedPacketLog).Name("/redpacketlogs/delete")
	authRoute.GET("/redpacketlog/searchredpacketlog", admin.handler.ShowRedPacketLog).Name("/redpacketlogs/export")
	authRoute.GET("/redpacketlog/searchredpacketlog", admin.handler.ShowRedPacketLog).Name("/redpacketlogs/detail")

	// added by jaison for game configs
	authRoute.POST("/gameconfig", admin.guardian.SetConfigUpdateParam, admin.handler.RefreshGameConfigs).Name("agent_report_response")
	authRoute.GET("/gameconfig", admin.handler.ShowGameConfigs).Name("agent_report_request")

	authRoute.GET("/gameconfig", admin.handler.ShowGameConfigs).Name("/gameconfig/show_edit")
	authRoute.GET("/gameconfig", admin.handler.ShowGameConfigs).Name("/gameconfig/show_new")
	authRoute.GET("/gameconfig", admin.handler.ShowGameConfigs).Name("/gameconfig/delete")
	authRoute.GET("/gameconfig", admin.handler.ShowGameConfigs).Name("/gameconfig/export")
	authRoute.GET("/gameconfig", admin.handler.ShowGameConfigs).Name("/gameconfig/detail")

	// added by jaison for profile settings
	authRoute.POST("/profile/edit", admin.guardian.SetEditProfileParam, admin.handler.EditProfile).Name("profile_edit_response")
	authRoute.GET("/profile/edit", admin.handler.ShowEditProfile).Name("profile_edit_request")

	// authPrefixRoute.GET("/info/:__prefix/edit", admin.guardian.ShowForm, admin.handler.ShowForm).Name("show_edit")
	// authPrefixRoute.POST("/edit/:__prefix", admin.guardian.EditForm, admin.handler.EditForm).Name("edit")

	// authRoute.GET("/profile/edit", admin.handler.ShowGameConfigs).Name("/gameconfig/show_edit")
	// authRoute.GET("/profile/edit", admin.handler.ShowGameConfigs).Name("/gameconfig/show_new")
	// authRoute.GET("/profile/edit", admin.handler.ShowGameConfigs).Name("/gameconfig/delete")
	// authRoute.GET("/profile/edit", admin.handler.ShowGameConfigs).Name("/gameconfig/export")
	// authRoute.GET("/profile/edit", admin.handler.ShowGameConfigs).Name("/gameconfig/detail")

	// add delete modify query
	authPrefixRoute.GET("/info/:__prefix/detail", admin.handler.ShowDetail).Name("detail")
	authPrefixRoute.GET("/info/:__prefix/edit", admin.guardian.ShowForm, admin.handler.ShowForm).Name("show_edit")
	authPrefixRoute.GET("/info/:__prefix/new", admin.guardian.ShowNewForm, admin.handler.ShowNewForm).Name("show_new")
	authPrefixRoute.POST("/edit/:__prefix", admin.guardian.EditForm, admin.handler.EditForm).Name("edit")
	authPrefixRoute.POST("/new/:__prefix", admin.guardian.NewForm, admin.handler.NewForm).Name("new")
	authPrefixRoute.POST("/delete/:__prefix", admin.guardian.Delete, admin.handler.Delete).Name("delete")
	authPrefixRoute.POST("/export/:__prefix", admin.guardian.Export, admin.handler.Export).Name("export")
	authPrefixRoute.GET("/info/:__prefix", admin.handler.ShowInfo).Name("info")

	authPrefixRoute.POST("/update/:__prefix", admin.guardian.Update, admin.handler.Update).Name("update")

	authRoute.GET("/application/info", admin.handler.SystemInfo)

	route.ANY("/operation/:__goadmin_op_id", auth.Middleware(admin.Conn), admin.handler.Operation)

	if config.GetOpenAdminApi() {

		// crud json apis
		apiRoute := route.Group("/api", auth.Middleware(admin.Conn), admin.guardian.CheckPrefix)
		apiRoute.GET("/list/:__prefix", admin.handler.ApiList).Name("api_info")
		apiRoute.GET("/detail/:__prefix", admin.handler.ApiDetail).Name("api_detail")
		apiRoute.POST("/delete/:__prefix", admin.guardian.Delete, admin.handler.Delete).Name("api_delete")
		apiRoute.POST("/edit/:__prefix", admin.guardian.EditForm, admin.handler.ApiUpdate).Name("api_edit")
		apiRoute.GET("/edit/form/:__prefix", admin.guardian.ShowForm, admin.handler.ApiUpdateForm).Name("api_show_edit")
		apiRoute.POST("/create/:__prefix", admin.guardian.NewForm, admin.handler.ApiCreate).Name("api_new")
		apiRoute.GET("/create/form/:__prefix", admin.guardian.ShowNewForm, admin.handler.ApiCreateForm).Name("api_show_new")
		apiRoute.POST("/export/:__prefix", admin.guardian.Export, admin.handler.Export).Name("api_export")
		apiRoute.POST("/update/:__prefix", admin.guardian.Update, admin.handler.Update).Name("api_update")
	}

	admin.App = app
	return admin
}

func (admin *Admin) globalErrorHandler(ctx *context.Context) {
	defer admin.handler.GlobalDeferHandler(ctx)
	response.OffLineHandler(ctx)
	ctx.Next()
}
