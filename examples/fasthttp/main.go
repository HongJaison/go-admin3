package main

import (
	"log"
	"os"
	"os/signal"

	_ "github.com/HongJaison/go-admin3/adapter/fasthttp"
	_ "github.com/HongJaison/go-admin3/modules/db/drivers/mysql"
	_ "github.com/HongJaison/themes3/adminlte"

	"github.com/HongJaison/go-admin3/engine"
	"github.com/HongJaison/go-admin3/examples/datamodel"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/modules/language"
	"github.com/HongJaison/go-admin3/plugins/example"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/chartjs"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func main() {
	router := fasthttprouter.New()

	eng := engine.Default()

	cfg := config.Config{
		Databases: config.DatabaseList{
			"default": {
				Host:       "127.0.0.1",
				Port:       "3306",
				User:       "root",
				Pwd:        "root",
				Name:       "godmin",
				MaxIdleCon: 50,
				MaxOpenCon: 150,
				Driver:     config.DriverMysql,
			},
		},
		UrlPrefix: "admin",
		IndexUrl:  "/",
		Store: config.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Debug:    true,
		Language: language.CN,
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
		Use(router); err != nil {
		panic(err)
	}

	router.ServeFiles("/uploads/*filepath", "./uploads")

	eng.HTML("GET", "/admin", datamodel.GetContent)

	go func() {
		_ = fasthttp.ListenAndServe(":8897", router.Handler)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("closing database connection")
	eng.MysqlConnection().Close()
}