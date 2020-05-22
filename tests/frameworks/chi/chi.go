package chi

import (
	// add chi adapter
	_ "github.com/HongJaison/go-admin3/adapter/chi"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/table"
	"github.com/HongJaison/themes3/adminlte"

	// add mysql driver
	_ "github.com/HongJaison/go-admin3/modules/db/drivers/mysql"
	// add postgresql driver
	_ "github.com/HongJaison/go-admin3/modules/db/drivers/postgres"
	// add sqlite driver
	_ "github.com/HongJaison/go-admin3/modules/db/drivers/sqlite"
	// add mssql driver
	_ "github.com/HongJaison/go-admin3/modules/db/drivers/mssql"
	// add adminlte ui theme
	_ "github.com/HongJaison/themes3/adminlte"

	"github.com/HongJaison/go-admin3/engine"
	"github.com/HongJaison/go-admin3/plugins/admin"
	"github.com/HongJaison/go-admin3/plugins/example"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/chartjs"
	"github.com/HongJaison/go-admin3/tests/tables"
	"github.com/go-chi/chi"
	"net/http"
	"os"
)

func newHandler() http.Handler {
	r := chi.NewRouter()

	eng := engine.Default()

	adminPlugin := admin.NewAdmin(tables.Generators)
	adminPlugin.AddGenerator("user", tables.GetUserTable)
	examplePlugin := example.NewExample()
	template.AddComp(chartjs.NewChart())

	if err := eng.AddConfigFromJSON(os.Args[len(os.Args)-1]).
		AddPlugins(adminPlugin, examplePlugin).Use(r); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", tables.GetContent)

	return r
}

func NewHandler(dbs config.DatabaseList, gens table.GeneratorList) http.Handler {
	r := chi.NewRouter()

	eng := engine.Default()

	adminPlugin := admin.NewAdmin(gens)
	template.AddComp(chartjs.NewChart())

	if err := eng.AddConfig(config.Config{
		Databases: dbs,
		UrlPrefix: "admin",
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Language:    language.EN,
		IndexUrl:    "/",
		Debug:       true,
		ColorScheme: adminlte.ColorschemeSkinBlack,
	}).
		AddPlugins(adminPlugin).Use(r); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", tables.GetContent)

	return r
}
