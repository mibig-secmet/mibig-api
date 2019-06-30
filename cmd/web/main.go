package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pelletier/go-toml"
	zap "go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
)

type mail struct {
	username  string
	password  string
	host      string
	port      int64
	recipient string
}

type application struct {
	logger     *zap.SugaredLogger
	MibigModel *postgres.MibigModel
	BuildTime  string
	GitVersion string
	Mail       mail
}

type config struct {
	Addr        string
	DatabaseUri string
	Mail        mail
}

var (
	buildTime string
	gitVer    string
)

func main() {
	var configFile string
	var debug bool
	var version bool

	flag.StringVar(&configFile, "config", "config.toml", "Path to the config file")
	flag.BoolVar(&debug, "debug", false, "Debug level logging")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.Parse()

	if version {
		fmt.Printf("Git version: %s\nBuild time: %s\n", gitVer, buildTime)
		os.Exit(0)
	}

	logger := setupLogging(debug)
	defer logger.Sync()

	conf, err := createConfig(configFile)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	db, err := initDb(conf)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	app := &application{
		logger:     logger,
		MibigModel: &postgres.MibigModel{DB: db},
		BuildTime:  buildTime,
		GitVersion: gitVer,
		Mail:       conf.Mail,
	}

	mux := app.routes()

	logger.Infow("starting server",
		"address", conf.Addr,
	)
	err = http.ListenAndServe(conf.Addr, mux)
	logger.Fatalf(err.Error())
}

func setupLogging(debug bool) *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if debug {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to set up logging: %s", err.Error())
	}
	return logger.Sugar()
}

func createConfig(filename string) (*config, error) {

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	tomlConf, err := toml.LoadFile(filename)
	if err != nil {
		return nil, err
	}

	conf := config{
		Addr:        tomlConf.GetDefault("address", ":6424").(string),
		DatabaseUri: tomlConf.GetDefault("database.uri", "host=localhost port=5432 user=postgres password=secret dbname=mibig sslmode=disable").(string),
		Mail: mail{
			username:  tomlConf.Get("mail.user").(string),
			password:  tomlConf.Get("mail.password").(string),
			host:      tomlConf.Get("mail.host").(string),
			port:      tomlConf.Get("mail.port").(int64),
			recipient: tomlConf.Get("mail.recipient").(string),
		},
	}

	return &conf, nil
}

func initDb(conf *config) (*sql.DB, error) {
	db, err := sql.Open("postgres", conf.DatabaseUri)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
