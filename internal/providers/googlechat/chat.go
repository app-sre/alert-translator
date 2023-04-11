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
	// collect common values among alerts in payload
	var commonAlertname string
	var commonSeverity string
	var commonDesc string
	if _, exists := data.CommonLabels["alertname"]; exists {
		commonAlertname = data.CommonLabels["alertname"]
	} else {
		commonAlertname = "UNKNOWN_ALERT_NAME"
	}
	if _, exists := data.CommonLabels["severity"]; exists {
		commonSeverity = data.CommonLabels["severity"]
	} else {
		commonSeverity = "UNKNOWN_SEVERITY"
	}
	if _, exists := data.CommonAnnotations["description"]; exists {
		commonDesc = data.CommonAnnotations["description"]
	} else {
		commonDesc = "UNKNOWN_DESCRIPTION"
	}

	// iterate over raws alerts and format
	formattedAlerts := []*GChatMessage{}
	for _, alert := range data.Alerts {
		alertName, exists := alert.Labels["alertname"]
		if !exists {
			alertName = commonAlertname
		}
		severity, exists := alert.Labels["severity"]
		if !exists {
			severity = commonSeverity
		}
		desc, exists := alert.Annotations["description"]
		if !exists {
			desc = commonDesc
		}
		message := fmt.Sprintf("%s [%s]: %s (%s)",
			alertName,
			severity,
			desc,
			alert.Status,
		)
		formattedAlerts = append(formattedAlerts, &GChatMessage{Text: message})
	}
	return formattedAlerts
}
