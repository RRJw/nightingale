package sender

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ccfos/nightingale/v6/models"

	"github.com/toolkits/pkg/logger"
)

type ReqJson struct {
	Business           string `json:"business"`
	Title              string `json:"title"`
	Host               string
	Application        string `json:"application"`
	Application_filter int    `json:"application_filter"`
	Item               string `json:"item"`
	Item_value         string `json:"item_value"`
	AlarmValue         string
	Content            string `json:"content"`
	Level              int    `json:"level"`
	Status             int    `json:"status"`
	Auto_Recover       int    `json:"auto_Recover"`
}

func SendWebhooks(webhooks []*models.Webhook, event *models.AlertCurEvent) {
	for _, conf := range webhooks {
		if conf.Url == "" || !conf.Enable {
			continue
		}

		reqData := ReqJson{
			Business:           "中台",
			Title:              "告警标题",
			Host:               "告警主机",
			Application:        "告警类别",
			Application_filter: 2,
			Item:               "告警项",
			Item_value:         "监控指标值",
			AlarmValue:         "告警阈值",
			Content:            "测试",
			Level:              1,
			Status:             1,
			Auto_Recover:       1,
		}

		bs, err := json.Marshal(reqData)
		if err != nil {
			continue
		}

		bf := bytes.NewBuffer(bs)

		req, err := http.NewRequest("POST", conf.Url, bf)
		if err != nil {
			logger.Warning("alertingWebhook failed to new request", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		if conf.BasicAuthUser != "" && conf.BasicAuthPass != "" {
			req.SetBasicAuth(conf.BasicAuthUser, conf.BasicAuthPass)
		}

		if len(conf.Headers) > 0 && len(conf.Headers)%2 == 0 {
			for i := 0; i < len(conf.Headers); i += 2 {
				if conf.Headers[i] == "host" {
					req.Host = conf.Headers[i+1]
					continue
				}
				req.Header.Set(conf.Headers[i], conf.Headers[i+1])
			}
		}

		// todo add skip verify
		client := http.Client{
			Timeout: time.Duration(conf.Timeout) * time.Second,
		}

		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			logger.Errorf("event_webhook_fail, ruleId: [%d], eventId: [%d], url: [%s], error: [%s]", event.RuleId, event.Id, conf.Url, err)
			continue
		}

		var body []byte
		if resp.Body != nil {
			defer resp.Body.Close()
			body, _ = ioutil.ReadAll(resp.Body)
		}

		logger.Debugf("event_webhook_succ, url: %s, response code: %d, body: %s", conf.Url, resp.StatusCode, string(body))
	}
}
