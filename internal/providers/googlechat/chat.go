package googlechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/alertmanager/template"
)

type QueryParameters struct {
	Space string
	Key   string
	Token string
}

const BASE_URL = "https://chat.googleapis.com"

// Sends raw alertmanager payload to specified google chat webhook
func SendAlert(client *http.Client, params *QueryParameters, data *template.Data) error {
	alerts := format(data)
	for _, alert := range alerts {
		messageBytes, err := json.Marshal(*alert)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/v1/spaces/%s/messages?key=%s&token=%s",
			BASE_URL,
			params.Space,
			params.Key,
			params.Token,
		)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(messageBytes))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		resp, err := client.Do(req)
		resp.Body.Close()
		if err != nil {
			return err
		}
	}
	log.Println("Translated alert successfully forwarded")
	return nil
}

type GChatMessage struct {
	Text string `json:"text"`
}

// processes raw AlertManager webhook payload and returns google chat formatted messages
func format(data *template.Data) []*GChatMessage {
	formattedAlerts := []*GChatMessage{}
	for _, alert := range data.Alerts {
		alertName, exists := alert.Labels["alertname"]
		if !exists {
			alertName = "UNKNOWN_NAME"
		}
		alertDesc, exists := alert.Annotations["description"]
		if !exists {
			alertDesc = "UNKNOWN_DESCRIPTION"
		}
		message := fmt.Sprintf("%s: %s (%s)",
			alertName,
			alertDesc,
			alert.Status,
		)
		formattedAlerts = append(formattedAlerts, &GChatMessage{Text: message})
	}
	return formattedAlerts
}
