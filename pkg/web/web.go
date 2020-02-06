package web

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	zap "go.uber.org/zap"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
)

type application struct {
	logger      *zap.SugaredLogger
	MibigModel  models.MibigModel
	LegacyModel models.LecagyModel
	Mail        models.EmailSender
	Mux         *gin.Engine
}

func Run(debug bool) {

	if !debug {
		// set Gin to release mode
		gin.SetMode(gin.ReleaseMode)
	}

	logger := setupLogging(debug)
	defer logger.Sync()

	db, err := initDb(viper.GetString("database.uri"))
	if err != nil {
		logger.Fatalf(err.Error())
	}

	legacy_db, err := initDb(viper.GetString("legacy_database.uri"))
	if err != nil {
		logger.Fatalf(err.Error())
	}

	mailConfig := models.MailConfig{
		Host:      viper.GetString("mail.host"),
		Port:      viper.GetInt("mail.port"),
		Username:  viper.GetString("mail.username"),
		Password:  viper.GetString("mail.password"),
		Recipient: viper.GetString("mail.recipient"),
	}

	mailSender := models.NewProductionSender(mailConfig)
	mux := setupMux(debug, logger.Desugar())

	app := &application{
		logger:      logger,
		MibigModel:  &postgres.MibigModel{DB: db},
		LegacyModel: &postgres.LegacyModel{DB: legacy_db},
		Mail:        mailSender,
		Mux:         mux,
	}

	mux = app.routes()

	address := fmt.Sprintf("%s:%d", viper.GetString("server.address"), viper.GetInt("server.port"))

	logger.Infow("starting server",
		"address", address,
	)
	err = http.ListenAndServe(address, mux)
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

func initDb(dbUri string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbUri)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
