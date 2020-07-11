package datamodel

import (
	"github.com/HongJaison/go-admin3/context"
	tmpl "github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/chartjs"
	"github.com/HongJaison/go-admin3/template/types"
)

// GetContent return the content of index page.
func GetContent(ctx *context.Context) (types.Panel, error) {
	components := tmpl.Default()
	colComp := components.Col()

	/**************************
	 * Info Box
	/**************************/

	/**************************
	 * Box
	/**************************/

	tableBalanceInfo := components.
		Table().
		SetType("table").
		// SetStyle("striped").
		SetInfoList([]map[string]types.InfoItem{
			{
				"Title":   {Content: "Today Turn Over"},
				"Content": {Content: "200,200.00"},
			}, {
				"Title":   {Content: "Today Valid Turn Over"},
				"Content": {Content: "200,200.00"},
			}, {
				"Title":   {Content: "Today Member Win Lose"},
				"Content": {Content: "200,200.00"},
			}, {
				"Title":   {Content: "Today Agent Win Lose"},
				"Content": {Content: "-10,082.08"},
			},

			{
				"Title":   {Content: ""},
				"Content": {Content: ""},
			},

			{
				"Title":   {Content: "Yesterday Turn Over"},
				"Content": {Content: "200,200.00"},
			}, {
				"Title":   {Content: "Yesterday Valid Turn Over"},
				"Content": {Content: "200,200.00"},
			}, {
				"Title":   {Content: "Yesterday Member Win Lose"},
				"Content": {Content: "200,200.00"},
			}, {
				"Title":   {Content: "Yesterday Agent Win Lose"},
				"Content": {Content: "-10,082.08"},
			},
		}).
		SetHideThead().
		SetMinWidth("5%").
		SetThead(types.Thead{
			{Head: "Title"},
			{Head: "Content"},
		}).GetContent()

	boxInfoBalanceInfo := components.Box().
		SetTheme("default").
		WithHeadBorder().
		SetHeader("<h3>Balance info</h3>").
		SetHeadColor("#FFC68F").
		SetStyle("display: block;").
		SetBody(tableBalanceInfo).
		// SetFooter(`<div class="clearfix"><a href="javascript:void(0)" class="btn btn-sm btn-info btn-flat pull-left">处理订单</a><a href="javascript:void(0)" class="btn btn-sm btn-default btn-flat pull-right">查看所有新订单</a> </div>`).
		GetContent()

	tableOutstanding := components.
		Table().
		SetType("table").
		SetInfoList([]map[string]types.InfoItem{
			{
				"Title":   {Content: "Total Outstanding Bets"},
				"Content": {Content: "57.00"},
			}, {
				"Title":   {Content: "Total Outstanding Balance"},
				"Content": {Content: "-10,082.08"},
			},
		}).
		SetHideThead().
		SetMinWidth("5%").
		SetThead(types.Thead{
			{Head: "Title"},
			{Head: "Content"},
		}).GetContent()

	boxInfoOutstanding := components.Box().
		SetTheme("default").
		WithHeadBorder().
		SetHeader("<h3>Outstanding</h3>").
		SetHeadColor("#FFC68F").
		SetStyle("display: block;").
		SetBody(tableOutstanding).
		GetContent()

	tableDownline := components.
		Table().
		SetType("table").
		SetInfoList([]map[string]types.InfoItem{
			{
				"Members": {Content: "0"},
				"Agents":  {Content: "2"},
			},
		}).
		SetMinWidth("5%").
		SetThead(types.Thead{
			{Head: "Members"},
			{Head: "Agents"},
		}).GetContent()

	boxInfoDownline := components.Box().
		SetTheme("default").
		WithHeadBorder().
		SetHeader("<h3>Your Downline</h3>").
		SetHeadColor("#FFC68F").
		SetStyle("display: block;").
		SetBody(tableDownline).
		GetContent()

	tableNewMembers := components.
		Table().
		SetType("table").
		SetInfoList([]map[string]types.InfoItem{
			{
				"Today":      {Content: "0"},
				"Last Week":  {Content: "6"},
				"Last Month": {Content: "6"},
			},
		}).
		SetMinWidth("5%").
		SetThead(types.Thead{
			{Head: "Today"},
			{Head: "Last Week"},
			{Head: "Last Month"},
		}).GetContent()

	boxInfoNewMembers := components.Box().
		SetTheme("default").
		WithHeadBorder().
		SetHeader("<h3>New Members</h3>").
		SetHeadColor("#FFC68F").
		SetStyle("display: block;").
		SetBody(tableNewMembers).
		GetContent()

	tableCol := colComp.SetSize(types.SizeMD(5)).SetContent(boxInfoBalanceInfo + boxInfoOutstanding + boxInfoDownline + boxInfoNewMembers).GetContent()

	/**************************
	 * Chart
	/**************************/

	lineBalanceInfo := chartjs.Line()

	lineChartBalanceInfo := lineBalanceInfo.
		SetID("balancechart").
		SetHeight(180).
		SetTitle("In June, 2020").
		SetLabels([]string{"5", "10", "15", "20", "25", "30"}).
		AddDataSet("Win Lose").
		DSData([]float64{65, 59, 80, 81, 56, 55, 40, 65, 59, 80, 81, 56, 55, 40, 65, 59, 80, 81, 56, 55, 40, 65, 59, 80, 81, 56, 55, 40}).
		DSFill(false).
		DSBorderColor("rgb(210, 214, 222)").
		DSLineTension(0.1).
		AddDataSet("Valid Turn Over").
		DSData([]float64{28, 48, 40, 19, 86, 27, 90, 28, 48, 40, 19, 86, 27, 90, 28, 48, 40, 19, 86, 27, 90, 28, 48, 40, 19, 86, 27, 90}).
		DSFill(false).
		DSBorderColor("rgba(60,141,188,1)").
		DSLineTension(0.1).
		AddDataSet("Ticket").
		DSData([]float64{28, 48, 40, 19, 86, 27, 90, 28, 48, 40, 19, 86, 27, 90, 28, 48, 40, 19, 86, 27, 90, 28, 48, 40, 19, 86, 27, 90}).
		DSFill(false).
		DSBorderColor("rgba(60,141,188,1)").
		DSLineTension(0.1).
		GetContent()

	lineNewMembers := chartjs.Line()

	lineChartNewMembers := lineNewMembers.
		SetID("salechart").
		SetHeight(180).
		SetTitle("Sales: 1 Jan, 2019 - 30 Jul, 2019").
		SetLabels([]string{"January", "February", "March", "April", "May", "June", "July"}).
		AddDataSet("Electronics").
		DSData([]float64{65, 59, 80, 81, 56, 55, 40}).
		DSFill(false).
		DSBorderColor("rgb(210, 214, 222)").
		DSLineTension(0.1).
		AddDataSet("Digital Goods").
		DSData([]float64{28, 48, 40, 19, 86, 27, 90}).
		DSFill(false).
		DSBorderColor("rgba(60,141,188,1)").
		DSLineTension(0.1).
		GetContent()

	boxLineCharts := colComp.SetContent(lineChartBalanceInfo + lineChartNewMembers).SetSize(types.SizeMD(7)).GetContent()

	return types.Panel{
		Content:     tableCol + boxLineCharts,
		Title:       "Dashboard 1",
		Description: "dashboard example",
	}, nil
}
