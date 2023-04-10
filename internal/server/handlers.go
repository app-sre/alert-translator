package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dwelch0/alert-translator/internal/providers/googlechat"
	"github.com/dwelch0/alert-translator/internal/utils"

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

	status := utils.SUCCESS
	defer utils.RecordMetrics(status)

	// Route alert to specified provider for processing
	switch a.provider {
	case GCHAT:
		params, err := collectGChatParameters(r)
		if err != nil {
			log.Println(err)
			status = utils.FAILURE
			return
		}
		err = googlechat.SendAlert(a.httpClient, params, &data)
		if err != nil {
			log.Println(err)
			status = utils.FAILURE
			return
		}
	}
}

func collectGChatParameters(r *http.Request) (*googlechat.QueryParameters, error) {
	space := r.URL.Query().Get("space")
	if space == "" {
		return nil, errors.New("Required query parameter missing: space")
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		return nil, errors.New("Required query parameter missing: key")
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		return nil, errors.New("Required query parameter missing: token")
	}
	return &googlechat.QueryParameters{
		Space: space,
		Key:   key,
		Token: token,
	}, nil
}
