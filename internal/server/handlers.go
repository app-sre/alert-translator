package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dwelch0/alert-translator/internal/providers/googlechat"

	"github.com/prometheus/alertmanager/template"
)

func (a *api) alert(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal request body into the Data struct
	var data template.Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Error unmarshaling JSON", http.StatusBadRequest)
		return
	}

	// Route alert to specified provider for processing
	switch a.provider {
	case GCHAT:
		err = googlechat.SendAlert(a.httpClient, a.webhookUrl, &data)
		if err != nil {
			log.Println(err)
		}
		//TODO: send failure/success to metrics endpoint
	}
}

// TODO: metrics handler
