package gin

import (
	// add gin adapter
	ada "github.com/HongJaison/go-admin3/adapter/gin"
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
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/table"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/chartjs"
	"github.com/HongJaison/go-admin3/tests/tables"
	"github.com/HongJaison/themes3/adminlte"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
)

func newHandler() http.Handler {
	r := gin.Default()

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	eng := engine.Default()

	template.AddComp(chartjs.NewChart())

	if err := eng.AddConfigFromJSON(os.Args[len(os.Args)-1]).
		AddGenerators(tables.Generators).
		AddGenerator("user", tables.GetUserTable).
		Use(r); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", tables.GetContent)

	r.Static("/uploads", "./uploads")

	return r
}

func NewHandler(dbs config.DatabaseList, gens table.GeneratorList) http.Handler {
	r := gin.Default()

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	eng := engine.Default()

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
		AddAdapter(new(ada.Gin)).
		AddGenerators(gens).
		Use(r); err != nil {
		panic(err)
	}

	eng.HTML("GET", "/admin", tables.GetContent)

	r.Static("/uploads", "./uploads")

	return r
}
