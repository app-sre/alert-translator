package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/app-sre/alert-translator/internal/utils"
	"github.com/app-sre/alert-translator/pkg/providers/googlechat"

	"github.com/prometheus/alertmanager/template"
)

func (a *api) alert(w http.ResponseWriter, r *http.Request) {
	status := utils.FAILURE
	defer func() {
		utils.RecordMetrics(status)
	}()

	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body")
		return
	}
	defer r.Body.Close()

	// Unmarshal request body into the Data struct
	var data template.Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("Error unmarshaling JSON")
		return
	}

	// Route alert to specified provider for processing
	switch a.provider {
	case GCHAT:
		params, err := googlechat.CollectGChatParameters(r)
		if err != nil {
			log.Println(err)
			return
		}
		err = googlechat.SendAlert(a.httpClient, params, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}

	status = utils.SUCCESS
}
