package googlechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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
	alert := format(data)
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

	log.Println("Translated alert successfully forwarded")
	return nil
}

type GChatMessage struct {
	Text string `json:"text"`
}

// processes raw AlertManager webhook payload and returns google chat formatted messages
func format(data *template.Data) *GChatMessage {
	var msgSB strings.Builder
	// add header
	status := "FIRING"
	if len(data.Alerts.Firing()) == 0 {
		status = "RESOLVED"
	} else {
		status = fmt.Sprintf("%s:%d", status, len(data.Alerts.Firing()))
	}
	msgSB.WriteString(fmt.Sprintf("*ALERT: %s [%s]*\n", data.CommonLabels["alertname"], status))

	// append commons
	for _, pair := range data.CommonLabels.SortedPairs() {
		if pair.Name != "alertname" {
			msgSB.WriteString(fmt.Sprintf("%s: %s\n", pair.Name, pair.Value))
		}
	}
	for _, pair := range data.CommonAnnotations.SortedPairs() {
		msgSB.WriteString(fmt.Sprintf("%s: %s\n", pair.Name, pair.Value))
	}

	labels, annotations := findDiffKeys(data)
	firingMsgs := processUnique(data.Alerts.Firing(), labels, annotations)
	resolvedMsgs := processUnique(data.Alerts.Resolved(), labels, annotations)

	if len(firingMsgs) > 0 {
		msgSB.WriteString("*Firing*:\n")
		for _, firing := range firingMsgs {
			msgSB.WriteString(fmt.Sprintf("- %s", firing))
		}
	}

	if len(resolvedMsgs) > 0 {
		msgSB.WriteString("*Resolved*:\n")
		for _, resolved := range resolvedMsgs {
			msgSB.WriteString(fmt.Sprintf("- %s", resolved))
		}
	}

	return &GChatMessage{Text: msgSB.String()}
}

// AlertManager can group alerts with similiar labels/annotations
// the commonLabels/Annotations can be viewed as the intersection of all labels/annotations
// within the alerts of the payload
// Return keys for labels and annotations that are unique to alerts
func findDiffKeys(data *template.Data) ([]string, []string) {
	if len(data.Alerts) == 1 {
		return nil, nil
	}

	specificLabels := map[string]bool{}
	specificAnnotations := map[string]bool{}
	for _, alert := range data.Alerts {
		for _, l := range alert.Labels.Names() {
			if _, exist := data.CommonLabels[l]; !exist {
				specificLabels[l] = true
			}
		}
		for _, a := range alert.Annotations.Names() {
			if _, exist := data.CommonAnnotations[a]; !exist {
				specificAnnotations[a] = true
			}
		}
	}

	labels := []string{}
	for k := range specificLabels {
		labels = append(labels, k)
	}
	annotations := []string{}
	for k := range specificAnnotations {
		annotations = append(annotations, k)
	}

	return labels, annotations
}

// iterate over a slice of alerts and extract any unique label/annotations
// result is slice of strings where each string is all unique KVs for labels/annotations for an alert
func processUnique(alerts []template.Alert, labels, annotations []string) []string {
	msgs := []string{}
	var tmpSB strings.Builder
	for _, alert := range alerts {
		for _, l := range labels {
			if alert.Labels[l] != "" {
				tmpSB.WriteString(fmt.Sprintf("\t%s: %s\n", l, alert.Labels[l]))
			}
		}
		for _, a := range annotations {
			if alert.Annotations[a] != "" {
				tmpSB.WriteString(fmt.Sprintf("\t%s: %s\n", a, alert.Annotations[a]))
			}
		}
		if tmpSB.Len() > 0 {
			msgs = append(msgs, tmpSB.String())
		}
		tmpSB.Reset()
	}
	return msgs
}
