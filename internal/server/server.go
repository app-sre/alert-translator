package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type api struct {
	provider   string
	webhookUrl string
	httpClient *http.Client
}

const (
	PORT        = "PORT"
	PROVIDER    = "PROVIDER"
	WEBHOOK_URL = "WEBHOOK_URL"
	GCHAT       = "googlechat"
)

func Run() {
	a := initApi()
	http.HandleFunc("/alerts", a.alert)

	port, set := os.LookupEnv(PORT)
	if !set {
		port = "8080"
	}
	log.Println(fmt.Sprintf("Listening on port: %s", port))
	log.Println(fmt.Sprintf("Configured for: %s", a.provider))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func initApi() *api {
	provider, set := os.LookupEnv(PROVIDER)
	if !set {
		provider = GCHAT
	}
	webhookUrl, set := os.LookupEnv(WEBHOOK_URL)
	if !set {
		log.Fatalf("Required env var missing: %s", WEBHOOK_URL)
	}

	return &api{
		provider:   provider,
		httpClient: &http.Client{Timeout: time.Second * 5},
		webhookUrl: webhookUrl,
	}
}
