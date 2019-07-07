package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pelletier/go-toml"
	zap "go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
	"time"
)

type application struct {
	logger      *zap.SugaredLogger
	MibigModel  models.MibigModel
	LegacyModel models.LecagyModel
	BuildTime   string
	GitVersion  string
	Mail        models.EmailSender
	Mux         *gin.Engine
}

type config struct {
	Addr        string
	DatabaseUri string
	Mail        models.MailConfig
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

	if !debug {
		// set Gin to release mode
		gin.SetMode(gin.ReleaseMode)
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

	mailSender := models.NewProductionSender(conf.Mail)
	mux := setupMux(debug, logger.Desugar())

	app := &application{
		logger:      logger,
		MibigModel:  &postgres.MibigModel{DB: db},
		LegacyModel: &postgres.LegacyModel{DB: db},
		BuildTime:   buildTime,
		GitVersion:  gitVer,
		Mail:        mailSender,
		Mux:         mux,
	}

	mux = app.routes()

	logger.Infow("starting server",
		"address", conf.Addr,
	)
	err = http.ListenAndServe(conf.Addr, mux)
	logger.Fatalf(err.Error())
}

func setupMux(debug bool, logger *zap.Logger) *gin.Engine {
	var mux *gin.Engine
	if !debug {
		// In production mode, use zap Logger middleware
		mux = gin.New()
		mux.Use(ginzap.Ginzap(logger, time.RFC3339, true))
		mux.Use(ginzap.RecoveryWithZap(logger, true))
	} else {
		// otherwise use the default Gin logging, which is prettier
		mux = gin.Default()
	}
	return mux
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
		Mail: models.MailConfig{
			Username:  tomlConf.Get("mail.user").(string),
			Password:  tomlConf.Get("mail.password").(string),
			Host:      tomlConf.Get("mail.host").(string),
			Port:      tomlConf.Get("mail.port").(int64),
			Recipient: tomlConf.Get("mail.recipient").(string),
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
