package googlechat

import (
	"bytes"
	"encoding/json"
	"errors"
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

func CollectGChatParameters(r *http.Request) (*QueryParameters, error) {
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
	return &QueryParameters{
		Space: space,
		Key:   key,
		Token: token,
	}, nil
}

// Sends raw alertmanager payload to specified google chat webhook
func SendAlert(client *http.Client, params *QueryParameters, data *template.Data) error {
	alert := format(data)
	messageBytes, err := json.Marshal(*alert.Message)
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

	log.Println(fmt.Sprintf("Alert: `%s` successfully forwarded", alert.Name))
	return nil
}

type GChatAlert struct {
	Name    string
	Message *GChatMessage
}

type GChatMessage struct {
	Text string `json:"text"`
}

// processes raw AlertManager webhook payload and returns google chat formatted messages
func format(data *template.Data) GChatAlert {
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

	return GChatAlert{
		Name:    data.CommonLabels["alertname"],
		Message: &GChatMessage{Text: msgSB.String()},
	}
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
			if val, exists := alert.Labels[l]; exists {
				tmpSB.WriteString(fmt.Sprintf("\t%s: %s\n", l, val))
			}
		}
		for _, a := range annotations {
			if val, exists := alert.Annotations[a]; exists {
				tmpSB.WriteString(fmt.Sprintf("\t%s: %s\n", a, val))
			}
		}
		if tmpSB.Len() > 0 {
			msgs = append(msgs, tmpSB.String())
		}
		tmpSB.Reset()
	}
	return msgs
}
