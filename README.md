# Alert Translator
Translation layer between AlertManager and destination applications.

The `/alerts` endpoint handles POST requests from an AlertManager `webhook_config` receiver

## Environment Variables
### Optional
* PORT - defaults to `8080`
* PROVIDER - defaults to `googlechat`
    * available options:
        * `googlechat`

## Local Testing
### Sample payload command
```
curl -X POST -H "Content-Type: application/json" -d '{
  "version": "4",
  "groupKey": 123456789,
  "status": "firing",
  "receiver": "googlechat",
  "groupLabels": {
    "alertname": "TestAlert"
  },
  "commonLabels": {
    "alertname": "TestAlert",
    "severity": "critical"
  },
  "commonAnnotations": {
    "description": "This is a test alert"
  },
  "externalURL": "http://localhost:9093",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "TestAlert",
        "severity": "critical",
        "instance": "example_instance"
      },
      "annotations": {
        "description": "This is a test alert"
      },
      "startsAt": "2023-04-06T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://localhost:9090/graph?g0.expr=vector%281%29&g0.tab=1"
    }
  ]
}' "http://localhost:8080/alerts?space=PLACEHOLDER&key=PLACEHOLDER&token=PLACEHOLDER"

```
## Metrics
### alert_translator_handled_alerts
Counter incremented with every request to `/alerts`
Labels:
* status
  * success
  * failure

