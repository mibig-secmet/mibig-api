package main

import (
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	zap "go.uber.org/zap"
	"log"
	"net/http"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
)

type application struct {
	logger     *zap.SugaredLogger
	MibigModel *postgres.MibigModel
}

type config struct {
	Addr        string
	DatabaseUri string
	Debug       bool
}

func main() {
	conf := new(config)
	flag.StringVar(&conf.Addr, "addr", ":6424", "HTTP network address")
	flag.StringVar(&conf.DatabaseUri, "db", "postgres://postgres:secret@localhost/mibig?sslmode=disable", "PostgreSQL database uri")
	flag.BoolVar(&conf.Debug, "debug", false, "Debug level logging")
	flag.Parse()

	logger := setupLogging(conf)
	defer logger.Sync()

	db, err := initDb(conf)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	app := &application{
		logger:     logger,
		MibigModel: &postgres.MibigModel{DB: db},
	}

	mux := app.routes()

	logger.Infow("starting server",
		"address", conf.Addr,
	)
	err = http.ListenAndServe(conf.Addr, mux)
	logger.Fatalf(err.Error())
}

func setupLogging(conf *config) *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if conf.Debug {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("Failed to set up logging: %s", err.Error())
	}
	return logger.Sugar()
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
