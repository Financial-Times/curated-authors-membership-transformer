package main

import (
	"fmt"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	app := cli.App("curated-authors-transformer", "A RESTful API for transforming Bertha Curated Authors to UP People JSON")

	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	berthaSrcUrl := app.String(cli.StringOpt{
		Name:   "bertha-source-url",
		Value:  "http://bertha.ig.ft.com/view/publish/gss/1wEdVRLtayZ6-XBfYM3vKAGaOV64cNJD3L8MlLM8-uFY/TestAuthors",
		Desc:   "The URL of the Bertha Authors JSON source",
		EnvVar: "BERTHA_SOURCE_URL",
	})

	app.Action = func() {
		log.Info("App started!!!")
		bs := &berthaService{berthaUrl: *berthaSrcUrl}
		ah := newAuthorHandler(bs)

		setupServiceHandlers(ah)

		log.Infof("Listening on [%d].\n", *port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
		if err != nil {
			log.Printf("Web server failed: [%v].\n", err)
		}
	}

	app.Run(os.Args)
}

func setupServiceHandlers(ah authorHandler) {
	r := mux.NewRouter()
	http.Handle("/", httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
		httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), r)))
	r.HandleFunc(status.PingPath, status.PingHandler)
	r.HandleFunc(status.PingPathDW, status.PingHandler)
	r.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	r.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)
	r.HandleFunc("/__health", v1a.Handler("Topics Transformer Healthchecks", "Checks for accessing TME", ah.HealthCheck()))
	r.HandleFunc("/__gtg", ah.GoodToGo)

	r.HandleFunc("/transformers/authors/__count", ah.getAuthorsCount).Methods("GET")
	r.HandleFunc("/transformers/authors/__ids", ah.getAuthorsUuids).Methods("GET")
	r.HandleFunc("/transformers/authors/{uuid}", ah.getAuthorByUuid).Methods("GET")

	return
}
