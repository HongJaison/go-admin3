package main

import (
	"log"
	"os"
	"os/signal"

	_ "github.com/HongJaison/go-admin3/adapter/beego"
	_ "github.com/HongJaison/go-admin3/modules/db/drivers/mssql"

	"github.com/HongJaison/go-admin3/engine"
	"github.com/HongJaison/go-admin3/examples/datamodel"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/plugins/example"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/chartjs"
	"github.com/HongJaison/themes3/adminlte"
	"github.com/astaxie/beego"
)

func main() {
	app := beego.NewApp()

	eng := engine.Default()

	cfg := config.Config{
		Databases: config.DatabaseList{
			"default": {
				Host:       "172.16.74.222",
				Port:       "1433",
				User:       "sa",
				Pwd:        "Akduifwkro1988",
				Name:       "goadmin_site",
				MaxIdleCon: 50,
				MaxOpenCon: 150,
				Driver:     config.DriverMssql,
			},
		},
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		UrlPrefix:   "admin",
		IndexUrl:    "/",
		Debug:       true,
		Language:    language.CN,
		ColorScheme: adminlte.ColorschemeSkinBlack,
	}

	template.AddComp(chartjs.NewChart())

	// customize a plugin

	examplePlugin := example.NewExample()

	// load from golang.Plugin
	//
	// examplePlugin := plugins.LoadFromPlugin("../datamodel/example.so")

	// customize the login page
	// example: https://github.com/HongJaison/demo.go-admin.cn/blob/master/main.go#L39
	//
	// template.AddComp("login", datamodel.LoginPage)

	// load config from json file
	//
	// eng.AddConfigFromJSON("../datamodel/config.json")

	beego.SetStaticPath("/uploads", "uploads")

	if err := eng.AddConfig(cfg).
		AddGenerators(datamodel.Generators).
		AddDisplayFilterXssJsFilter().
		// add generator, first parameter is the url prefix of table when visit.
		// example:
		//
		// "user" => http://localhost:9033/admin/info/user
		//
		AddGenerator("user", datamodel.GetUserTable).
		AddPlugins(examplePlugin).
		Use(app); err != nil {
		panic(err)
	}

	// you can custom your pages like:

	eng.HTML("GET", "/admin", datamodel.GetContent)

	beego.BConfig.Listen.HTTPAddr = "127.0.0.1"
	beego.BConfig.Listen.HTTPPort = 9088
	go app.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("closing database connection")
	eng.MysqlConnection().Close()
}
