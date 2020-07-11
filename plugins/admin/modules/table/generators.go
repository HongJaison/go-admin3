package table

import (
	"database/sql"
	"errors"
	"fmt"
	tmpl "html/template"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/HongJaison/go-admin3/modules/ui"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/tools"

	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/collection"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/modules/db"
	"github.com/HongJaison/go-admin3/modules/db/dialect"
	errs "github.com/HongJaison/go-admin3/modules/errors"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/modules/logger"
	"github.com/HongJaison/go-admin3/modules/utils"
	"github.com/HongJaison/go-admin3/plugins/admin/models"
	form2 "github.com/HongJaison/go-admin3/plugins/admin/modules/form"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/parameter"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/types"
	"github.com/HongJaison/go-admin3/template/types/action"
	"github.com/HongJaison/go-admin3/template/types/form"
	selection "github.com/HongJaison/go-admin3/template/types/form/select"
	"github.com/HongJaison/html"
	"golang.org/x/crypto/bcrypt"
)

type SystemTable struct {
	conn db.Connection
	c    *config.Config
}

func NewSystemTable(conn db.Connection, c *config.Config) *SystemTable {
	return &SystemTable{conn: conn, c: c}
}

// added by jaison
func (s *SystemTable) GetAgentsTable(ctx *context.Context) (agentTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetAgentTable")

	agentTableConfig := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)

	agentTable = NewDefaultTable(agentTableConfig)

	info := agentTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.SetSortAsc()

	info.AddField("#", "id", db.Int)
	info.AddField(lg("UserName"), "username", db.Varchar)
	info.AddField(lg("Name"), "name", db.Varchar)
	info.AddField(lg("Score"), "score", db.Decimal)
	info.AddField(lg("Operation"), "state", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			buttons := template.HTML(`<td id="op2_0" class="text-left">`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Agent/EditScore?id=` + model.ID + `'">set score</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/ScoreLog/SearchScoreLog?sid=` + model.ID + `'">score log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Agent/Edit?id=` + model.ID + `'">edit</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Report/Search?sid=` + model.ID + `'">report</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Report/Chart?sid=` + model.ID + `'">chart</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/BonusLog/SetBonus?sid=` + model.ID + `'">set bonus</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/BonusLog/SearchBonusLog?sid=` + model.ID + `'">bonus log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SetRedPacket?sid=` + model.ID + `'">set redpacket</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SearchRedPacketLog?sid=` + model.ID + `'">redpacket log</button>`)

			if model.Row[`state`] == 1 {
				buttons += template.HTML(`<button type="button" name="agentable" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">disable</button>`)
			} else {
				buttons += template.HTML(`<button type="button" name="agentable" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">enable</button>`)
			}

			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Agent/SetAgentLoginIP?agentid=` + model.ID + `'">agent login ip</button>`)
			buttons += template.HTML(`<button type="button" name="closeonlineagent" title="` + model.Row[`username`].(string) + `" rel="closeonlineagent" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">closeonline</button>`)
			buttons += template.HTML(`<button type="button" name="closeonlineagentall" title="` + model.Row[`username`].(string) + `" rel="closeonlineagentall" player="` + model.ID + `" class="btn btn-yahoo btn-xs" onfocus="this.blur();" onclick="">closeonlineforall</button>`)

			buttons += template.HTML(`</td>`)
			return buttons
		})

	info.HideDeleteButton()
	info.HideExportButton()

	info.HideNewButton()
	info.HideFilterButton()
	info.HideRowSelector()
	info.HideEditButton()
	info.HideDetailButton()

	info.SetTable("Agents").
		SetTitle("Agent List").
		SetDescription("").
		SetHeaderHtml(template.HTML(`
			<h6 class="box-title" id="d_tip_1" style="font-size:0.9em;">
				<span class="badge bg-red hidden-xs hidden-sm">My agent list</span>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" name="create" href="javascript:;" onclick="checkStateAgent();" target="_self">AddAgent</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" href="/Report/Search?action=TotalAgent&parentId=@ViewBag.parentId" target="_self">Agent total report</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" href="###" name="reload_all_agents" onclick="javascript:reloadAgentList();" target="_self">Reload agentList</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" href="/BonusLog/SearchBonusLogTotal" name="totalBonusLog" onclick="" target="_self">Total Bonus Log</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" href="/RedPacketLog/SearchRedPacketLogTotal" name="totalRedPacketLog" onclick="" target="_self">Total RedPacket Log</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" href="/SubAccount/Index?parentId=@ViewBag.parentId" target="_self">SubAccount List</a>
				<div class="margin visible-xs visible-sm">
					<div class="btn-group">
						<button class="btn btn-default btn-flat" type="button">AgentList Menu</button>
						<button data-toggle="dropdown" class="btn btn-default btn-flat dropdown-toggle" type="button">
							<span class="caret"></span>
							<span class="sr-only">Toggle Dropdown</span>
						</button>
						<ul role="menu" class="dropdown-menu">
							<li>
								<a name="create" href="javascript:;" onclick="checkStateAgent();" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> AddAgent</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="/Report/Search?action=TotalAgent&parentId=@ViewBag.parentId" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Agent total report</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="###" onclick="javascript:reloadAgentList();" name="reload_all_agents" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Reload agentList</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="/BonusLog/SearchBonusLogTotal" onclick="" name="totalBonusLog" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Total Bonus Log</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="/RedPacketLog/SearchRedPacketLogTotal" onclick="" name="totalRedPacketLog" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Total RedPacket Log</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="/SubAccount/Index?parentId=@ViewBag.parentId" onclick="" name="subaccount" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> SubAccount List</span>
								</a>
							</li>

						</ul>
					</div>
				</div>
			</h6>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`))

	countryOptions := types.FieldOptions{
		{Value: "ANY", Text: "Any"},
		{Value: "AF", Text: "Afghanistan"},
		{Value: "AL", Text: "Albania"},
		{Value: "DZ", Text: "Algeria"},
		{Value: "AS", Text: "American Samoa"},
		{Value: "AD", Text: "Andorra"},
		{Value: "AG", Text: "Angola"},
		{Value: "AI", Text: "Anguilla"},
		{Value: "AG", Text: "Antigua &amp; Barbuda"},
		{Value: "AR", Text: "Argentina"},
		{Value: "AA", Text: "Armenia"},
		{Value: "AW", Text: "Aruba"},
		{Value: "AU", Text: "Australia"},
		{Value: "AT", Text: "Austria"},
		{Value: "AZ", Text: "Azerbaijan"},
		{Value: "BS", Text: "Bahamas"},
		{Value: "BH", Text: "Bahrain"},
		{Value: "BD", Text: "Bangladesh"},
		{Value: "BB", Text: "Barbados"},
		{Value: "BY", Text: "Belarus"},
		{Value: "BE", Text: "Belgium"},
		{Value: "BZ", Text: "Belize"},
		{Value: "BJ", Text: "Benin"},
		{Value: "BM", Text: "Bermuda"},
		{Value: "BT", Text: "Bhutan"},
		{Value: "BO", Text: "Bolivia"},
		{Value: "BL", Text: "Bonaire"},
		{Value: "BA", Text: "Bosnia &amp; Herzegovina"},
		{Value: "BW", Text: "Botswana"},
		{Value: "BR", Text: "Brazil"},
		{Value: "BC", Text: "British Indian Ocean Ter"},
		{Value: "BN", Text: "Brunei"},
		{Value: "BG", Text: "Bulgaria"},
		{Value: "BF", Text: "Burkina Faso"},
		{Value: "BI", Text: "Burundi"},
		{Value: "KH", Text: "Cambodia"},
		{Value: "CM", Text: "Cameroon"},
		{Value: "CA", Text: "Canada"},
		{Value: "IC", Text: "Canary Islands"},
		{Value: "CV", Text: "Cape Verde"},
		{Value: "KY", Text: "Cayman Islands"},
		{Value: "CF", Text: "Central African Republic"},
		{Value: "TD", Text: "Chad"},
		{Value: "CD", Text: "Channel Islands"},
		{Value: "CL", Text: "Chile"},
		{Value: "CN", Text: "China"},
		{Value: "CI", Text: "Christmas Island"},
		{Value: "CS", Text: "Cocos Island"},
		{Value: "CO", Text: "Colombia"},
		{Value: "CC", Text: "Comoros"},
		{Value: "CG", Text: "Congo"},
		{Value: "CK", Text: "Cook Islands"},
		{Value: "CR", Text: "Costa Rica"},
		{Value: "CT", Text: "Cote D'Ivoire"},
		{Value: "HR", Text: "Croatia"},
		{Value: "CU", Text: "Cuba"},
		{Value: "CB", Text: "Curacao"},
		{Value: "CY", Text: "Cyprus"},
		{Value: "CZ", Text: "Czech Republic"},
		{Value: "DK", Text: "Denmark"},
		{Value: "DJ", Text: "Djibouti"},
		{Value: "DM", Text: "Dominica"},
		{Value: "DO", Text: "Dominican Republic"},
		{Value: "TM", Text: "East Timor"},
		{Value: "EC", Text: "Ecuador"},
		{Value: "EG", Text: "Egypt"},
		{Value: "SV", Text: "El Salvador"},
		{Value: "GQ", Text: "Equatorial Guinea"},
		{Value: "ER", Text: "Eritrea"},
		{Value: "EE", Text: "Estonia"},
		{Value: "ET", Text: "Ethiopia"},
		{Value: "FA", Text: "Falkland Islands"},
		{Value: "FO", Text: "Faroe Islands"},
		{Value: "FJ", Text: "Fiji"},
		{Value: "FI", Text: "Finland"},
		{Value: "FR", Text: "France"},
		{Value: "GF", Text: "French Guiana"},
		{Value: "PF", Text: "French Polynesia"},
		{Value: "FS", Text: "French Southern Ter"},
		{Value: "GA", Text: "Gabon"},
		{Value: "GM", Text: "Gambia"},
		{Value: "GE", Text: "Georgia"},
		{Value: "DE", Text: "Germany"},
		{Value: "GH", Text: "Ghana"},
		{Value: "GI", Text: "Gibraltar"},
		{Value: "GB", Text: "Great Britain"},
		{Value: "GR", Text: "Greece"},
		{Value: "GL", Text: "Greenland"},
		{Value: "GD", Text: "Grenada"},
		{Value: "GP", Text: "Guadeloupe"},
		{Value: "GU", Text: "Guam"},
		{Value: "GT", Text: "Guatemala"},
		{Value: "GN", Text: "Guinea"},
		{Value: "GY", Text: "Guyana"},
		{Value: "HT", Text: "Haiti"},
		{Value: "HW", Text: "Hawaii"},
		{Value: "HN", Text: "Honduras"},
		{Value: "HK", Text: "Hong Kong"},
		{Value: "HU", Text: "Hungary"},
		{Value: "IS", Text: "Iceland"},
		{Value: "IN", Text: "India"},
		{Value: "ID", Text: "Indonesia"},
		{Value: "IA", Text: "Iran"},
		{Value: "IQ", Text: "Iraq"},
		{Value: "IR", Text: "Ireland"},
		{Value: "IM", Text: "Isle of Man"},
		{Value: "IL", Text: "Israel"},
		{Value: "IT", Text: "Italy"},
		{Value: "JM", Text: "Jamaica"},
		{Value: "JP", Text: "Japan"},
		{Value: "JO", Text: "Jordan"},
		{Value: "KZ", Text: "Kazakhstan"},
		{Value: "KE", Text: "Kenya"},
		{Value: "KI", Text: "Kiribati"},
		{Value: "KP", Text: "Korea North"},
		{Value: "KR", Text: "Korea South"},
		{Value: "KW", Text: "Kuwait"},
		{Value: "KG", Text: "Kyrgyzstan"},
		{Value: "LA", Text: "Laos"},
		{Value: "LV", Text: "Latvia"},
		{Value: "LB", Text: "Lebanon"},
		{Value: "LS", Text: "Lesotho"},
		{Value: "LR", Text: "Liberia"},
		{Value: "LY", Text: "Libya"},
		{Value: "LI", Text: "Liechtenstein"},
		{Value: "LT", Text: "Lithuania"},
		{Value: "LU", Text: "Luxembourg"},
		{Value: "MO", Text: "Macau"},
		{Value: "MK", Text: "Macedonia"},
		{Value: "MG", Text: "Madagascar"},
		{Value: "MY", Text: "Malaysia"},
		{Value: "MW", Text: "Malawi"},
		{Value: "MV", Text: "Maldives"},
		{Value: "ML", Text: "Mali"},
		{Value: "MT", Text: "Malta"},
		{Value: "MH", Text: "Marshall Islands"},
		{Value: "MQ", Text: "Martinique"},
		{Value: "MR", Text: "Mauritania"},
		{Value: "MU", Text: "Mauritius"},
		{Value: "ME", Text: "Mayotte"},
		{Value: "MX", Text: "Mexico"},
		{Value: "MI", Text: "Midway Islands"},
		{Value: "MD", Text: "Moldova"},
		{Value: "MC", Text: "Monaco"},
		{Value: "MN", Text: "Mongolia"},
		{Value: "MS", Text: "Montserrat"},
		{Value: "MA", Text: "Morocco"},
		{Value: "MZ", Text: "Mozambique"},
		{Value: "MM", Text: "Myanmar"},
		{Value: "NA", Text: "Nambia"},
		{Value: "NU", Text: "Nauru"},
		{Value: "NP", Text: "Nepal"},
		{Value: "AN", Text: "Netherland Antilles"},
		{Value: "NL", Text: "Netherlands (Holland, Europe)"},
		{Value: "NV", Text: "Nevis"},
		{Value: "NC", Text: "New Caledonia"},
		{Value: "NZ", Text: "New Zealand"},
		{Value: "NI", Text: "Nicaragua"},
		{Value: "NE", Text: "Niger"},
		{Value: "NG", Text: "Nigeria"},
		{Value: "NW", Text: "Niue"},
		{Value: "NF", Text: "Norfolk Island"},
		{Value: "NO", Text: "Norway"},
		{Value: "OM", Text: "Oman"},
		{Value: "PK", Text: "Pakistan"},
		{Value: "PW", Text: "Palau Island"},
		{Value: "PS", Text: "Palestine"},
		{Value: "PA", Text: "Panama"},
		{Value: "PG", Text: "Papua New Guinea"},
		{Value: "PY", Text: "Paraguay"},
		{Value: "PE", Text: "Peru"},
		{Value: "PH", Text: "Philippines"},
		{Value: "PO", Text: "Pitcairn Island"},
		{Value: "PL", Text: "Poland"},
		{Value: "PT", Text: "Portugal"},
		{Value: "PR", Text: "Puerto Rico"},
		{Value: "QA", Text: "Qatar"},
		{Value: "ME", Text: "Republic of Montenegro"},
		{Value: "RS", Text: "Republic of Serbia"},
		{Value: "RE", Text: "Reunion"},
		{Value: "RO", Text: "Romania"},
		{Value: "RU", Text: "Russia"},
		{Value: "RW", Text: "Rwanda"},
		{Value: "NT", Text: "St Barthelemy"},
		{Value: "EU", Text: "St Eustatius"},
		{Value: "HE", Text: "St Helena"},
		{Value: "KN", Text: "St Kitts-Nevis"},
		{Value: "LC", Text: "St Lucia"},
		{Value: "MB", Text: "St Maarten"},
		{Value: "PM", Text: "St Pierre &amp; Miquelon"},
		{Value: "VC", Text: "St Vincent &amp; Grenadines"},
		{Value: "SP", Text: "Saipan"},
		{Value: "SO", Text: "Samoa"},
		{Value: "AS", Text: "Samoa American"},
		{Value: "SM", Text: "San Marino"},
		{Value: "ST", Text: "Sao Tome &amp; Principe"},
		{Value: "SA", Text: "Saudi Arabia"},
		{Value: "SN", Text: "Senegal"},
		{Value: "RS", Text: "Serbia"},
		{Value: "SC", Text: "Seychelles"},
		{Value: "SL", Text: "Sierra Leone"},
		{Value: "SG", Text: "Singapore"},
		{Value: "SK", Text: "Slovakia"},
		{Value: "SI", Text: "Slovenia"},
		{Value: "SB", Text: "Solomon Islands"},
		{Value: "OI", Text: "Somalia"},
		{Value: "ZA", Text: "South Africa"},
		{Value: "ES", Text: "Spain"},
		{Value: "LK", Text: "Sri Lanka"},
		{Value: "SD", Text: "Sudan"},
		{Value: "SR", Text: "Suriname"},
		{Value: "SZ", Text: "Swaziland"},
		{Value: "SE", Text: "Sweden"},
		{Value: "CH", Text: "Switzerland"},
		{Value: "SY", Text: "Syria"},
		{Value: "TA", Text: "Tahiti"},
		{Value: "TW", Text: "Taiwan"},
		{Value: "TJ", Text: "Tajikistan"},
		{Value: "TZ", Text: "Tanzania"},
		{Value: "TH", Text: "Thailand"},
		{Value: "TG", Text: "Togo"},
		{Value: "TK", Text: "Tokelau"},
		{Value: "TO", Text: "Tonga"},
		{Value: "TT", Text: "Trinidad &amp; Tobago"},
		{Value: "TN", Text: "Tunisia"},
		{Value: "TR", Text: "Turkey"},
		{Value: "TU", Text: "Turkmenistan"},
		{Value: "TC", Text: "Turks &amp; Caicos Is"},
		{Value: "TV", Text: "Tuvalu"},
		{Value: "UG", Text: "Uganda"},
		{Value: "UA", Text: "Ukraine"},
		{Value: "AE", Text: "United Arab Emirates"},
		{Value: "GB", Text: "United Kingdom"},
		{Value: "US", Text: "United States of America"},
		{Value: "UY", Text: "Uruguay"},
		{Value: "UZ", Text: "Uzbekistan"},
		{Value: "VU", Text: "Vanuatu"},
		{Value: "VS", Text: "Vatican City State"},
		{Value: "VE", Text: "Venezuela"},
		{Value: "VN", Text: "Vietnam"},
		{Value: "VB", Text: "Virgin Islands (Brit)"},
		{Value: "VA", Text: "Virgin Islands (USA)"},
		{Value: "WK", Text: "Wake Island"},
		{Value: "WF", Text: "Wallis &amp; Futana Is"},
		{Value: "YE", Text: "Yemen"},
		{Value: "ZR", Text: "Zaire"},
		{Value: "ZM", Text: "Zambia"},
		{Value: "ZW", Text: "Zimbabwe"},
	}

	formList := agentTable.GetForm().AddXssJsFilter().
		HideBackButton().
		HideContinueNewCheckBox()
		// HideResetButton()

	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("UserName"), "username", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("should be unique"))).FieldMust()
	formList.AddField(lg("Password"), "password", db.Varchar, form.Password).FieldMust()
	formList.AddField(lg("Country"), "country", db.Varchar, form.SelectSingle).FieldOptions(countryOptions).FieldMust()
	formList.AddField(lg("Score"), "score", db.Decimal, form.Text).FieldMust()
	formList.AddField(lg("Name"), "name", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("should be unique")))
	formList.AddField(lg("Tel"), "tel", db.Varchar, form.Text)
	formList.AddField(lg("Description"), "description", db.Varchar, form.Text)

	formList.AddField(lg(`DateRange`), `DateRange`, db.Varchar, form.DateRange)

	formList.SetTable("Agents").
		SetTitle(lg("Add new agent")).
		// SetPostValidator(func(values form2.Values) error {
		// 	fmt.Println("PostValidator")
		// 	fmt.Println(values)

		// 	return nil
		// }).
		SetPostHook(func(values form2.Values) error {
			fmt.Println("PostHook")
			fmt.Println(values)

			return nil
		}).
		SetInsertFn(func(values form2.Values) error {

			fmt.Println(values)

			if !models.User().SetConn(s.conn).FindByUserName(values.Get("username")).IsEmpty() {
				return errors.New("Username exists")
			}

			if !models.User().SetConn(s.conn).FindByName(values.Get("name")).IsEmpty() {
				return errors.New("Name exists")
			}

			scoreValue, err := strconv.ParseFloat(values.Get("score"), 64)

			if err != nil {
				return errors.New("Input correct score value")
			}

			username := values.Get("username")
			password := values.Get("password")
			country := values.Get("country")
			name := values.Get("name")
			tel := values.Get("tel")
			description := values.Get("description")
			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
				_, newUserErr := models.User().WithTx(tx).SetConn(s.conn).New(username, password, name, country, tel, description, "", scoreValue)

				if db.CheckError(newUserErr, db.INSERT) {
					return newUserErr, nil
				}

				return nil, nil
			})

			return txErr
		})

	return
}

// added by jaison
func (s *SystemTable) GetShareholdersTable(ctx *context.Context) (agentTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetShareholdersTable")

	agentTableConfig := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)

	agentTable = NewDefaultTable(agentTableConfig)

	info := agentTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.SetSortAsc()

	info.AddField("#", "id", db.Int)
	info.AddField(lg("Level"), "level", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			return template.HTML(`Shareholder`)
		})
	info.AddField(lg("Login name"), "username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// "/Manage/Management?parentId=" + model.Value
			return template.HTML("<a>" + model.Value + "</a>")
		})
	info.AddField(lg("Nickname"), "name", db.Varchar)
	info.AddField(lg("Phone"), "tel", db.Nvarchar)
	info.AddField(lg("Suspend"), "state", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Row[`state`].(int64) == 1 {
				tag += template.HTML(`NO`)
			} else if model.Row[`state`].(int64) == 0 {
				tag += template.HTML(`YES`)
			} else if model.Row[`state`].(int64) == -1 {
				tag += template.HTML(`YES`)
			}

			return tag
		})
	info.AddField(lg("Lock"), "lock", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Row[`state`].(int64) == -1 {
				tag += template.HTML(`Lock`)
			} else {
				tag += template.HTML(`Unlock`)
			}

			return tag
		})
	info.AddField(lg("Credit"), "score", db.Decimal)
	info.AddField(lg("PT"), "pt", db.Varchar)
	info.AddField(lg("Currency"), "currency", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// button := template.HTML(`<td id="op2_0" class="text-center">`)
			button := template.HTML(`<button type="button" name="shareholders" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">View</button>`)
			// button += template.HTML(`</td>`)
			return button
		})
	info.AddField(lg("Commission"), "commission", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// button := template.HTML(`<td id="op2_0" class="text-center">`)
			button := template.HTML(`<button type="button" name="shareholders" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">View</button>`)
			// button += template.HTML(`</td>`)
			return button
		})
	info.AddField(lg("Last Login Date"), "lastlogindate", db.Varchar)
	info.AddField(lg("Last Login IP"), "lastloginip", db.Varchar)

	info.HideDeleteButton()
	info.HideExportButton()

	info.HideNewButton()
	info.HideFilterButton()
	info.HideRowSelector()
	info.HideEditButton()
	info.HideDetailButton()

	info.SetTable("Agents")

	return
}

// added by jaison
func (s *SystemTable) GetSubAccountsTable(ctx *context.Context) (playerTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetSubAccountsTable")

	playerTableConfig := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)

	playerTable = NewDefaultTable(playerTableConfig)

	info := playerTable.GetInfo().AddXssJsFilter().HideFilterArea()
	info.SetSortAsc()

	// #	UserName	Country	Score	PlayerStatus	DisOnlineDay	Name	Operation
	info.AddField("#", "id", db.Int)
	info.AddField(lg("Login Name"), "accountname", db.Varchar)
	info.AddField(lg("Nickname"), "nickname", db.Varchar)
	info.AddField(lg("Phone"), "tel", db.Varchar)
	info.AddField(lg("Edit"), "edit", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			return template.HTML("<a><span>Edit</span></a>")
		})
	info.AddField(lg("password"), "password", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			return template.HTML("<a><span>Password</span></a>")
		})
	info.AddField(lg("Lock"), "state", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Row[`state`].(int64) == 1 {
				tag += template.HTML(`NO`)
			} else if model.Row[`state`].(int64) == 0 {
				tag += template.HTML(`YES`)
			} else if model.Row[`state`].(int64) == -1 {
				tag += template.HTML(`YES`)
			}

			return tag
		})
	info.AddField(lg("Account"), "level", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			strLevel := strconv.FormatInt(model.Row[`level`].(int64), 3)
			if len(strLevel) < 5 {
				gaps := 5 - len(strLevel)

				for i := 0; i < gaps; i++ {
					strLevel = "0" + strLevel
				}
			}

			if strLevel[0] == '0' {
				return template.HTML("Off")
			}

			if strLevel[0] == '1' {
				return template.HTML("View")
			}

			if strLevel[0] == '2' {
				return template.HTML("Edit")
			}

			return template.HTML(`Undefined`)
		})
	info.AddField(lg("Member"), "member", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			strLevel := strconv.FormatInt(model.Row[`level`].(int64), 3)
			if len(strLevel) < 5 {
				gaps := 5 - len(strLevel)

				for i := 0; i < gaps; i++ {
					strLevel = "0" + strLevel
				}
			}

			if strLevel[1] == '0' {
				return template.HTML("Off")
			}

			if strLevel[1] == '1' {
				return template.HTML("View")
			}

			if strLevel[1] == '2' {
				return template.HTML("Edit")
			}

			return template.HTML(`Undefined`)
		})
	info.AddField(lg("Stock"), "stock", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			strLevel := strconv.FormatInt(model.Row[`level`].(int64), 3)
			if len(strLevel) < 5 {
				gaps := 5 - len(strLevel)

				for i := 0; i < gaps; i++ {
					strLevel = "0" + strLevel
				}
			}

			if strLevel[2] == '0' {
				return template.HTML("Off")
			}

			if strLevel[2] == '1' {
				return template.HTML("View")
			}

			if strLevel[2] == '2' {
				return template.HTML("Edit")
			}

			return template.HTML(`Undefined`)
		})
	info.AddField(lg("Report"), "report", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			strLevel := strconv.FormatInt(model.Row[`level`].(int64), 3)
			if len(strLevel) < 5 {
				gaps := 5 - len(strLevel)

				for i := 0; i < gaps; i++ {
					strLevel = "0" + strLevel
				}
			}

			if strLevel[3] == '0' {
				return template.HTML("Off")
			}

			if strLevel[3] == '1' {
				return template.HTML("View")
			}

			if strLevel[3] == '2' {
				return template.HTML("Edit")
			}

			return template.HTML(`Undefined`)
		})
	info.AddField(lg("Payment"), "payment", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			strLevel := strconv.FormatInt(model.Row[`level`].(int64), 3)
			if len(strLevel) < 5 {
				gaps := 5 - len(strLevel)

				for i := 0; i < gaps; i++ {
					strLevel = "0" + strLevel
				}
			}

			if strLevel[4] == '0' {
				return template.HTML("Off")
			}

			if strLevel[4] == '1' {
				return template.HTML("View")
			}

			if strLevel[4] == '2' {
				return template.HTML("Edit")
			}

			return template.HTML(`Undefined`)
		})
	info.AddField(lg("Last Login Date"), "lastlogindate", db.Decimal)
	info.AddField(lg("Last Login IP"), "lastloginip", db.Decimal)

	info.HideDeleteButton()
	info.HideExportButton()

	info.HideNewButton()
	info.HideFilterButton()
	info.HideRowSelector()
	info.HideEditButton()
	info.HideDetailButton()

	info.SetTable("SubAccounts")
	return
}

// added by jaison
func (s *SystemTable) GetPlayersTable(ctx *context.Context) (playerTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetPlayersTable")

	playerTableConfig := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)

	playerTable = NewDefaultTable(playerTableConfig)

	info := playerTable.GetInfo().AddXssJsFilter().HideFilterArea()
	info.SetSortAsc()

	// #	UserName	Country	Score	PlayerStatus	DisOnlineDay	Name	Operation
	info.AddField("#", "id", db.Int)
	info.AddField(lg("UserName"), "username", db.Varchar)
	info.AddField(lg("Country"), "country", db.Varchar)
	info.AddField(lg("Score"), "score", db.Decimal)
	info.AddField(lg("PlayerStatus"), "playerstatus", db.Decimal)
	info.AddField(lg("DisOnlineDay"), "disonlineday", db.Decimal)
	info.AddField(lg("Name"), "name", db.Varchar)
	info.AddField(lg("Process"), "operation", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			buttons := template.HTML(`<td id="op2_0" class="text-left">`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Player/EditScore?id=` + model.ID + `'">set score</button>`)

			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/ScoreLog/SearchScoreLog?id=` + model.ID + `'">score log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Player/Edit?id=` + model.ID + `'">edit</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Report/Search?id=` + model.ID + `'">report</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/GameLog/SearchGameLog?id=` + model.ID + `'">game log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SetRedPacket?id=` + model.ID + `'">set redpacket</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SearchRedPacketLog?id=` + model.ID + `'">redpacket log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Player/SendEventPlayer?playerid=` + model.ID + `'">send event</button>`)
			buttons += template.HTML(`<button type="button" name="forcequite" title="` + model.Row[`username`].(string) + `" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">quit game</button>`)

			if model.Row[`state`] == 1 {
				buttons += template.HTML(`<button type="button" name="able" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">disable</button>`)
			} else {
				buttons += template.HTML(`<button type="button" name="able" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">enable</button>`)
			}

			buttons += template.HTML(`<button type="button" name="stuckremove" title="` + model.Row[`username`].(string) + `" player="` + model.ID + `" class="btn btn-dropbox btn-xs" onfocus="this.blur();" onclick="">remove stuck</button>`)
			buttons += template.HTML(`<button type="button" name="closeonlineplayer" title="` + model.Row[`username`].(string) + `" player="` + model.ID + `" class="btn btn-dropbox btn-xs" onfocus="this.blur();" onclick="">closeonline</button>`)

			buttons += template.HTML(`</td>`)
			return buttons
		})

	info.HideDeleteButton()
	info.HideExportButton()

	info.HideNewButton()
	info.HideFilterButton()
	info.HideRowSelector()
	info.HideEditButton()
	info.HideDetailButton()

	info.SetTable("Players").
		SetTitle("Player List").
		SetDescription("").
		SetHeaderHtml(template.HTML(`
			<h6 class="box-title" id="d_tip_2" style="font-size:0.9em;">
				<span class="badge bg-red hidden-xs hidden-sm">My player list</span>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" name="create" href="javascript:;" onclick="checkStatePlayer();" target="_self">AddPlayer</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" href="/Report/Search?action=TotalPlayer&parentId=@ViewBag.parentId" target="_self">Player total report</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" name="enable_all" style="cursor:pointer">Enable all player</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" name="disable_all" style="cursor:pointer">Disable all player</a>
				<span class="hidden-xs hidden-sm">　｜　</span>
				<a class="hidden-xs hidden-sm" name="reload_all_players" href="###" onclick="javascript:reloadPlayerList();">Reload playerList</a>
				<div class="margin visible-xs visible-sm">
					<div class="btn-group">
						<button class="btn btn-default btn-flat" type="button">PlayerList Menu</button>
						<button data-toggle="dropdown" class="btn btn-default btn-flat dropdown-toggle" type="button">
							<span class="caret"></span>
							<span class="sr-only">Toggle Dropdown</span>
						</button>
						<ul role="menu" class="dropdown-menu">
							<li>
								<a name="create" href="javascript:;" onclick="checkStatePlayer();" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Add player</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="/Report/Search?action=TotalPlayer&parentId=@ViewBag.parentId" target="_self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Player total report</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a target="_self" name="enable_all">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Enable all player</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a name="disable_all" target=" _self">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Disable all player</span>
								</a>
							</li>
							<li class="divider"></li>
							<li>
								<a href="###" target="_self" name="reload_all_players" onclick="javascript:reloadPlayerList();">
									<span class="text-bold"><i class="fa fa-angle-right"></i> Reload playerList</span>
								</a>
							</li>
						</ul>
					</div>
				</div>
			</h6>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`))

	return
}

// for search users
// added by jaison
func (s *SystemTable) GetPlayerAgentTable(ctx *context.Context) (playerAgentTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetPlayerAgentTable")

	playerAgentTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := playerAgentTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldHide()
	info.AddField(lg("Username"), "username", db.Varchar)
	info.AddField(lg("Score"), "score", db.Decimal)
	info.AddField(lg("Name"), "name", db.Varchar)
	info.AddField(lg("Agent"), "agent", db.Varchar)
	info.AddField(lg("Operation"), "state", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// fmt.Println(model)

			// buttons := template.HTML(`<td id="op2_0" class="text-left">`)
			buttons := template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Agent/EditScore?id=` + model.ID + `'">set score</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/ScoreLog/SearchScoreLog?sid=` + model.ID + `'">score log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Agent/Edit?id=` + model.ID + `'">edit</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Report/Search?sid=` + model.ID + `'">report</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Report/Chart?sid=` + model.ID + `'">chart</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="">Total</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/BonusLog/SetBonus?sid=` + model.ID + `'">set bonus</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/BonusLog/SearchBonusLog?sid=` + model.ID + `'">bonus log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SetRedPacket?sid=` + model.ID + `'">set redpacket</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SearchRedPacketLog?sid=` + model.ID + `'">redpacket log</button>`)

			if model.Row[`status`] == 1 {
				buttons += template.HTML(`<button type="button" name="agentable" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">disable</button>`)
			} else {
				buttons += template.HTML(`<button type="button" name="agentable" title="` + model.Row[`username`].(string) + `" rel="enable" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">enable</button>`)
			}

			buttons += template.HTML(`<button type="button" name="closeonlineagent" title="` + model.Row[`username`].(string) + `" rel="closeonlineagent" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">closeonline</button>`)
			buttons += template.HTML(`<button type="button" name="closeonlineagentall" title="` + model.Row[`username`].(string) + `" rel="closeonlineagentall" player="` + model.ID + `" class="btn btn-yahoo btn-xs" onfocus="this.blur();" onclick="">closeonlineforall</button>`)
			// buttons += template.HTML(`</td>`)

			return buttons
		})

	info.SetTable("Agents").
		SetTitle("Higher Level AgentList").
		SetSortAsc()

	return
}

// added by jaison
func (s *SystemTable) GetPlayerTable(ctx *context.Context) (playerTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetPlayerTable")

	playerTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := playerTable.GetInfo().AddXssJsFilter().HideFilterArea()

	// UserName	Online	PlayerStatus	Agent	Balance	Name	Tel	Description	SADescription	Operation
	info.AddField("ID", "id", db.Int).FieldHide()
	info.AddField(lg("UserName"), "username", db.Varchar)
	info.AddField(lg("Online"), "online", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// fmt.Println(model.Row[`isonline`])

			var compareValue int64 = 0

			tag := template.HTML(``)
			if model.Row[`isonline`] == compareValue {
				onlineDate, err := time.Parse(time.RFC3339, model.Row[`online`].(string))

				if err != nil || onlineDate.Year() <= 1 {
					tag += template.HTML(`<span class='badge bg-gray'>N/A</span>`)
					return tag
				}

				nowTime := time.Now().UTC()

				offDays := int(math.Floor(nowTime.Sub(onlineDate.UTC()).Hours() / 24))

				tag += template.HTML(`<span class="badge bg-gray">OFF: ` + strconv.Itoa(offDays) + ` Days </span>`)

				return tag
			}
			tag = template.HTML(`<span class='badge bg-green'>Yes</span>`)

			return tag
		})
	info.AddField(lg("PlayerStatus"), "isonline", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// fmt.Println(model.Row[`isonline`])

			tag := template.HTML(``)
			if model.Value == "0" {
				tag += template.HTML(`<span class="badge bg-gray">Disconnected</span>`)
			}

			if model.Value == "1" {
				tag += template.HTML(`<span class="badge bg-blue-gradient">In Lobby</span>`)
			}

			if model.Value == "2" {
				tag += template.HTML(`<span class="badge bg-red">Playing Game</span>`)
			}

			return tag
		})
	info.AddField(lg("Agent"), "agent", db.Varchar)
	info.AddField(lg("Balance"), "balance", db.Decimal)
	info.AddField(lg("Name"), "name", db.Varchar)
	info.AddField(lg("Operation"), "operation", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// buttons := template.HTML(`<td id="op1_0" class="text-left">`)
			buttons := template.HTML(`<button class="btn btn-info btn-xs"  onfocus="this.blur();" onclick="document.location='/Player/EditScore?id=` + model.ID + `'">set score</button>`)
			buttons += template.HTML(`<button class="btn btn-info btn-xs"  onfocus="this.blur();" onclick="document.location='/ScoreLog/SearchScoreLog?id=` + model.ID + `'">score log</button>`)
			buttons += template.HTML(`<button class="btn btn-info btn-xs"  onfocus="this.blur();" onclick="document.location='/Player/Edit?id=` + model.ID + `'">edit</button>`)
			buttons += template.HTML(`<button class="btn btn-info btn-xs"  onfocus="this.blur();" onclick="document.location='/Report/Search?id=` + model.ID + `'">report</button>`)
			buttons += template.HTML(`<button class="btn btn-info btn-xs"  onfocus="this.blur();" onclick="document.location='/GameLog/SearchGameLog?id=` + model.ID + `'">game log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SetRedPacket?id=` + model.ID + `'">set redpacket</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SearchRedPacketLog?id=` + model.ID + `'">redpacket log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Player/SendEventPlayer?playerid=` + model.ID + `'">send event</button>`)
			buttons += template.HTML(`<button class="btn btn-info btn-xs" name="forcequite" title="` + model.Row[`username`].(string) + `" player="` + model.ID + `" onfocus="this.blur();" onclick="">quit game</button>`)

			if model.Row[`state`] == 1 {
				buttons += template.HTML(`<button type="button" name="able" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="">disable</button>`)
			} else {
				buttons += template.HTML(`<button type="button" name="able" title="` + model.Row[`username`].(string) + `" rel ="enable" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">enable</button>`)
			}
			buttons += template.HTML(`<button type="button" name="stuckremove" title="` + model.Row[`username`].(string) + `" player="` + model.ID + `" class="btn btn-dropbox btn-xs" onfocus="this.blur();" onclick="">remove stuck</button>`)
			buttons += template.HTML(`<button type="button" name="closeonlineplayer" title="` + model.Row[`username`].(string) + `" player="` + model.ID + `" class="btn btn-dropbox btn-xs" onfocus="this.blur();" onclick="">closeonline</button>`)
			// buttons += template.HTML(`</td>`)

			return buttons
		})

	info.SetTable("Players").
		SetSortAsc()
	return
}

// added by jaison
func (s *SystemTable) GetAgentTable(ctx *context.Context) (agentTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetAgentTable")

	agentTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := agentTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldHide()
	info.AddField(lg("Username"), "username", db.Varchar)
	info.AddField(lg("Score"), "score", db.Decimal)
	info.AddField(lg("Name"), "name", db.Varchar)
	info.AddField(lg("Agent"), "agent", db.Varchar)
	info.AddField(lg("Operation"), "operation", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// fmt.Println(model)

			// buttons := template.HTML(`<td id="op2_0" class="text-left">`)
			buttons := template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Agent/EditScore?id=` + model.ID + `'">set score</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/ScoreLog/SearchScoreLog?sid=` + model.ID + `'">score log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Agent/Edit?id=` + model.ID + `'">edit</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Report/Search?sid=` + model.ID + `'">report</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/Report/Chart?sid=` + model.ID + `'">chart</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="">Total</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/BonusLog/SetBonus?sid=` + model.ID + `'">set bonus</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/BonusLog/SearchBonusLog?sid=` + model.ID + `'">bonus log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SetRedPacket?sid=` + model.ID + `'">set redpacket</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/RedPacketLog/SearchRedPacketLog?sid=` + model.ID + `'">redpacket log</button>`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" onfocus="this.blur();" onclick="document.location='/Agent/SendEventAgent?agentid=` + model.ID + `'">send event</button>`)

			if model.Row[`state`] == 1 {
				buttons += template.HTML(`<button type="button" name="agentable" title="` + model.Row[`username`].(string) + `" rel="disable" player="` + model.ID + `" class="btn btn-warning btn-xs" onfocus="this.blur();" onclick="">disable</button>`)
			} else {
				buttons += template.HTML(`<button type="button" name="agentable" title="` + model.Row[`username`].(string) + `" rel="enable" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">enable</button>`)
			}

			buttons += template.HTML(`<button type="button" name="closeonlineagent" title="` + model.Row[`username`].(string) + `" rel="closeonlineagent" player="` + model.ID + `" class="btn btn-danger btn-xs" onfocus="this.blur();" onclick="">closeonline</button>`)
			buttons += template.HTML(`<button type="button" name="closeonlineagentall" title="` + model.Row[`username`].(string) + `" rel="closeonlineagentall" player="` + model.ID + `" class="btn btn-yahoo btn-xs" onfocus="this.blur();" onclick="">closeonlineforall</button>`)
			// buttons += template.HTML(`</td>`)

			return buttons
		})

	info.SetTable("Agents").
		SetSortAsc()
	return
}

// added by jaison
func (s *SystemTable) GetInGamePlayers(ctx *context.Context) (inGamePlayersTable Table) {
	inGamePlayersTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	// #	UserName	Score	PlayerStatus	DisOnlineDay	Name	Tel	Description	SADescription
	info := inGamePlayersTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.SetSortAsc()
	info.AddField("#", "id", db.Int)
	info.AddField(lg("UserName"), "username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			// fmt.Println(model)

			tag := template.HTML(`<a href="/admin/Search/Index?username=` + model.Value + `&type=1">` + model.Value + `</a>`)

			return tag
		})
	info.AddField(lg("Name"), "name", db.Varchar)
	info.AddField(lg("Score"), "balance", db.Decimal)
	info.AddField(lg("PlayerStatus"), "isonline", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)
			if model.Value == "0" {
				tag += template.HTML(`<span class="badge bg-gray">Disconnected</span>`)
			}

			if model.Value == "1" {
				tag += template.HTML(`<span class="badge bg-blue-gradient">In Lobby</span>`)
			}

			if model.Value == "2" {
				tag += template.HTML(`<span class="badge bg-red">Playing Game</span>`)
			}

			return tag
		})

	info.AddField(lg("DisOnlineDay"), "online", db.Datetime).
		FieldDisplay(func(model types.FieldModel) interface{} {
			onlineDate, err := time.Parse(time.RFC3339, model.Value)

			tag := template.HTML(``)

			if err != nil || onlineDate.Year() <= 1 {
				tag += template.HTML(`<span class="badge bg-gray">NO LOGIN GAME</span>`)
				return tag
			}

			nowTime := time.Now().UTC()

			offDays := int(math.Floor(nowTime.Sub(onlineDate.UTC()).Hours() / 24))

			tag += template.HTML(`<span class="badge bg-light-blue">OFF: ` + strconv.Itoa(offDays) + ` Days </span>`)

			return tag
		})
	info.Where("isonline", "=", "2")

	info.HideDeleteButton()
	info.HideExportButton()

	info.HideNewButton()
	info.HideFilterButton()
	info.HideRowSelector()
	info.HideEditButton()
	info.HideDetailButton()

	info.SetTable("Players").
		SetHeaderHtml(template.HTML(`
			<h6 class="box-title" id="d_tip_2" style="font-size:0.9em;">
				<span class="badge bg-red hidden-sm" id="onlinecount"></span>
			</h6>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`))

	inGamePlayersTable.SetLoadFinishedCallBack(func(values ...interface{}) {
		fmt.Println("CallBack in GetInGamePlayers")

		if len(values) == 0 {
			return
		}

		info.SetHeaderHtml(template.HTML(`
			<h6 class="box-title" id="d_tip_2" style="font-size:0.9em;">
				<span class="badge bg-red hidden-sm" id="onlinecount">Online User Count - ` + strconv.Itoa(len(values[0].(PanelInfo).InfoList)) + ` players</span>
			</h6>
			<div class="box-tools pull-right">
				<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
			</div>`))
	})
	return
}

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
func (s *SystemTable) GetWinningPlayers(ctx *context.Context) (winningPlayersTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetWinningPlayers")
	config := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)
	config.PrimaryKey = PrimaryKey{
		Type: db.Int,
		Name: `username`,
	}

	winningPlayersTable = NewDefaultTable(config)

	// #	UserName	OnlineState	ReportStartTime	Total Bet	Total Win	Total Win - Total Bet
	info := winningPlayersTable.GetInfo().AddXssJsFilter().HideFilterArea().
		HideDeleteButton().
		HideDetailButton().
		HideEditButton().
		HideNewButton().
		HideRowSelector().
		HideExportButton().
		HideFilterButton()

	info.SetSortAsc()
	// info.AddField("#", "id", db.Int)
	info.AddField(lg("UserName"), "username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {

			tag := template.HTML(model.Row[`username`].(string))

			return tag
		})
	info.AddField(lg("OnlineState"), "isonline", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Row[`isonline`] == int64(0) {
				tag += template.HTML(`<span class="badge bg-gray">Disconnected</span>`)
				return tag
			}
			if model.Row[`isonline`] == int64(1) {
				tag += template.HTML(`<span class="badge bg-blue-gradient">In Lobby</span>`)
				return tag
			}

			tag += template.HTML(`<span class="badge bg-red">Playing Game</span>`)

			return tag
		})
	info.AddField(lg("Total Bet"), "BetField", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			betField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`BetField`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", betField))

			return tag
		})
	info.AddField(lg("Total Win"), "WinField", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			winField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`WinField`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", winField))

			return tag
		})
	info.AddField(lg("Profit(Total Win - Total Bet)"), "profit", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			winField, err1 := strconv.ParseFloat(ConvertInterface_A(model.Row[`WinField`]), 64)
			betField, err2 := strconv.ParseFloat(ConvertInterface_A(model.Row[`BetField`]), 64)

			tag := template.HTML(``)

			if err1 != nil || err2 != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", winField-betField))
			return tag
		})

	info.SetTable("Reports")
	return
}

// added by jaison
func (s *SystemTable) GetLoginLogs(ctx *context.Context) (loginLogsTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetLoginLogs")

	loginLogsTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := loginLogsTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldHide()
	info.AddField(lg("Username"), "username", db.Varchar)
	info.AddField(lg("IP"), "ip", db.Varchar)
	info.AddField(lg("DateTime"), "datetime", db.Varchar)

	info.SetTable("LoginLogs").
		SetTitle(lg("Player Login Logs")).
		SetSortField("datetime").
		SetSortDesc().
		AddLimitFilter(10)
	return
}

// added by jaison
func (s *SystemTable) GetMemberOutstandings(ctx *context.Context) (reportTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetMemberOutstandings")

	reportTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := reportTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldHide()
	info.AddField(lg("Login Name"), "username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {

			tag := template.HTML("<a>" + model.Row[`Username`].(string) + "</a>")

			return tag
		})
	info.AddField(lg("Position"), "name", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			return template.HTML(`Shareholder`)
		})
	info.AddField(lg("Bet"), "bet", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			betField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`bet`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", betField))

			return tag
		})
	info.AddField(lg("Win"), "win", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			winField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`win`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", winField))

			return tag
		})
	info.AddField(lg("Turnover"), "turnover", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			reportField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`turnover`]), 64)

			tag := template.HTML(``)

			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", reportField))
			return tag
		})

	info.SetTable("Reports").
		SetSortAsc()
	return
}

// added by jaison
func (s *SystemTable) GetAgentScoresTable(ctx *context.Context) (agentTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetAgentScoresTable")

	agentTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := agentTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldHide()
	info.AddField(lg("Login Name"), "username", db.Varchar)
	info.AddField(lg("Nickname"), "name", db.Varchar)
	info.AddField(lg("Level"), "level", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			return template.HTML(`Shareholder`)
		})
	// info.AddField(lg("Currency"), "setscore", db.Decimal).
	// 	FieldDisplay(func(model types.FieldModel) interface{} {
	// 		tag := template.HTML(model.Row[`username`].(string))
	// 		return tag
	// 	})
	info.AddField(lg("Credit"), "score", db.Decimal)
	info.AddField(lg("Deposit"), "deposit", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			buttons := template.HTML(`<div align="center">`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/scorelog/agentscores/EditScore?id=` + model.ID + `'" style="width:40px;"> + </button>`)
			buttons += template.HTML(`</div>`)

			return buttons
		})
	info.AddField(lg("Withdrawal"), "withdrawl", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			buttons := template.HTML(`<div align="center">`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/scorelog/agentscores/EditScore?id=` + model.ID + `'" style="width:40px;"> - </button>`)
			buttons += template.HTML(`</div>`)

			return buttons
		})
	info.AddField(lg("Detail"), "detail", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			buttons := template.HTML(`<div align="center">`)
			buttons += template.HTML(`<button type="button" class="btn btn-info btn-xs" title="" onfocus="this.blur();" onclick="document.location='/scorelog/agentscores/Detail?id=` + model.ID + `'" style="width:50px;"> Detail </button>`)
			buttons += template.HTML(`</div>`)

			return buttons
		})
	info.AddField(lg("Last login date"), "lastlogin", db.Varchar)
	info.AddField(lg("Last login IP"), "lastloginip", db.Varchar)

	info.SetTable("Agents").
		SetSortAsc()
	return
}

// added by jaison
func (s *SystemTable) GetScoreLogs(ctx *context.Context) (scoreLogsTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetScoreLogs")

	scoreLogsTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := scoreLogsTable.GetInfo().AddXssJsFilter().HideFilterArea()

	// Account UserName SetScore BeforeScore AfterScore IP DateTime
	info.AddField("#", "id", db.Int).FieldHide()
	info.AddField(lg("Login Name"), "username", db.Varchar)
	// info.AddField(lg("Currency"), "setscore", db.Decimal).
	// 	FieldDisplay(func(model types.FieldModel) interface{} {
	// 		tag := template.HTML(model.Row[`username`].(string))
	// 		return tag
	// 	})
	info.AddField(lg("Action"), "action", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)
			scoredValue, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`setscore`]), 64)

			if err != nil {
				tag = template.HTML(err.Error())
			} else if scoredValue > 0 {
				tag = template.HTML(`DEPOSIT`)
			} else if scoredValue < 0 {
				tag = template.HTML(`WITHDRAW`)
			}

			return tag
		})
	info.AddField(lg("Amount"), "setscore", db.Decimal)
	info.AddField(lg("Request By"), "account", db.Varchar)
	// info.AddField(lg("BeforeScore"), "beforescore", db.Decimal)
	// info.AddField(lg("AfterScore"), "afterscore", db.Decimal)
	// info.AddField(lg("IP"), "ip", db.Varchar)
	info.AddField(lg("Date"), "datetime", db.Datetime)

	info.SetTable("ScoreLogs").
		SetSortField("datetime").
		SetSortDesc()

	return
}

// added by jaison
func (s *SystemTable) GetBonusLogs(ctx *context.Context) (bonusLogsTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetBonusLogs")

	bonusLogsTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := bonusLogsTable.GetInfo().AddXssJsFilter().HideFilterArea()

	// #	Agent Name	Bonus	Processed	RegDate	Game Name	Player Name	Type	ProcessedDate
	info.AddField("#", "id", db.Int).FieldHide()
	info.AddField(lg("Agent Name"), "agentname", db.Varchar)
	info.AddField(lg("Bonus"), "bonus", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			bonus, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`bonus`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", bonus))

			return tag
		})
	info.AddField(lg("Processed"), "processed", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Value == "0" {
				tag += template.HTML(`Not Processed`)
			} else {
				tag += template.HTML(`Processed`)
			}

			return tag
		})
	info.AddField(lg("RegDate"), "regdate", db.Datetime)
	info.AddField(lg("Game Name"), "gamename", db.Varchar)
	info.AddField(lg("Player Name"), "playerid", db.Integer)
	info.AddField(lg("Type"), "bonustype", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Value == "0" {
				tag += template.HTML(`Random`)
				return tag
			}

			if model.Value == "1" {
				tag += template.HTML(`Minor`)
				return tag
			}

			if model.Value == "2" {
				tag += template.HTML(`Major`)
				return tag
			}

			tag += template.HTML(`Undefined`)
			return tag
		})
	info.AddField(lg("ProcessedDate"), "processeddate", db.Varchar)

	info.SetTable("BonusLogs").
		// SetSortField("datetime").
		SetSortDesc()

	return
}

// added by jaison
func (s *SystemTable) GetRedPacketLogs(ctx *context.Context) (redPacketLogsTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetRedPacketLogs")

	redPacketLogsTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := redPacketLogsTable.GetInfo().AddXssJsFilter().HideFilterArea()

	// #	Agent Name	RedPacket	Comment	Processed	RegDate		Player Name		ProcessedDate
	info.AddField("#", "id", db.Int).FieldHide()
	info.AddField(lg("Agent Name"), "agentname", db.Varchar)
	info.AddField(lg("RedPacket"), "redpacket", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			redpacket, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`redpacket`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", redpacket))

			return tag
		})
	info.AddField(lg("Comment"), "comment", db.Varchar)
	info.AddField(lg("Processed"), "processed", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)

			if model.Value == "0" {
				tag += template.HTML(`Not Processed`)
			} else {
				tag += template.HTML(`Processed`)
			}

			return tag
		})
	info.AddField(lg("RegDate"), "regdate", db.Datetime)
	info.AddField(lg("Player Name"), "playername", db.Varchar)
	info.AddField(lg("ProcessedDate"), "processeddate", db.Varchar)

	info.SetTable("RedPacketLogs")
	// SetHeader(`
	// 	<h3 class="box-title text-bold">
	// 		<span id="td_currMoney" class="badge bg-yellow"></span>
	// 		<span id="s_tip1" class="text-sm text-success" style=""></span>
	// 	</h3>
	// 	<div class="box-tools pull-right">
	// 		<button data-widget="collapse" class="btn btn-box-tool" type="button"><i class="fa fa-minus"></i></button>
	// 	</div>`)

	// SetSortField("datetime").
	// SetSortDesc()

	return
}

// added by jaison
func (s *SystemTable) GetPlayerReportLogs(ctx *context.Context) (playerReportLogsTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetPlayerReportLogs")

	playerReportLogsTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := playerReportLogsTable.GetInfo().AddXssJsFilter().HideFilterArea()

	// #	DateTime TotalActivePlayer	WinPlayers	WinAmount	LosePlayers	LoseAmount
	info.AddField("#", "id", db.Bigint).FieldHide()
	info.AddField(lg("DateTime"), "datetime", db.Datetime)
	info.AddField(lg("TotalActivePlayer"), "totalplayers", db.Integer)
	info.AddField(lg("WinPlayers"), "winplayers", db.Integer)
	info.AddField(lg("WinAmount"), "winamount", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			bonus, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`bonus`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", bonus))

			return tag
		})
	info.AddField(lg("LosePlayers"), "lostplayers", db.Integer)
	info.AddField(lg("LoseAmount"), "lostamount", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			bonus, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`bonus`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", bonus))

			return tag
		})

	info.SetTable("BonusLogs").
		// SetSortField("datetime").
		SetSortDesc()

	return
}

// added by jaison
func (s *SystemTable) GetTopWinPlayers(ctx *context.Context) (topWinUsers Table) {
	fmt.Println("plugins.modules.table.generator.go GetTopWinPlayers")

	topWinUsers = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := topWinUsers.GetInfo().AddXssJsFilter().HideFilterArea()

	// username Bet Win Report
	info.AddField("#", "id", db.Int).FieldHide()
	info.AddField(lg("UserName"), "username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(model.Row[`username`].(string))

			return tag
		})
	info.AddField(lg("Bet"), "Bet", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			bet, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Bet`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", bet))

			return tag
		})
	info.AddField(lg("Win"), "Win", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			win, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Win`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", win))

			return tag
		})
	info.AddField(lg("Report"), "Report", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			report, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Report`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", report))

			return tag
		})

	info.SetTable("BonusLogs").
		SetSortField("Report").
		SetSortAsc()

	return
}

// added by jaison
func (s *SystemTable) GetTopLostPlayers(ctx *context.Context) (topLoseUsers Table) {
	fmt.Println("plugins.modules.table.generator.go GetTopLostPlayers")

	topLoseUsers = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := topLoseUsers.GetInfo().AddXssJsFilter().HideFilterArea()

	// username Bet Win Report
	info.AddField("#", "id", db.Int).FieldHide()
	info.AddField(lg("UserName"), "username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(model.Row[`username`].(string))

			return tag
		})
	info.AddField(lg("Bet"), "Bet", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			bet, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Bet`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", bet))

			return tag
		})
	info.AddField(lg("Win"), "Win", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			win, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Win`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", win))

			return tag
		})
	info.AddField(lg("Report"), "Report", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			report, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Report`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", report))

			return tag
		})

	info.SetTable("BonusLogs").
		SetSortField("Report").
		SetSortDesc()

	return
}

// added by jaison
func (s *SystemTable) GetAgentReportLogs(ctx *context.Context) (agentReportTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetAgentReportLogs")

	config := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)
	config.PrimaryKey = PrimaryKey{
		Type: db.Int,
		Name: `Username`,
	}

	agentReportTable = NewDefaultTable(config)

	info := agentReportTable.GetInfo().AddXssJsFilter().HideFilterArea().
		HideDeleteButton().
		HideDetailButton().
		HideEditButton().
		HideNewButton().
		HideRowSelector().
		HideExportButton().
		HideFilterButton()

	// AgentName Bet Win Report
	// info.AddField("#", "id", db.Int)
	info.AddField(lg("UserName"), "Username", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {

			tag := template.HTML(`<a href="/Search/Index?username='` + model.Row[`Username`].(string) + `'&type=1">` + model.Row[`Username`].(string) + `</a>`)

			return tag
		})
	info.AddField(lg("Bet"), "Bet", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			betField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Bet`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", betField))

			return tag
		})
	info.AddField(lg("Win"), "Win", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			winField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Win`]), 64)

			tag := template.HTML(``)
			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", winField))

			return tag
		})
	info.AddField(lg("Report"), "Report", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			reportField, err := strconv.ParseFloat(ConvertInterface_A(model.Row[`Report`]), 64)

			tag := template.HTML(``)

			if err != nil {
				tag += template.HTML(`Failed to parse value.`)
				return tag
			}

			tag = template.HTML(fmt.Sprintf("%.2f", reportField))
			return tag
		})

	info.SetTable("Reports").
		SetSortField("Report").
		SetSortAsc()

	return
}

// added by jaison
func (s *SystemTable) GetGameConfigs(ctx *context.Context) (gameConfigsTable Table) {
	fmt.Println("plugins.modules.table.generator.go GetGameConfigs")
	config := DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver)
	// config.PrimaryKey = PrimaryKey{
	// 	Type: db.Int,
	// 	Name: `username`,
	// }

	gameConfigsTable = NewDefaultTable(config)

	info := gameConfigsTable.GetInfo().AddXssJsFilter().HideFilterArea().
		HideDeleteButton().
		HideDetailButton().
		HideEditButton().
		HideNewButton().
		HideRowSelector().
		HideExportButton().
		HideFilterButton()

	info.SetSortAsc()

	// #	Game Type	Game Name	GameID	PayoutReset	PayoutRate	EventRate	FreeSpinWinRate	RandomBonusLimit	AllowOnlineTableCustomize	AllowOpenClose	Operation
	info.AddField("#", "id", db.Int)
	info.AddField(lg("Game Type"), "gametype", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(``)
			if model.Value == "1" {
				tag += template.HTML(`<span style="color:blueviolet">Single Game</span>`)
			} else {
				tag += template.HTML(`<span style="color:red">Live Game</span>`)
			}

			return tag
		})
	info.AddField(lg("Game Name"), "gamename", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(model.Value)

			if model.Row[`gameid`].(int64) == int64(29) {
				tag += template.HTML(`<span style="color:red;">` + model.Value + `</span>`)
			}
			if model.Row[`gameid`].(int64) == int64(57) {
				tag += template.HTML(`<span style="color:green">` + model.Value + `</span>`)
			}
			if model.Row[`gameid`].(int64) == int64(141) {
				tag += template.HTML(`<span style="color:blue">` + model.Value + `</span>`)
			}

			return tag
		})
	info.AddField(lg("GameID"), "gameid", db.Varchar)
	info.AddField(lg("PayoutReset"), "username", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`
				<div>
					<input type="text" name="resetpercent" maxlength="8" style="width: 40px; vertical-align: middle; text-align: center" />
					<button class="btn btn-primary btn-xs" name="payout_reset" data-gameid="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" style="vertical-align: middle">ResetPayout</button>
				</div>`)

			return tag
		})
	info.AddField(lg("PayoutRate"), "winchance", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`
				<div>
					<input type="text" value="` + model.Value + `" name="percent" maxlength="8" style="width: 50px; vertical-align: middle; text-align: center" />&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
					<button class="btn btn-info btn-xs" name="game_update" style="vertical-align: middle">Update</button>
					<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">
				</div>`)

			return tag
		})
	info.AddField(lg("HasEvent"), "hasevent", db.Integer).FieldHide()
	info.AddField(lg("EventRate"), "eventrate", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`<div>`)

			if model.Row[`hasevent`].(int64) == int64(1) {
				tag += template.HTML(`<input type="checkbox" name="hasevent" style="vertical-align: middle; display: none" id="hasevent` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" checked="checked" />`)
			} else {
				tag += template.HTML(`<input type="checkbox" name="hasevent" style="vertical-align: middle; display: none" id="hasevent` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" />`)
			}

			tag += template.HTML(`<input type="text" name="eventrate" value="` + model.Value + `" id="eventrate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" style="width: 50px; vertical-align: middle; text-align: center" />&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`)
			tag += template.HTML(`<button class="btn btn-info btn-xs" name="eventrate_update" style="vertical-align: middle" id="eventrateupdate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `">Update</button>`)
			tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
			tag += template.HTML(`</div>`)

			return tag
		})
	info.AddField(lg("HasFreespinWinrate"), "hasfreespinwinrate", db.Integer).FieldHide()
	info.AddField(lg("FreeSpinWinRate"), "freespinwinrate", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`<div>`)

			if model.Row[`hasfreespinwinrate`].(int64) == int64(1) {
				tag += template.HTML(`<input type="checkbox" name="hasfreespinwinrate" style="vertical-align: middle; display: none" id="hasfreespinwinrate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" checked="checked" />`)
			} else {
				tag += template.HTML(`<input type="checkbox" name="hasfreespinwinrate" style="vertical-align: middle; display: none" id="hasfreespinwinrate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" />`)
			}

			tag += template.HTML(`<input type="text" name="freespinwinrate" value="` + model.Value + `" id="freespinwinrate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" style="width: 50px; vertical-align: middle; text-align: center" />&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`)
			tag += template.HTML(`<button class="btn btn-info btn-xs" name="freespinwinrate_update" style="vertical-align: middle" id="freespinwinrateupdate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `">Update</button>`)
			tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
			tag += template.HTML(`</div>`)

			return tag
		})
	info.AddField(lg("HasRandomBonusLimit"), "hasrandombonuslimit", db.Integer).FieldHide()
	info.AddField(lg("RandomBonusLimit"), "randombonuslimit", db.Decimal).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`<div>`)

			if model.Row[`hasrandombonuslimit`].(int64) == int64(1) {
				tag += template.HTML(`<input type="checkbox" name="hasrandombonuslimit" style="vertical-align: middle; display: none" id="hasrandombonuslimit` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" checked="checked" />`)
			} else {
				tag += template.HTML(`<input type="checkbox" name="hasrandombonuslimit" style="vertical-align: middle; display: none" id="hasrandombonuslimit` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" />`)
			}

			tag += template.HTML(`<input type="text" name="randombonuslimit" value="` + model.Value + `" id="randombonuslimit` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" style="width: 50px; vertical-align: middle; text-align: center" />&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`)
			tag += template.HTML(`<button class="btn btn-info btn-xs" name="randombonuslimit_update" style="vertical-align: middle" id="randombonuslimitupdate` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `">Update</button>`)
			tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
			tag += template.HTML(`</div>`)

			return tag
		})
	info.AddField(lg("TableSet"), "tableset", db.Integer).FieldHide()
	info.AddField(lg("AllowOnlineTableCustomize"), "customize", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`<div>`)

			if model.Row[`gametype`].(int64) == int64(1) {
				tag += template.HTML(`<input type="checkbox" name="dynamictable" style="vertical-align:middle;" disabled/>`)
				tag += template.HTML(`<button class="btn btn-info btn-xs" name="dynamictable_update" style="vertical-align: middle" disabled id="dynamictable_update` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `">Update</button>`)
			} else {
				if model.Row[`tableset`].(int64) == int64(1) {
					tag += template.HTML(`<input type="checkbox" name="dynamictable" style="vertical-align:middle;" checked="checked" />`)
				} else {
					tag += template.HTML(`<input type="checkbox" name="dynamictable" style="vertical-align:middle;" />`)
				}
				tag += template.HTML(`<button class="btn btn-info btn-xs" name="dynamictable_update" style="vertical-align: middle" id="dynamictable_update` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `">Update</button>`)
			}

			tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
			tag += template.HTML(`</div>`)

			return tag
		})
	info.AddField(lg("AllowOpenClose"), "canclose", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`<div>`)

			if model.Value == "1" {
				tag += template.HTML(`<input type="checkbox" name="cancloseset" style="vertical-align:middle;" checked="checked" />`)
			} else {
				tag += template.HTML(`<input type="checkbox" name="cancloseset" style="vertical-align:middle;" />`)
			}

			tag += template.HTML(`<button class="btn btn-info btn-xs" name="canclose_update" style="vertical-align: middle" id="canclose_update` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `">Update</button>`)
			tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
			tag += template.HTML(`</div>`)

			return tag
		})
	info.AddField(lg("Operation"), "openclose", db.Integer).
		FieldDisplay(func(model types.FieldModel) interface{} {
			tag := template.HTML(`<div>`)

			if model.Value == "1" {
				tag += template.HTML(`<button type="button" class="btn btn-yahoo btn-xs" id="openclose` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="openclose" value="1" style="color: white; border-color: none">Close</button>`)
				tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
				tag += template.HTML(`<button type="button" class="btn btn-danger btn-xs" id="del` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="del">Delete</button>`)
			} else {
				tag += template.HTML(`<button type="button" class="btn btn-facebook btn-xs" id="openclose` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="openclose" value="0" style="color: white; border-color: none">Open</button>`)
				tag += template.HTML(`<input type="hidden" value="` + strconv.FormatInt(model.Row[`id`].(int64), 10) + `" name="game_id">`)
				tag += template.HTML(`<button type="button" class="btn btn-danger btn-xs" name="del">Delete</button>`)
			}

			tag += template.HTML(`</div>`)

			return tag
		})

	info.SetTable("Configs")
	return
}

func (s *SystemTable) GetManagerTable(ctx *context.Context) (managerTable Table) {
	managerTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := managerTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("Name"), "username", db.Varchar).FieldFilterable()
	info.AddField(lg("Nickname"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("role"), "name", db.Varchar).
		FieldJoin(types.Join{
			Table:     "goadmin_role_users",
			JoinField: "user_id",
			Field:     "id",
		}).
		FieldJoin(types.Join{
			Table:     "goadmin_roles",
			JoinField: "id",
			Field:     "role_id",
			BaseTable: "goadmin_role_users",
		}).
		FieldDisplay(func(model types.FieldModel) interface{} {
			labels := template.HTML("")
			labelTpl := label().SetType("success")

			labelValues := strings.Split(model.Value, types.JoinFieldValueDelimiter)
			for key, label := range labelValues {
				if key == len(labelValues)-1 {
					labels += labelTpl.SetContent(template.HTML(label)).GetContent()
				} else {
					labels += labelTpl.SetContent(template.HTML(label)).GetContent() + "<br><br>"
				}
			}

			if labels == template.HTML("") {
				return lg("no roles")
			}

			return labels
		}).FieldFilterable()
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_users").
		SetTitle(lg("Managers")).
		SetDescription(lg("Managers")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)

			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteUserRoleErr := s.connection().WithTx(tx).
					Table("goadmin_role_users").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}

				deleteUserPermissionErr := s.connection().WithTx(tx).
					Table("goadmin_user_permissions").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserPermissionErr, db.DELETE) {
					return deleteUserPermissionErr, nil
				}

				deleteUserErr := s.connection().WithTx(tx).
					Table("goadmin_users").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteUserErr, db.DELETE) {
					return deleteUserErr, nil
				}

				return nil, nil
			})

			return txErr
		})

	formList := managerTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("Name"), "username", db.Varchar, form.Text).
		FieldHelpMsg(template.HTML(lg("use for login"))).FieldMust()
	formList.AddField(lg("Nickname"), "name", db.Varchar, form.Text).
		FieldHelpMsg(template.HTML(lg("use to display"))).FieldMust()
	formList.AddField(lg("Avatar"), "avatar", db.Varchar, form.File)
	formList.AddField(lg("role"), "role_id", db.Varchar, form.Select).
		FieldOptionsFromTable("goadmin_roles", "slug", "id").
		FieldDisplay(func(model types.FieldModel) interface{} {
			var roles []string

			if model.ID == "" {
				return roles
			}
			roleModel, _ := s.table("goadmin_role_users").Select("role_id").
				Where("user_id", "=", model.ID).All()
			for _, v := range roleModel {
				roles = append(roles, strconv.FormatInt(v["role_id"].(int64), 10))
			}
			return roles
		}).FieldHelpMsg(template.HTML(lg("no corresponding options?")) +
		link("/admin/info/roles/new", "Create here."))

	formList.AddField(lg("permission"), "permission_id", db.Varchar, form.Select).
		FieldOptionsFromTable("goadmin_permissions", "slug", "id").
		FieldDisplay(func(model types.FieldModel) interface{} {
			var permissions []string

			if model.ID == "" {
				return permissions
			}
			permissionModel, _ := s.table("goadmin_user_permissions").
				Select("permission_id").Where("user_id", "=", model.ID).All()
			for _, v := range permissionModel {
				permissions = append(permissions, strconv.FormatInt(v["permission_id"].(int64), 10))
			}
			return permissions
		}).FieldHelpMsg(template.HTML(lg("no corresponding options?")) +
		link("/admin/info/permission/new", "Create here."))

	formList.AddField(lg("password"), "password", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		})
	formList.AddField(lg("confirm password"), "password_again", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		})

	formList.SetTable("goadmin_users").SetTitle(lg("Managers")).SetDescription(lg("Managers"))
	formList.SetUpdateFn(func(values form2.Values) error {

		if values.IsEmpty("name", "username") {
			return errors.New("username and password can not be empty")
		}

		user := models.UserWithId(values.Get("id")).SetConn(s.conn)

		password := values.Get("password")

		if password != "" {

			if password != values.Get("password_again") {
				return errors.New("password does not match")
			}

			password = encodePassword([]byte(values.Get("password")))
		}

		_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

			_, updateUserErr := user.WithTx(tx).Update(values.Get("username"), password, values.Get("name"), values.Get("avatar"))

			if db.CheckError(updateUserErr, db.UPDATE) {
				return updateUserErr, nil
			}

			delRoleErr := user.WithTx(tx).DeleteRoles()

			if db.CheckError(delRoleErr, db.DELETE) {
				return delRoleErr, nil
			}

			for i := 0; i < len(values["role_id[]"]); i++ {
				_, addRoleErr := user.WithTx(tx).AddRole(values["role_id[]"][i])
				if db.CheckError(addRoleErr, db.INSERT) {
					return addRoleErr, nil
				}
			}

			delPermissionErr := user.WithTx(tx).DeletePermissions()

			if db.CheckError(delPermissionErr, db.DELETE) {
				return delPermissionErr, nil
			}

			for i := 0; i < len(values["permission_id[]"]); i++ {
				_, addPermissionErr := user.WithTx(tx).AddPermission(values["permission_id[]"][i])
				if db.CheckError(addPermissionErr, db.INSERT) {
					return addPermissionErr, nil
				}
			}

			return nil, nil
		})

		return txErr
	})
	// formList.SetInsertFn(func(values form2.Values) error {
	// 	if values.IsEmpty("name", "username", "password") {
	// 		return errors.New("username and password can not be empty")
	// 	}

	// 	password := values.Get("password")

	// 	if password != values.Get("password_again") {
	// 		return errors.New("password does not match")
	// 	}

	// 	_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

	// 		user, createUserErr := models.User().WithTx(tx).SetConn(s.conn).New(values.Get("username"),
	// 			encodePassword([]byte(values.Get("password"))),
	// 			values.Get("name"),
	// 			values.Get("avatar"))

	// 		if db.CheckError(createUserErr, db.INSERT) {
	// 			return createUserErr, nil
	// 		}

	// 		for i := 0; i < len(values["role_id[]"]); i++ {
	// 			_, addRoleErr := user.WithTx(tx).AddRole(values["role_id[]"][i])
	// 			if db.CheckError(addRoleErr, db.INSERT) {
	// 				return addRoleErr, nil
	// 			}
	// 		}

	// 		for i := 0; i < len(values["permission_id[]"]); i++ {
	// 			_, addPermissionErr := user.WithTx(tx).AddPermission(values["permission_id[]"][i])
	// 			if db.CheckError(addPermissionErr, db.INSERT) {
	// 				return addPermissionErr, nil
	// 			}
	// 		}

	// 		return nil, nil
	// 	})
	// 	return txErr
	// })

	detail := managerTable.GetDetail()
	detail.AddField("ID", "id", db.Int)
	detail.AddField(lg("Name"), "username", db.Varchar)
	detail.AddField(lg("Avatar"), "avatar", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			if model.Value == "" || config.GetStore().Prefix == "" {
				model.Value = config.Url("/assets/dist/img/avatar04.png")
			} else {
				model.Value = config.GetStore().URL(model.Value)
			}
			return template.Default().Image().
				SetSrc(template.HTML(model.Value)).
				SetHeight("120").SetWidth("120").WithModal().GetContent()
		})
	detail.AddField(lg("Nickname"), "name", db.Varchar)
	detail.AddField(lg("role"), "roles", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			labelModels, _ := s.table("goadmin_role_users").
				Select("goadmin_roles.name").
				LeftJoin("goadmin_roles", "goadmin_roles.id", "=", "goadmin_role_users.role_id").
				Where("user_id", "=", model.ID).
				All()

			labels := template.HTML("")
			labelTpl := label().SetType("success")

			for key, label := range labelModels {
				if key == len(labelModels)-1 {
					labels += labelTpl.SetContent(template.HTML(label["name"].(string))).GetContent()
				} else {
					labels += labelTpl.SetContent(template.HTML(label["name"].(string))).GetContent() + "<br><br>"
				}
			}

			if labels == template.HTML("") {
				return lg("no roles")
			}

			return labels
		})
	detail.AddField(lg("permission"), "roles", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			permissionModel, _ := s.table("goadmin_user_permissions").
				Select("goadmin_permissions.name").
				LeftJoin("goadmin_permissions", "goadmin_permissions.id", "=", "goadmin_user_permissions.permission_id").
				Where("user_id", "=", model.ID).
				All()

			permissions := template.HTML("")
			permissionTpl := label().SetType("success")

			for key, label := range permissionModel {
				if key == len(permissionModel)-1 {
					permissions += permissionTpl.SetContent(template.HTML(label["name"].(string))).GetContent()
				} else {
					permissions += permissionTpl.SetContent(template.HTML(label["name"].(string))).GetContent() + "<br><br>"
				}
			}

			return permissions
		})
	detail.AddField(lg("createdAt"), "created_at", db.Timestamp)
	detail.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	return
}

func (s *SystemTable) GetNormalManagerTable(ctx *context.Context) (managerTable Table) {
	managerTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := managerTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("Name"), "username", db.Varchar).FieldFilterable()
	info.AddField(lg("Nickname"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("role"), "name", db.Varchar).
		FieldJoin(types.Join{
			Table:     "goadmin_role_users",
			JoinField: "user_id",
			Field:     "id",
		}).
		FieldJoin(types.Join{
			Table:     "goadmin_roles",
			JoinField: "id",
			Field:     "role_id",
			BaseTable: "goadmin_role_users",
		}).
		FieldDisplay(func(model types.FieldModel) interface{} {
			labels := template.HTML("")
			labelTpl := label().SetType("success")

			labelValues := strings.Split(model.Value, types.JoinFieldValueDelimiter)
			for key, label := range labelValues {
				if key == len(labelValues)-1 {
					labels += labelTpl.SetContent(template.HTML(label)).GetContent()
				} else {
					labels += labelTpl.SetContent(template.HTML(label)).GetContent() + "<br><br>"
				}
			}

			if labels == template.HTML("") {
				return lg("no roles")
			}

			return labels
		})
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_users").
		SetTitle(lg("Managers")).
		SetDescription(lg("Managers")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)

			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteUserRoleErr := s.connection().WithTx(tx).
					Table("goadmin_role_users").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserRoleErr, db.DELETE) {
					return deleteUserRoleErr, nil
				}

				deleteUserPermissionErr := s.connection().WithTx(tx).
					Table("goadmin_user_permissions").
					WhereIn("user_id", ids).
					Delete()

				if db.CheckError(deleteUserPermissionErr, db.DELETE) {
					return deleteUserPermissionErr, nil
				}

				deleteUserErr := s.connection().WithTx(tx).
					Table("goadmin_users").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteUserErr, db.DELETE) {
					return deleteUserErr, nil
				}

				return nil, nil
			})

			return txErr
		})

	formList := managerTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("Name"), "username", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("use for login"))).FieldMust()
	formList.AddField(lg("Nickname"), "name", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("use to display"))).FieldMust()
	formList.AddField(lg("Avatar"), "avatar", db.Varchar, form.File)
	formList.AddField(lg("password"), "password", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		})
	formList.AddField(lg("confirm password"), "password_again", db.Varchar, form.Password).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return ""
		})

	formList.SetTable("goadmin_users").SetTitle(lg("Managers")).SetDescription(lg("Managers"))
	formList.SetUpdateFn(func(values form2.Values) error {

		if values.IsEmpty("name", "username") {
			return errors.New("username and password can not be empty")
		}

		user := models.UserWithId(values.Get("id")).SetConn(s.conn)

		if values.Has("permission", "role") {
			return errors.New(errs.NoPermission)
		}

		password := values.Get("password")

		if password != "" {

			if password != values.Get("password_again") {
				return errors.New("password does not match")
			}

			password = encodePassword([]byte(values.Get("password")))
		}

		_, updateUserErr := user.Update(values.Get("username"), password, values.Get("name"), values.Get("avatar"))

		if db.CheckError(updateUserErr, db.UPDATE) {
			return updateUserErr
		}

		return nil
	})
	// formList.SetInsertFn(func(values form2.Values) error {
	// 	if values.IsEmpty("name", "username", "password") {
	// 		return errors.New("username and password can not be empty")
	// 	}

	// 	password := values.Get("password")

	// 	if password != values.Get("password_again") {
	// 		return errors.New("password does not match")
	// 	}

	// 	if values.Has("permission", "role") {
	// 		return errors.New(errs.NoPermission)
	// 	}

	// 	_, createUserErr := models.User().SetConn(s.conn).New(values.Get("username"),
	// 		encodePassword([]byte(values.Get("password"))),
	// 		values.Get("name"),
	// 		values.Get("avatar"))

	// 	if db.CheckError(createUserErr, db.INSERT) {
	// 		return createUserErr
	// 	}

	// 	return nil
	// })

	return
}

func (s *SystemTable) GetPermissionTable(ctx *context.Context) (permissionTable Table) {
	permissionTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := permissionTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("permission"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("slug"), "slug", db.Varchar).FieldFilterable()
	info.AddField(lg("method"), "http_method", db.Varchar).FieldDisplay(func(value types.FieldModel) interface{} {
		if value.Value == "" {
			return "All methods"
		}
		return value.Value
	})
	info.AddField(lg("path"), "http_path", db.Varchar).
		FieldDisplay(func(model types.FieldModel) interface{} {
			pathArr := strings.Split(model.Value, "\n")
			res := ""
			for i := 0; i < len(pathArr); i++ {
				if i == len(pathArr)-1 {
					res += string(label().SetContent(template.HTML(pathArr[i])).GetContent())
				} else {
					res += string(label().SetContent(template.HTML(pathArr[i])).GetContent()) + "<br><br>"
				}
			}
			return res
		})
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_permissions").
		SetTitle(lg("Permission Manage")).
		SetDescription(lg("Permission Manage")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)

			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteRolePermissionErr := s.connection().WithTx(tx).
					Table("goadmin_role_permissions").
					WhereIn("permission_id", ids).
					Delete()

				if db.CheckError(deleteRolePermissionErr, db.DELETE) {
					return deleteRolePermissionErr, nil
				}

				deleteUserPermissionErr := s.connection().WithTx(tx).
					Table("goadmin_user_permissions").
					WhereIn("permission_id", ids).
					Delete()

				if db.CheckError(deleteUserPermissionErr, db.DELETE) {
					return deleteUserPermissionErr, nil
				}

				deletePermissionsErr := s.connection().WithTx(tx).
					Table("goadmin_permissions").
					WhereIn("id", ids).
					Delete()

				if deletePermissionsErr != nil {
					return deletePermissionsErr, nil
				}

				return nil, nil
			})

			return txErr
		})

	formList := permissionTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("permission"), "name", db.Varchar, form.Text).FieldMust()
	formList.AddField(lg("slug"), "slug", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("should be unique"))).FieldMust()
	formList.AddField(lg("method"), "http_method", db.Varchar, form.Select).
		FieldOptions(types.FieldOptions{
			{Value: "GET", Text: "GET"},
			{Value: "PUT", Text: "PUT"},
			{Value: "POST", Text: "POST"},
			{Value: "DELETE", Text: "DELETE"},
			{Value: "PATCH", Text: "PATCH"},
			{Value: "OPTIONS", Text: "OPTIONS"},
			{Value: "HEAD", Text: "HEAD"},
		}).
		FieldDisplay(func(model types.FieldModel) interface{} {
			return strings.Split(model.Value, ",")
		}).
		FieldPostFilterFn(func(model types.PostFieldModel) interface{} {
			return strings.Join(model.Value, ",")
		}).
		FieldHelpMsg(template.HTML(lg("all method if empty")))

	formList.AddField(lg("path"), "http_path", db.Varchar, form.TextArea).
		FieldPostFilterFn(func(model types.PostFieldModel) interface{} {
			return strings.TrimSpace(model.Value.Value())
		}).
		FieldHelpMsg(template.HTML(lg("a path a line, without global prefix")))
	formList.AddField(lg("updatedAt"), "updated_at", db.Timestamp, form.Default).FieldNotAllowAdd()
	formList.AddField(lg("createdAt"), "created_at", db.Timestamp, form.Default).FieldNotAllowAdd()

	formList.SetTable("goadmin_permissions").
		SetTitle(lg("Permission Manage")).
		SetDescription(lg("Permission Manage")).
		SetPostValidator(func(values form2.Values) error {

			if values.IsEmpty("slug", "http_path", "name") {
				return errors.New("slug or http_path or name should not be empty")
			}

			if models.Permission().SetConn(s.conn).IsSlugExist(values.Get("slug"), values.Get("id")) {
				return errors.New("slug exists")
			}
			return nil
		}).SetPostHook(func(values form2.Values) error {
		_, err := s.connection().Table("goadmin_permissions").
			Where("id", "=", values.Get("id")).Update(dialect.H{
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		})
		return err
	})

	return
}

func (s *SystemTable) GetRolesTable(ctx *context.Context) (roleTable Table) {
	roleTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := roleTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("role"), "name", db.Varchar).FieldFilterable()
	info.AddField(lg("slug"), "slug", db.Varchar).FieldFilterable()
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_roles").
		SetTitle(lg("Roles Manage")).
		SetDescription(lg("Roles Manage")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)

			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteRoleUserErr := s.connection().WithTx(tx).
					Table("goadmin_role_users").
					WhereIn("role_id", ids).
					Delete()

				if db.CheckError(deleteRoleUserErr, db.DELETE) {
					return deleteRoleUserErr, nil
				}

				deleteRoleMenuErr := s.connection().WithTx(tx).
					Table("goadmin_role_menu").
					WhereIn("role_id", ids).
					Delete()

				if db.CheckError(deleteRoleMenuErr, db.DELETE) {
					return deleteRoleMenuErr, nil
				}

				deleteRolePermissionErr := s.connection().WithTx(tx).
					Table("goadmin_role_permissions").
					WhereIn("role_id", ids).
					Delete()

				if db.CheckError(deleteRolePermissionErr, db.DELETE) {
					return deleteRolePermissionErr, nil
				}

				deleteRolesErr := s.connection().WithTx(tx).
					Table("goadmin_roles").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteRolesErr, db.DELETE) {
					return deleteRolesErr, nil
				}

				return nil, nil
			})

			return txErr
		})

	formList := roleTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("role"), "name", db.Varchar, form.Text).FieldMust()
	formList.AddField(lg("slug"), "slug", db.Varchar, form.Text).FieldHelpMsg(template.HTML(lg("should be unique"))).FieldMust()
	formList.AddField(lg("permission"), "permission_id", db.Varchar, form.SelectBox).
		FieldOptionsFromTable("goadmin_permissions", "name", "id").
		FieldDisplay(func(model types.FieldModel) interface{} {
			var permissions = make([]string, 0)

			if model.ID == "" {
				return permissions
			}
			perModel, _ := s.table("goadmin_role_permissions").
				Select("permission_id").
				Where("role_id", "=", model.ID).
				All()
			for _, v := range perModel {
				permissions = append(permissions, strconv.FormatInt(v["permission_id"].(int64), 10))
			}
			return permissions
		}).FieldHelpMsg(template.HTML(lg("no corresponding options?")) +
		link("/admin/info/permission/new", "Create here."))

	formList.AddField(lg("updatedAt"), "updated_at", db.Timestamp, form.Default).FieldNotAllowAdd()
	formList.AddField(lg("createdAt"), "created_at", db.Timestamp, form.Default).FieldNotAllowAdd()

	formList.SetTable("goadmin_roles").
		SetTitle(lg("Roles Manage")).
		SetDescription(lg("Roles Manage"))

	formList.SetUpdateFn(func(values form2.Values) error {

		if models.Role().SetConn(s.conn).IsSlugExist(values.Get("slug"), values.Get("id")) {
			return errors.New("slug exists")
		}

		role := models.RoleWithId(values.Get("id")).SetConn(s.conn)

		_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

			_, updateRoleErr := role.WithTx(tx).Update(values.Get("name"), values.Get("slug"))

			if db.CheckError(updateRoleErr, db.UPDATE) {
				return updateRoleErr, nil
			}

			delPermissionErr := role.WithTx(tx).DeletePermissions()

			if db.CheckError(delPermissionErr, db.DELETE) {
				return delPermissionErr, nil
			}

			for i := 0; i < len(values["permission_id[]"]); i++ {
				_, addPermissionErr := role.WithTx(tx).AddPermission(values["permission_id[]"][i])
				if db.CheckError(addPermissionErr, db.INSERT) {
					return addPermissionErr, nil
				}
			}

			return nil, nil
		})

		return txErr
	})

	formList.SetInsertFn(func(values form2.Values) error {

		if models.Role().SetConn(s.conn).IsSlugExist(values.Get("slug"), "") {
			return errors.New("slug exists")
		}

		_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {
			role, createRoleErr := models.Role().WithTx(tx).SetConn(s.conn).New(values.Get("name"), values.Get("slug"))

			if db.CheckError(createRoleErr, db.INSERT) {
				return createRoleErr, nil
			}

			for i := 0; i < len(values["permission_id[]"]); i++ {
				_, addPermissionErr := role.WithTx(tx).AddPermission(values["permission_id[]"][i])
				if db.CheckError(addPermissionErr, db.INSERT) {
					return addPermissionErr, nil
				}
			}

			return nil, nil
		})

		return txErr
	})

	return
}

func (s *SystemTable) GetOpTable(ctx *context.Context) (opTable Table) {
	opTable = NewDefaultTable(Config{
		Driver:     config.GetDatabases().GetDefault().Driver,
		CanAdd:     false,
		Editable:   false,
		Deletable:  false,
		Exportable: true,
		Connection: "default",
		PrimaryKey: PrimaryKey{
			Type: db.Int,
			Name: DefaultPrimaryKeyName,
		},
	})

	info := opTable.GetInfo().AddXssJsFilter().
		HideFilterArea().HideDeleteButton().HideDetailButton().HideEditButton().HideNewButton()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField("userID", "user_id", db.Int).FieldHide()
	info.AddField(lg("user"), "name", db.Varchar).FieldJoin(types.Join{
		Table:     config.GetAuthUserTable(),
		JoinField: "id",
		Field:     "user_id",
	}).FieldDisplay(func(value types.FieldModel) interface{} {
		return template.Default().
			Link().
			SetURL(config.Url("/info/manager/detail?__goadmin_detail_pk=") + strconv.Itoa(int(value.Row["user_id"].(int64)))).
			SetContent(template.HTML(value.Value)).
			OpenInNewTab().
			SetTabTitle("Manager Detail").
			GetContent()
	}).FieldFilterable()
	info.AddField(lg("path"), "path", db.Varchar).FieldFilterable()
	info.AddField(lg("method"), "method", db.Varchar).FieldFilterable()
	info.AddField(lg("ip"), "ip", db.Varchar).FieldFilterable()
	info.AddField(lg("content"), "input", db.Text).FieldWidth(230)
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)

	users, _ := s.table(config.GetAuthUserTable()).Select("id", "name").All()
	options := make(types.FieldOptions, len(users))
	for k, user := range users {
		options[k].Value = fmt.Sprintf("%v", user["id"])
		options[k].Text = fmt.Sprintf("%v", user["name"])
	}
	info.AddSelectBox(language.Get("user"), options, action.FieldFilter("user_id"))
	info.AddSelectBox(language.Get("method"), types.FieldOptions{
		{Value: "GET", Text: "GET"},
		{Value: "POST", Text: "POST"},
		{Value: "OPTIONS", Text: "OPTIONS"},
		{Value: "PUT", Text: "PUT"},
		{Value: "HEAD", Text: "HEAD"},
		{Value: "DELETE", Text: "DELETE"},
	}, action.FieldFilter("method"))

	info.SetTable("goadmin_operation_log").
		SetTitle(lg("operation log")).
		SetDescription(lg("operation log"))

	formList := opTable.GetForm().AddXssJsFilter()

	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("userID"), "user_id", db.Int, form.Text)
	formList.AddField(lg("path"), "path", db.Varchar, form.Text)
	formList.AddField(lg("method"), "method", db.Varchar, form.Text)
	formList.AddField(lg("ip"), "ip", db.Varchar, form.Text)
	formList.AddField(lg("content"), "input", db.Varchar, form.Text)
	formList.AddField(lg("updatedAt"), "updated_at", db.Timestamp, form.Default).FieldNotAllowAdd()
	formList.AddField(lg("createdAt"), "created_at", db.Timestamp, form.Default).FieldNotAllowAdd()

	formList.SetTable("goadmin_operation_log").
		SetTitle(lg("operation log")).
		SetDescription(lg("operation log"))

	return
}

func (s *SystemTable) GetMenuTable(ctx *context.Context) (menuTable Table) {
	menuTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver))

	info := menuTable.GetInfo().AddXssJsFilter().HideFilterArea()

	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField(lg("parent"), "parent_id", db.Int)
	info.AddField(lg("menu name"), "title", db.Varchar)
	info.AddField(lg("icon"), "icon", db.Varchar)
	info.AddField(lg("uri"), "uri", db.Varchar)
	info.AddField(lg("role"), "roles", db.Varchar)
	info.AddField(lg("header"), "header", db.Varchar)
	info.AddField(lg("createdAt"), "created_at", db.Timestamp)
	info.AddField(lg("updatedAt"), "updated_at", db.Timestamp)

	info.SetTable("goadmin_menu").
		SetTitle(lg("Menus Manage")).
		SetDescription(lg("Menus Manage")).
		SetDeleteFn(func(idArr []string) error {

			var ids = interfaces(idArr)

			_, txErr := s.connection().WithTransaction(func(tx *sql.Tx) (e error, i map[string]interface{}) {

				deleteRoleMenuErr := s.connection().WithTx(tx).
					Table("goadmin_role_menu").
					WhereIn("menu_id", ids).
					Delete()

				if db.CheckError(deleteRoleMenuErr, db.DELETE) {
					return deleteRoleMenuErr, nil
				}

				deleteMenusErr := s.connection().WithTx(tx).
					Table("goadmin_menu").
					WhereIn("id", ids).
					Delete()

				if db.CheckError(deleteMenusErr, db.DELETE) {
					return deleteMenusErr, nil
				}

				return nil, map[string]interface{}{}
			})

			return txErr
		})

	var parentIDOptions = types.FieldOptions{
		{
			Text:  "ROOT",
			Value: "0",
		},
	}

	allMenus, _ := s.connection().Table("goadmin_menu").
		Where("parent_id", "=", 0).
		Select("id", "title").
		OrderBy("order", "asc").
		All()
	allMenuIDs := make([]interface{}, len(allMenus))

	if len(allMenuIDs) > 0 {
		for i := 0; i < len(allMenus); i++ {
			allMenuIDs[i] = allMenus[i]["id"]
		}

		secondLevelMenus, _ := s.connection().Table("goadmin_menu").
			WhereIn("parent_id", allMenuIDs).
			Select("id", "title", "parent_id").
			All()

		secondLevelMenusCol := collection.Collection(secondLevelMenus)

		for i := 0; i < len(allMenus); i++ {
			parentIDOptions = append(parentIDOptions, types.FieldOption{
				TextHTML: "&nbsp;&nbsp;┝  " + language.GetFromHtml(template.HTML(allMenus[i]["title"].(string))),
				Value:    strconv.Itoa(int(allMenus[i]["id"].(int64))),
			})
			col := secondLevelMenusCol.Where("parent_id", "=", allMenus[i]["id"].(int64))
			for i := 0; i < len(col); i++ {
				parentIDOptions = append(parentIDOptions, types.FieldOption{
					TextHTML: "&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;┝  " +
						language.GetFromHtml(template.HTML(col[i]["title"].(string))),
					Value: strconv.Itoa(int(col[i]["id"].(int64))),
				})
			}
		}
	}

	formList := menuTable.GetForm().AddXssJsFilter()
	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField(lg("parent"), "parent_id", db.Int, form.SelectSingle).
		FieldOptions(parentIDOptions).
		FieldDisplay(func(model types.FieldModel) interface{} {
			var menuItem []string

			if model.ID == "" {
				return menuItem
			}

			menuModel, _ := s.table("goadmin_menu").Select("parent_id").Find(model.ID)
			menuItem = append(menuItem, strconv.FormatInt(menuModel["parent_id"].(int64), 10))
			return menuItem
		})
	formList.AddField(lg("menu name"), "title", db.Varchar, form.Text).FieldMust()
	formList.AddField(lg("header"), "header", db.Varchar, form.Text)
	formList.AddField(lg("icon"), "icon", db.Varchar, form.IconPicker)
	formList.AddField(lg("uri"), "uri", db.Varchar, form.Text)
	formList.AddField(lg("role"), "roles", db.Int, form.Select).
		FieldOptionsFromTable("goadmin_roles", "slug", "id").
		FieldDisplay(func(model types.FieldModel) interface{} {
			var roles []string

			if model.ID == "" {
				return roles
			}

			roleModel, _ := s.table("goadmin_role_menu").
				Select("role_id").
				Where("menu_id", "=", model.ID).
				All()

			for _, v := range roleModel {
				roles = append(roles, strconv.FormatInt(v["role_id"].(int64), 10))
			}
			return roles
		})

	formList.AddField(lg("updatedAt"), "updated_at", db.Timestamp, form.Default).FieldNotAllowAdd()
	formList.AddField(lg("createdAt"), "created_at", db.Timestamp, form.Default).FieldNotAllowAdd()

	formList.SetTable("goadmin_menu").
		SetTitle(lg("Menus Manage")).
		SetDescription(lg("Menus Manage"))

	return
}

func (s *SystemTable) GetSiteTable(ctx *context.Context) (siteTable Table) {
	siteTable = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver).
		SetOnlyUpdateForm().
		SetGetDataFun(func(params parameter.Parameters) (i []map[string]interface{}, i2 int) {
			return []map[string]interface{}{models.Site().SetConn(s.conn).AllToMapInterface()}, 1
		}))

	trueStr := lgWithConfigScore("true")
	falseStr := lgWithConfigScore("false")

	formList := siteTable.GetForm().AddXssJsFilter()
	formList.AddField("ID", "id", db.Varchar, form.Default).FieldDefault("1").FieldHide()
	formList.AddField(lgWithConfigScore("site off"), "site_off", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("debug"), "debug", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("env"), "env", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: lgWithConfigScore("test"), Value: config.EnvTest},
			{Text: lgWithConfigScore("prod"), Value: config.EnvProd},
			{Text: lgWithConfigScore("local"), Value: config.EnvLocal},
		})

	langOps := make(types.FieldOptions, len(language.Langs))
	for k, t := range language.Langs {
		langOps[k] = types.FieldOption{Text: lgWithConfigScore(t, "language"), Value: t}
	}
	formList.AddField(lgWithConfigScore("language"), "language", db.Varchar, form.SelectSingle).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return language.FixedLanguageKey(value.Value)
		}).
		FieldOptions(langOps)
	themes := template.Themes()
	themesOps := make(types.FieldOptions, len(themes))
	for k, t := range themes {
		themesOps[k] = types.FieldOption{Text: t, Value: t}
	}

	formList.AddField(lgWithConfigScore("theme"), "theme", db.Varchar, form.SelectSingle).
		FieldOptions(themesOps).
		FieldOnChooseShow("adminlte",
			"color_scheme")
	formList.AddField(lgWithConfigScore("title"), "title", db.Varchar, form.Text).FieldMust()
	formList.AddField(lgWithConfigScore("color scheme"), "color_scheme", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "skin-black", Value: "skin-black"},
			{Text: "skin-black-light", Value: "skin-black-light"},
			{Text: "skin-blue", Value: "skin-blue"},
			{Text: "skin-blue-light", Value: "skin-blue-light"},
			{Text: "skin-green", Value: "skin-green"},
			{Text: "skin-green-light", Value: "skin-green-light"},
			{Text: "skin-purple", Value: "skin-purple"},
			{Text: "skin-purple-light", Value: "skin-purple-light"},
			{Text: "skin-red", Value: "skin-red"},
			{Text: "skin-red-light", Value: "skin-red-light"},
			{Text: "skin-yellow", Value: "skin-yellow"},
			{Text: "skin-yellow-light", Value: "skin-yellow-light"},
		}).FieldHelpMsg(template.HTML(lgWithConfigScore("It will work when theme is adminlte")))
	formList.AddField(lgWithConfigScore("login title"), "login_title", db.Varchar, form.Text).FieldMust()
	formList.AddField(lgWithConfigScore("extra"), "extra", db.Varchar, form.TextArea)
	//formList.AddField(lgWithConfigScore("databases"), "databases", db.Varchar, form.TextArea).
	//	FieldDisplay(func(value types.FieldModel) interface{} {
	//		var buf = new(bytes.Buffer)
	//		_ = json.Indent(buf, []byte(value.Value), "", "    ")
	//		return template.HTML(buf.String())
	//	}).FieldNotAllowEdit()

	formList.AddField(lgWithConfigScore("logo"), "logo", db.Varchar, form.Code).FieldMust()
	formList.AddField(lgWithConfigScore("mini logo"), "mini_logo", db.Varchar, form.Code).FieldMust()
	formList.AddField(lgWithConfigScore("session life time"), "session_life_time", db.Varchar, form.Number).
		FieldMust().
		FieldHelpMsg(template.HTML(lgWithConfigScore("must bigger than 900 seconds")))
	formList.AddField(lgWithConfigScore("custom head html"), "custom_head_html", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("custom foot Html"), "custom_foot_html", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("custom 404 html"), "custom_404_html", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("custom 403 html"), "custom_403_html", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("custom 500 Html"), "custom_500_html", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("footer info"), "footer_info", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("login logo"), "login_logo", db.Varchar, form.Code)
	formList.AddField(lgWithConfigScore("no limit login ip"), "no_limit_login_ip", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("hide config center entrance"), "hide_config_center_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("hide app info entrance"), "hide_app_info_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("hide tool entrance"), "hide_tool_entrance", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("animation type"), "animation_type", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "", Value: ""},
			{Text: "bounce", Value: "bounce"}, {Text: "flash", Value: "flash"}, {Text: "pulse", Value: "pulse"},
			{Text: "rubberBand", Value: "rubberBand"}, {Text: "shake", Value: "shake"}, {Text: "swing", Value: "swing"},
			{Text: "tada", Value: "tada"}, {Text: "wobble", Value: "wobble"}, {Text: "jello", Value: "jello"},
			{Text: "heartBeat", Value: "heartBeat"}, {Text: "bounceIn", Value: "bounceIn"}, {Text: "bounceInDown", Value: "bounceInDown"},
			{Text: "bounceInLeft", Value: "bounceInLeft"}, {Text: "bounceInRight", Value: "bounceInRight"}, {Text: "bounceInUp", Value: "bounceInUp"},
			{Text: "fadeIn", Value: "fadeIn"}, {Text: "fadeInDown", Value: "fadeInDown"}, {Text: "fadeInDownBig", Value: "fadeInDownBig"},
			{Text: "fadeInLeft", Value: "fadeInLeft"}, {Text: "fadeInLeftBig", Value: "fadeInLeftBig"}, {Text: "fadeInRight", Value: "fadeInRight"},
			{Text: "fadeInRightBig", Value: "fadeInRightBig"}, {Text: "fadeInUp", Value: "fadeInUp"}, {Text: "fadeInUpBig", Value: "fadeInUpBig"},
			{Text: "flip", Value: "flip"}, {Text: "flipInX", Value: "flipInX"}, {Text: "flipInY", Value: "flipInY"},
			{Text: "lightSpeedIn", Value: "lightSpeedIn"}, {Text: "rotateIn", Value: "rotateIn"}, {Text: "rotateInDownLeft", Value: "rotateInDownLeft"},
			{Text: "rotateInDownRight", Value: "rotateInDownRight"}, {Text: "rotateInUpLeft", Value: "rotateInUpLeft"}, {Text: "rotateInUpRight", Value: "rotateInUpRight"},
			{Text: "slideInUp", Value: "slideInUp"}, {Text: "slideInDown", Value: "slideInDown"}, {Text: "slideInLeft", Value: "slideInLeft"}, {Text: "slideInRight", Value: "slideInRight"},
			{Text: "slideOutRight", Value: "slideOutRight"}, {Text: "zoomIn", Value: "zoomIn"}, {Text: "zoomInDown", Value: "zoomInDown"},
			{Text: "zoomInLeft", Value: "zoomInLeft"}, {Text: "zoomInRight", Value: "zoomInRight"}, {Text: "zoomInUp", Value: "zoomInUp"},
			{Text: "hinge", Value: "hinge"}, {Text: "jackInTheBox", Value: "jackInTheBox"}, {Text: "rollIn", Value: "rollIn"},
		}).FieldOnChooseHide("", "animation_duration", "animation_delay").
		FieldOptionExt(map[string]interface{}{"allowClear": true}).
		FieldHelpMsg(`see more: <a href="https://daneden.github.io/animate.css/">https://daneden.github.io/animate.css/</a>`)

	formList.AddField(lgWithConfigScore("animation duration"), "animation_duration", db.Varchar, form.Number)
	formList.AddField(lgWithConfigScore("animation delay"), "animation_delay", db.Varchar, form.Number)

	formList.AddField(lgWithConfigScore("file upload engine"), "file_upload_engine", db.Varchar, form.Text)

	formList.AddField(lgWithConfigScore("cdn url"), "asset_url", db.Varchar, form.Text).
		FieldHelpMsg(template.HTML(lgWithConfigScore("Do not modify when you have not set up all assets")))

	formList.AddField(lgWithConfigScore("info log path"), "info_log_path", db.Varchar, form.Text)
	formList.AddField(lgWithConfigScore("error log path"), "error_log_path", db.Varchar, form.Text)
	formList.AddField(lgWithConfigScore("access log path"), "access_log_path", db.Varchar, form.Text)
	formList.AddField(lgWithConfigScore("info log off"), "info_log_off", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("error log off"), "error_log_off", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("access log off"), "access_log_off", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("access assets log off"), "access_assets_log_off", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("sql log on"), "sql_log", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		})
	formList.AddField(lgWithConfigScore("log level"), "logger_level", db.Varchar, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Debug", Value: "-1"},
			{Text: "Info", Value: "0"},
			{Text: "Warn", Value: "1"},
			{Text: "Error", Value: "2"},
		}).FieldDisplay(defaultFilterFn("0"))

	formList.AddField(lgWithConfigScore("logger rotate max size"), "logger_rotate_max_size", db.Varchar, form.Number).
		FieldDivider(lgWithConfigScore("logger rotate")).FieldDisplay(defaultFilterFn("10", "0"))
	formList.AddField(lgWithConfigScore("logger rotate max backups"), "logger_rotate_max_backups", db.Varchar, form.Number).
		FieldDisplay(defaultFilterFn("5", "0"))
	formList.AddField(lgWithConfigScore("logger rotate max age"), "logger_rotate_max_age", db.Varchar, form.Number).
		FieldDisplay(defaultFilterFn("30", "0"))
	formList.AddField(lgWithConfigScore("logger rotate compress"), "logger_rotate_compress", db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: trueStr, Value: "true"},
			{Text: falseStr, Value: "false"},
		}).FieldDisplay(defaultFilterFn("false"))

	formList.AddField(lgWithConfigScore("logger rotate encoder encoding"), "logger_encoder_encoding", db.Varchar,
		form.SelectSingle).
		FieldDivider(lgWithConfigScore("logger rotate encoder")).
		FieldOptions(types.FieldOptions{
			{Text: "JSON", Value: "json"},
			{Text: "Console", Value: "console"},
		}).FieldDisplay(defaultFilterFn("console")).
		FieldOnChooseHide("Console",
			"logger_encoder_time_key", "logger_encoder_level_key", "logger_encoder_caller_key",
			"logger_encoder_message_key", "logger_encoder_stacktrace_key", "logger_encoder_name_key")

	formList.AddField(lgWithConfigScore("logger rotate encoder time key"), "logger_encoder_time_key", db.Varchar, form.Text).
		FieldDisplay(defaultFilterFn("ts"))
	formList.AddField(lgWithConfigScore("logger rotate encoder level key"), "logger_encoder_level_key", db.Varchar, form.Text).
		FieldDisplay(defaultFilterFn("level"))
	formList.AddField(lgWithConfigScore("logger rotate encoder name key"), "logger_encoder_name_key", db.Varchar, form.Text).
		FieldDisplay(defaultFilterFn("logger"))
	formList.AddField(lgWithConfigScore("logger rotate encoder caller key"), "logger_encoder_caller_key", db.Varchar, form.Text).
		FieldDisplay(defaultFilterFn("caller"))
	formList.AddField(lgWithConfigScore("logger rotate encoder message key"), "logger_encoder_message_key", db.Varchar, form.Text).
		FieldDisplay(defaultFilterFn("msg"))
	formList.AddField(lgWithConfigScore("logger rotate encoder stacktrace key"), "logger_encoder_stacktrace_key", db.Varchar, form.Text).
		FieldDisplay(defaultFilterFn("stacktrace"))

	formList.AddField(lgWithConfigScore("logger rotate encoder level"), "logger_encoder_level", db.Varchar,
		form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: lgWithConfigScore("capital"), Value: "capital"},
			{Text: lgWithConfigScore("capitalcolor"), Value: "capitalColor"},
			{Text: lgWithConfigScore("lowercase"), Value: "lowercase"},
			{Text: lgWithConfigScore("lowercasecolor"), Value: "color"},
		}).FieldDisplay(defaultFilterFn("capitalColor"))
	formList.AddField(lgWithConfigScore("logger rotate encoder time"), "logger_encoder_time", db.Varchar,
		form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "ISO8601(2006-01-02T15:04:05.000Z0700)", Value: "iso8601"},
			{Text: lgWithConfigScore("millisecond"), Value: "millis"},
			{Text: lgWithConfigScore("nanosecond"), Value: "nanos"},
			{Text: "RFC3339(2006-01-02T15:04:05Z07:00)", Value: "rfc3339"},
			{Text: "RFC3339 Nano(2006-01-02T15:04:05.999999999Z07:00)", Value: "rfc3339nano"},
		}).FieldDisplay(defaultFilterFn("iso8601"))
	formList.AddField(lgWithConfigScore("logger rotate encoder duration"), "logger_encoder_duration", db.Varchar,
		form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: lgWithConfigScore("seconds"), Value: "string"},
			{Text: lgWithConfigScore("nanosecond"), Value: "nanos"},
			{Text: lgWithConfigScore("microsecond"), Value: "ms"},
		}).FieldDisplay(defaultFilterFn("string"))
	formList.AddField(lgWithConfigScore("logger rotate encoder caller"), "logger_encoder_caller", db.Varchar,
		form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: lgWithConfigScore("full path"), Value: "full"},
			{Text: lgWithConfigScore("short path"), Value: "short"},
		}).FieldDisplay(defaultFilterFn("full"))

	formList.HideBackButton().HideContinueEditCheckBox().HideContinueNewCheckBox()
	formList.SetTabGroups(types.NewTabGroups("id", "debug", "env", "language", "theme", "color_scheme",
		"asset_url", "title", "login_title", "session_life_time", "no_limit_login_ip",
		"hide_config_center_entrance", "hide_app_info_entrance", "hide_tool_entrance",
		"animation_type",
		"animation_duration", "animation_delay", "file_upload_engine", "extra").
		AddGroup("access_log_off", "access_assets_log_off", "info_log_off", "error_log_off", "sql_log", "logger_level",
			"info_log_path", "error_log_path",
			"access_log_path", "logger_rotate_max_size", "logger_rotate_max_backups",
			"logger_rotate_max_age", "logger_rotate_compress",
			"logger_encoder_encoding", "logger_encoder_time_key", "logger_encoder_level_key", "logger_encoder_name_key",
			"logger_encoder_caller_key", "logger_encoder_message_key", "logger_encoder_stacktrace_key", "logger_encoder_level",
			"logger_encoder_time", "logger_encoder_duration", "logger_encoder_caller").
		AddGroup("logo", "mini_logo", "custom_head_html", "custom_foot_html", "footer_info", "login_logo",
			"custom_404_html", "custom_403_html", "custom_500_html")).
		SetTabHeaders(lgWithConfigScore("general"), lgWithConfigScore("log"), lgWithConfigScore("custom"))

	formList.SetTable("goadmin_site").
		SetTitle(lgWithConfigScore("site setting")).
		SetDescription(lgWithConfigScore("site setting"))

	formList.SetUpdateFn(func(values form2.Values) error {

		ses := values.Get("session_life_time")
		sesInt, _ := strconv.Atoi(ses)
		if sesInt < 900 {
			return errors.New("wrong session life time, must bigger than 900 seconds")
		}
		if err := checkJSON(values, "file_upload_engine"); err != nil {
			return err
		}

		values["logo"][0] = escape(values.Get("logo"))
		values["mini_logo"][0] = escape(values.Get("mini_logo"))
		values["custom_head_html"][0] = escape(values.Get("custom_head_html"))
		values["custom_foot_html"][0] = escape(values.Get("custom_foot_html"))
		values["custom_404_html"][0] = escape(values.Get("custom_404_html"))
		values["custom_403_html"][0] = escape(values.Get("custom_403_html"))
		values["custom_500_html"][0] = escape(values.Get("custom_500_html"))
		values["footer_info"][0] = escape(values.Get("footer_info"))
		values["login_logo"][0] = escape(values.Get("login_logo"))

		var err error
		if s.c.UpdateProcessFn != nil {
			values, err = s.c.UpdateProcessFn(values)
			if err != nil {
				return err
			}
		}

		ui.GetService(services).RemoveOrShowSiteNavButton(values["hide_config_center_entrance"][0] == "true")
		ui.GetService(services).RemoveOrShowInfoNavButton(values["hide_app_info_entrance"][0] == "true")
		ui.GetService(services).RemoveOrShowToolNavButton(values["hide_tool_entrance"][0] == "true")

		// TODO: add transaction
		err = models.Site().SetConn(s.conn).Update(values.RemoveSysRemark())
		if err != nil {
			return err
		}
		return s.c.Update(values.ToMap())
	})

	formList.EnableAjax(lg("success"), lg("fail"))

	return
}

func (s *SystemTable) GetGenerateForm(ctx *context.Context) (generateTool Table) {
	generateTool = NewDefaultTable(DefaultConfigWithDriver(config.GetDatabases().GetDefault().Driver).
		SetOnlyNewForm())

	formList := generateTool.GetForm().AddXssJsFilter().
		SetHeadWidth(1).
		SetInputWidth(4).
		HideBackButton().
		HideContinueNewCheckBox().
		HideResetButton()

	formList.AddField("ID", "id", db.Varchar, form.Default).FieldDefault("1").FieldHide()

	connNames := config.GetDatabases().Connections()
	ops := make(types.FieldOptions, len(connNames))
	for i, name := range connNames {
		ops[i] = types.FieldOption{Text: name, Value: name}
	}

	formList.AddField(lgWithScore("connection", "tool"), "conn", db.Varchar, form.SelectSingle).
		FieldOptions(ops).
		FieldOnChooseAjax("table", "/tool/choose/conn",
			func(ctx *context.Context) (success bool, msg string, data interface{}) {
				connName := ctx.FormValue("value")
				if connName == "" {
					return false, "wrong parameter", nil
				}
				cfg := s.c.Databases[connName]
				conn := db.GetConnectionFromService(services.Get(cfg.Driver))
				tables, err := db.WithDriverAndConnection(connName, conn).Table(cfg.Name).ShowTables()
				if err != nil {
					return false, err.Error(), nil
				}
				ops := make(selection.Options, len(tables))
				for i, table := range tables {
					ops[i] = selection.Option{Text: table, ID: table}
				}
				return true, "ok", ops
			})
	formList.AddField(lgWithScore("table", "tool"), "table", db.Varchar, form.SelectSingle).
		FieldOnChooseAjax("xxxx", "/tool/choose/table",
			func(ctx *context.Context) (success bool, msg string, data interface{}) {

				var (
					tableName       = ctx.FormValue("value")
					connName        = ctx.FormValue("conn")
					driver          = s.c.Databases[connName].Driver
					conn            = db.GetConnectionFromService(services.Get(driver))
					columnsModel, _ = db.WithDriver(conn).Table(tableName).ShowColumns()

					fieldField = "Field"
					typeField  = "Type"
				)

				if driver == "postgresql" {
					fieldField = "column_name"
					typeField = "udt_name"
				}
				if driver == "sqlite" {
					fieldField = "name"
					typeField = "type"
				}
				if driver == "mssql" {
					fieldField = "column_name"
					typeField = "data_type"
				}

				headName := make([]string, len(columnsModel))
				fieldName := make([]string, len(columnsModel))
				dbTypeList := make([]string, len(columnsModel))
				formTypeList := make([]string, len(columnsModel))

				for i, model := range columnsModel {
					typeName := getType(model[typeField].(string))

					headName[i] = strings.Title(model[fieldField].(string))
					fieldName[i] = model[fieldField].(string)
					dbTypeList[i] = typeName
					formTypeList[i] = form.GetFormTypeFromFieldType(db.DT(strings.ToUpper(typeName)),
						model[fieldField].(string))
				}

				return true, "ok", [][]string{headName, fieldName, dbTypeList, formTypeList}
			}, `
				$("tbody.fields-table").find("tr").remove();
				let tpl = $("template.fields-tpl").html();
				for (let i = 0; i < data.data[0].length; i++) {
					$("tbody.fields-table").append(tpl);
				}
				let trs = $("tbody.fields-table").find("tr");
				for (let i = 0; i < data.data[0].length; i++) {
					$(trs[i]).find('.field_head').val(data.data[0][i]);
					$(trs[i]).find('.field_name').val(data.data[1][i]);
					$(trs[i]).find('select.field_db_type').val(data.data[2][i]).select2();
				}
				$("tbody.fields_form-table").find("tr").remove();
				let tpl_form = $("template.fields_form-tpl").html();
				for (let i = 0; i < data.data[0].length; i++) {
					$("tbody.fields_form-table").append(tpl_form);
				}
				let trs_form = $("tbody.fields_form-table").find("tr");
				for (let i = 0; i < data.data[0].length; i++) {
					$(trs_form[i]).find('.field_head_form').val(data.data[0][i]);
					$(trs_form[i]).find('.field_name_form').val(data.data[1][i]);
					$(trs_form[i]).find('select.field_db_type_form').val(data.data[2][i]).select2();
					$(trs_form[i]).find('select.field_form_type_form').val(data.data[3][i]).select2();
				}
				`, `"conn":$('.conn').val(),`)
	formList.AddField(lgWithScore("package", "tool"), "package", db.Varchar, form.Text).FieldDefault("tables")
	formList.AddField(lgWithScore("primarykey", "tool"), "pk", db.Varchar, form.Text).FieldDefault("id")

	formList.AddRow(func(panel *types.FormPanel) {
		addYesNoSwitchForTool(panel, "filter area", "hide_filter_area", "n", 2)
		panel.AddField(lgWithScore("filter form layout", "tool"), "filter_form_layout", db.Varchar, form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: form.LayoutDefault.String(), Value: form.LayoutDefault.String()},
				{Text: form.LayoutTwoCol.String(), Value: form.LayoutTwoCol.String()},
				{Text: form.LayoutThreeCol.String(), Value: form.LayoutThreeCol.String()},
				{Text: form.LayoutFourCol.String(), Value: form.LayoutFourCol.String()},
				{Text: form.LayoutFlow.String(), Value: form.LayoutFlow.String()},
			}).FieldDefault(form.LayoutDefault.String()).
			FieldRowWidth(4).FieldHeadWidth(3)
	})

	formList.AddRow(func(panel *types.FormPanel) {
		addYesNoSwitchForTool(panel, "new button", "hide_new_button", "n", 2)
		addYesNoSwitchForTool(panel, "export button", "hide_export_button", "n", 4, 3)
		addYesNoSwitchForTool(panel, "edit button", "hide_edit_button", "n", 4, 2)
	})

	formList.AddRow(func(panel *types.FormPanel) {
		addYesNoSwitchForTool(panel, "pagination", "hide_pagination", "n", 2)
		addYesNoSwitchForTool(panel, "delete button", "hide_delete_button", "n", 4, 3)
		addYesNoSwitchForTool(panel, "detail button", "hide_detail_button", "n", 4, 2)
	})

	formList.AddRow(func(panel *types.FormPanel) {
		addYesNoSwitchForTool(panel, "filter button", "hide_filter_button", "n", 2)
		addYesNoSwitchForTool(panel, "row selector", "hide_row_selector", "n", 4, 3)
		addYesNoSwitchForTool(panel, "query info", "hide_query_info", "n", 4, 2)
	})

	formList.AddField(lgWithScore("output", "tool"), "path", db.Varchar, form.Text).
		FieldDefault("").FieldMust().FieldHelpMsg(template.HTML(lgWithScore("use absolute path", "tool")))
	formList.AddTable(lgWithScore("field", "tool"), "fields", func(pa *types.FormPanel) {
		pa.AddField(lgWithScore("title", "tool"), "field_head", db.Varchar, form.Text).FieldHideLabel().
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{""}
			})
		pa.AddField(lgWithScore("field name", "tool"), "field_name", db.Varchar, form.Text).FieldHideLabel().
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{""}
			})
		pa.AddField(lgWithScore("field filterable", "tool"), "field_filterable", db.Varchar, form.CheckboxSingle).
			FieldOptions(types.FieldOptions{
				{Text: "", Value: "y"},
				{Text: "", Value: "n"},
			}).
			FieldDefault("n").
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{"n"}
			})
		pa.AddField(lgWithScore("field sortable", "tool"), "field_sortable", db.Varchar, form.CheckboxSingle).
			FieldOptions(types.FieldOptions{
				{Text: "", Value: "y"},
				{Text: "", Value: "n"},
			}).
			FieldDefault("n").
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{"n"}
			})
		pa.AddField(lgWithScore("db type", "tool"), "field_db_type", db.Varchar, form.SelectSingle).
			FieldOptions(databaseTypeOptions()).FieldDisplay(func(value types.FieldModel) interface{} {
			return []string{""}
		})
	}).FieldInputWidth(11)

	formList.AddRow(func(panel *types.FormPanel) {
		addYesNoSwitchForTool(panel, "continue edit checkbox", "hide_continue_edit_check_box", "n", 2)
		addYesNoSwitchForTool(panel, "reset button", "hide_reset_button", "n", 5, 3)
	})

	formList.AddRow(func(panel *types.FormPanel) {
		addYesNoSwitchForTool(panel, "continue new checkbox", "hide_continue_new_check_box", "n", 2)
		addYesNoSwitchForTool(panel, "back button", "hide_back_button", "n", 5, 3)
	})

	formList.AddTable(lgWithScore("field", "tool"), "fields_form", func(pa *types.FormPanel) {
		pa.AddField(lgWithScore("title", "tool"), "field_head_form", db.Varchar, form.Text).FieldHideLabel().
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{""}
			})
		pa.AddField(lgWithScore("field name", "tool"), "field_name_form", db.Varchar, form.Text).FieldHideLabel().
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{""}
			})
		pa.AddField(lgWithScore("field editable", "tool"), "field_canedit", db.Varchar, form.CheckboxSingle).
			FieldOptions(types.FieldOptions{
				{Text: "", Value: "y"},
				{Text: "", Value: "n"},
			}).
			FieldDefault("y").
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{"y"}
			})
		pa.AddField(lgWithScore("field can add", "tool"), "field_canadd", db.Varchar, form.CheckboxSingle).
			FieldOptions(types.FieldOptions{
				{Text: "", Value: "y"},
				{Text: "", Value: "n"},
			}).
			FieldDefault("y").
			FieldDisplay(func(value types.FieldModel) interface{} {
				return []string{"y"}
			})
		pa.AddField(lgWithScore("db type", "tool"), "field_db_type_form", db.Varchar, form.SelectSingle).
			FieldOptions(databaseTypeOptions()).FieldDisplay(func(value types.FieldModel) interface{} {
			return []string{""}
		})
		pa.AddField(lgWithScore("form type", "tool"), "field_form_type_form", db.Varchar, form.SelectSingle).
			FieldOptions(formTypeOptions()).FieldDisplay(func(value types.FieldModel) interface{} {
			return []string{""}
		})
	}).FieldInputWidth(11)

	formList.SetTabGroups(types.
		NewTabGroups("conn", "table", "package", "pk", "path").
		AddGroup("hide_filter_area", "filter_form_layout",
			"hide_new_button", "hide_export_button", "hide_edit_button",
			"hide_pagination", "hide_delete_button", "hide_detail_button",
			"hide_filter_button", "hide_row_selector", "hide_query_info",
			"fields").
		AddGroup("hide_continue_edit_check_box", "hide_reset_button",
			"hide_continue_new_check_box", "hide_back_button",
			"fields_form")).
		SetTabHeaders(lgWithScore("basic info", "tool"), lgWithScore("table info", "tool"),
			lgWithScore("form info", "tool"))

	formList.SetTable("tool").
		SetTitle(lgWithScore("tool", "tool")).
		SetDescription(lgWithScore("tool", "tool")).
		SetHeader(template.HTML(`<h3 class="box-title">` +
			lgWithScore("generate table model", "tool") + `</h3>`))

	formList.SetInsertFn(func(values form2.Values) error {

		output := values.Get("path")

		if output == "" {
			return errors.New("output path is empty")
		}

		connName := values.Get("conn")

		fields := make(tools.Fields, len(values["field_head"]))

		for i := 0; i < len(values["field_head"]); i++ {
			fields[i] = tools.Field{
				Head:       values["field_head"][i],
				Name:       values["field_name"][i],
				DBType:     values["field_db_type"][i],
				Filterable: values["field_filterable"][i] == "y",
				Sortable:   values["field_sortable"][i] == "y",
			}
		}

		formFields := make(tools.Fields, len(values["field_head_form"]))

		for i := 0; i < len(values["field_head_form"]); i++ {
			formFields[i] = tools.Field{
				Head:     values["field_head_form"][i],
				Name:     values["field_name_form"][i],
				FormType: values["field_form_type_form"][i],
				DBType:   values["field_db_type_form"][i],
				CanAdd:   values["field_canadd"][i] == "y",
				Editable: values["field_canedit"][i] == "y",
			}
		}

		err := tools.Generate(tools.NewParamWithFields(tools.Config{
			Connection:               connName,
			Driver:                   s.c.Databases[connName].Driver,
			Package:                  values.Get("package"),
			Table:                    values.Get("table"),
			HideFilterArea:           values.Get("hide_filter_area") == "y",
			HideNewButton:            values.Get("hide_new_button") == "y",
			HideExportButton:         values.Get("hide_export_button") == "y",
			HideEditButton:           values.Get("hide_edit_button") == "y",
			HideDeleteButton:         values.Get("hide_delete_button") == "y",
			HideDetailButton:         values.Get("hide_detail_button") == "y",
			HideFilterButton:         values.Get("hide_filter_button") == "y",
			HideRowSelector:          values.Get("hide_row_selector") == "y",
			HidePagination:           values.Get("hide_pagination") == "y",
			HideQueryInfo:            values.Get("hide_query_info") == "y",
			HideContinueEditCheckBox: values.Get("hide_continue_edit_check_box") == "y",
			HideContinueNewCheckBox:  values.Get("hide_continue_new_check_box") == "y",
			HideResetButton:          values.Get("hide_reset_button") == "y",
			HideBackButton:           values.Get("hide_back_button") == "y",
			FilterFormLayout:         form.GetLayoutFromString(values.Get("filter_form_layout")),
			Schema:                   values.Get("schema"),
			Output:                   output,
		}, fields, formFields))

		if err != nil {
			return err
		}

		if utils.FileExist(output + "/tables.go") {
			return tools.GenerateTables(output, []string{values.Get("table")},
				values.Get("package"))
		}

		return nil
	})

	formList.EnableAjax(lg("success"), lg("fail"), s.c.Url("/info/generate/new"))

	return generateTool
}

// -------------------------
// helper functions
// -------------------------

func encodePassword(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash[:])
}

func label() types.LabelAttribute {
	return template.Get(config.GetTheme()).Label().SetType("success")
}

func lg(v string) string {
	return language.Get(v)
}

func defaultFilterFn(val string, def ...string) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		if len(def) > 0 {
			if value.Value == def[0] {
				return val
			}
		} else {
			if value.Value == "" {
				return val
			}
		}
		return value.Value
	}
}

func lgWithScore(v string, score ...string) string {
	return language.GetWithScope(v, score...)
}

func lgWithConfigScore(v string, score ...string) string {
	scores := append([]string{"config"}, score...)
	return language.GetWithScope(v, scores...)
}

func link(url, content string) tmpl.HTML {
	return html.AEl().
		SetAttr("href", url).
		SetContent(template.HTML(lg(content))).
		Get()
}

func escape(s string) string {
	if s == "" {
		return ""
	}
	s, err := url.QueryUnescape(s)
	if err != nil {
		logger.Error("config set error", err)
	}
	return s
}

func checkJSON(values form2.Values, key string) error {
	v := values.Get(key)
	if v != "" && !utils.IsJSON(v) {
		return errors.New("wrong " + key)
	}
	return nil
}

func (s *SystemTable) table(table string) *db.SQL {
	return s.connection().Table(table)
}

func (s *SystemTable) connection() *db.SQL {
	return db.WithDriver(s.conn)
}

func interfaces(arr []string) []interface{} {
	var iarr = make([]interface{}, len(arr))

	for key, v := range arr {
		iarr[key] = v
	}

	return iarr
}

func addYesNoSwitchForTool(formList *types.FormPanel, head, field, def string, row ...int) {
	formList.AddField(lgWithScore(head, "tool"), field, db.Varchar, form.Switch).
		FieldOptions(types.FieldOptions{
			{Text: lgWithScore("show", "tool"), Value: "n"},
			{Text: lgWithScore("hide", "tool"), Value: "y"},
		}).FieldDefault(def)
	if len(row) > 0 {
		formList.FieldRowWidth(row[0])
	}
	if len(row) > 1 {
		formList.FieldHeadWidth(row[1])
	}
	if len(row) > 2 {
		formList.FieldInputWidth(row[2])
	}
}

func formTypeOptions() types.FieldOptions {
	return types.FieldOptions{
		{Text: "Default", Value: "Default"},
		{Text: "Text", Value: "Text"},
		{Text: "SelectSingle", Value: "SelectSingle"},
		{Text: "Select", Value: "Select"},
		{Text: "IconPicker", Value: "IconPicker"},
		{Text: "SelectBox", Value: "SelectBox"},
		{Text: "File", Value: "File"},
		{Text: "Multifile", Value: "Multifile"},
		{Text: "Password", Value: "Password"},
		{Text: "RichText", Value: "RichText"},
		{Text: "Datetime", Value: "Datetime"},
		{Text: "DatetimeRange", Value: "DatetimeRange"},
		{Text: "Radio", Value: "Radio"},
		{Text: "Email", Value: "Email"},
		{Text: "Url", Value: "Url"},
		{Text: "Ip", Value: "Ip"},
		{Text: "Color", Value: "Color"},
		{Text: "Array", Value: "Array"},
		{Text: "Currency", Value: "Currency"},
		{Text: "Number", Value: "Number"},
		{Text: "Table", Value: "Table"},
		{Text: "NumberRange", Value: "NumberRange"},
		{Text: "TextArea", Value: "TextArea"},
		{Text: "Custom", Value: "Custom"},
		{Text: "Switch", Value: "Switch"},
		{Text: "Code", Value: "Code"},
	}
}

func databaseTypeOptions() types.FieldOptions {
	return types.FieldOptions{
		{Text: "INT", Value: "Int"},
		{Text: "TINYINT", Value: "Tinyint"},
		{Text: "MEDIUMINT", Value: "Mediumint"},
		{Text: "SMALLINT", Value: "Smallint"},
		{Text: "BIGINT", Value: "Bigint"},
		{Text: "BIT", Value: "Bit"},
		{Text: "INT8", Value: "Int8"},
		{Text: "INT4", Value: "Int4"},
		{Text: "INT2", Value: "Int2"},
		{Text: "INTEGER", Value: "Integer"},
		{Text: "NUMERIC", Value: "Numeric"},
		{Text: "SMALLSERIAL", Value: "Smallserial"},
		{Text: "SERIAL", Value: "Serial"},
		{Text: "BIGSERIAL", Value: "Bigserial"},
		{Text: "MONEY", Value: "Money"},
		{Text: "REAL", Value: "Real"},
		{Text: "FLOAT", Value: "Float"},
		{Text: "FLOAT4", Value: "Float4"},
		{Text: "FLOAT8", Value: "Float8"},
		{Text: "DOUBLE", Value: "Double"},
		{Text: "DECIMAL", Value: "Decimal"},
		{Text: "DOUBLEPRECISION", Value: "Doubleprecision"},
		{Text: "DATE", Value: "Date"},
		{Text: "TIME", Value: "Time"},
		{Text: "YEAR", Value: "Year"},
		{Text: "DATETIME", Value: "Datetime"},
		{Text: "TIMESTAMP", Value: "Timestamp"},
		{Text: "TEXT", Value: "Text"},
		{Text: "LONGTEXT", Value: "Longtext"},
		{Text: "MEDIUMTEXT", Value: "Mediumtext"},
		{Text: "TINYTEXT", Value: "Tinytext"},
		{Text: "VARCHAR", Value: "Varchar"},
		{Text: "CHAR", Value: "Char"},
		{Text: "BPCHAR", Value: "Bpchar"},
		{Text: "JSON", Value: "Json"},
		{Text: "BLOB", Value: "Blob"},
		{Text: "TINYBLOB", Value: "Tinyblob"},
		{Text: "MEDIUMBLOB", Value: "Mediumblob"},
		{Text: "LONGBLOB", Value: "Longblob"},
		{Text: "INTERVAL", Value: "Interval"},
		{Text: "BOOLEAN", Value: "Boolean"},
		{Text: "Bool", Value: "Bool"},
		{Text: "POINT", Value: "Point"},
		{Text: "LINE", Value: "Line"},
		{Text: "LSEG", Value: "Lseg"},
		{Text: "BOX", Value: "Box"},
		{Text: "PATH", Value: "Path"},
		{Text: "POLYGON", Value: "Polygon"},
		{Text: "CIRCLE", Value: "Circle"},
		{Text: "CIDR", Value: "Cidr"},
		{Text: "INET", Value: "Inet"},
		{Text: "MACADDR", Value: "Macaddr"},
		{Text: "CHARACTER", Value: "Character"},
		{Text: "VARYINGCHARACTER", Value: "Varyingcharacter"},
		{Text: "NCHAR", Value: "Nchar"},
		{Text: "NATIVECHARACTER", Value: "Nativecharacter"},
		{Text: "NVARCHAR", Value: "Nvarchar"},
		{Text: "CLOB", Value: "Clob"},
		{Text: "BINARY", Value: "Binary"},
		{Text: "VARBINARY", Value: "Varbinary"},
		{Text: "ENUM", Value: "Enum"},
		{Text: "SET", Value: "Set"},
		{Text: "GEOMETRY", Value: "Geometry"},
		{Text: "MULTILINESTRING", Value: "Multilinestring"},
		{Text: "MULTIPOLYGON", Value: "Multipolygon"},
		{Text: "LINESTRING", Value: "Linestring"},
		{Text: "MULTIPOINT", Value: "Multipoint"},
		{Text: "GEOMETRYCOLLECTION", Value: "Geometrycollection"},
		{Text: "NAME", Value: "Name"},
		{Text: "UUID", Value: "Uuid"},
		{Text: "TIMESTAMPTZ", Value: "Timestamptz"},
		{Text: "TIMETZ", Value: "Timetz"},
	}
}

func getType(typeName string) string {
	r, _ := regexp.Compile(`\(.*?\)`)
	typeName = r.ReplaceAllString(typeName, "")
	r2, _ := regexp.Compile(`unsigned(.*)`)
	return strings.TrimSpace(strings.Title(strings.ToLower(r2.ReplaceAllString(typeName, ""))))
}
