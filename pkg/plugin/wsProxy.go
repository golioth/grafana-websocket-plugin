package plugin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type wsDataProxy struct {
	wsUrl  string
	wsConn *websocket.Conn // conexao ws com o endpoint
	sender *backend.StreamSender
	done   chan bool
}

func NewWsDataProxy(req *backend.RunStreamRequest, sender *backend.StreamSender) (*wsDataProxy, error) {
	url, err := encodeURL(req)
	if err != nil {
		log.DefaultLogger.Error("Encode URL Error", "error", err.Error())
		return nil, err
	}

	c, err := wsConnect(url)
	if err != nil {
		log.DefaultLogger.Error("webSocket Connection Error", "error", err.Error())
		return nil, err
	}

	return &wsDataProxy{
		wsUrl:  url,
		wsConn: c,
		sender: sender,
		done:   make(chan bool, 1),
	}, nil
}

func (wsp *wsDataProxy) startDataProxy() {
	defer close(wsp.done)

	frame := data.NewFrame("response")
	m := make(map[string]interface{})
	for {
		select {
		case <-wsp.done:
			wsp.wsConn.Close()
			return
		default:
			_, message, err := wsp.wsConn.ReadMessage()
			if err != nil {
				wsp.wsConn, err = wsConnect(wsp.wsUrl)
				if err != nil {
					return
				}
			}

			json.Unmarshal(message, &m)

			frame.Fields = append(frame.Fields, data.NewField("time", nil, []time.Time{time.Now()}))
			frame.Fields = append(frame.Fields, data.NewField("data", nil, []string{string(message)}))

			// Kept this commented block while in dev mode, will be removed before release
			// logData := m["result"].(map[string]interface{})["data"].(map[string]interface{})
			// frame.Fields = append(frame.Fields, data.NewField("deviceId", nil, []string{logData["deviceId"].(string)}))
			// newfield := data.NewFieldFromFieldType(data.FieldTypeFor(logData["counter"]), 1)
			// newfield.Name = "counter"
			// newfield.Set(0, logData["counter"])
			// log.DefaultLogger.Info("new field: ", newfield)
			// frame.Fields = append(frame.Fields, newfield)
			// newfield2 := data.NewFieldFromFieldType(data.FieldTypeFor(logData["env"].(map[string]interface{})["test"]), 1)
			// newfield2.Name = "envTest"
			// newfield2.Set(0, logData["env"].(map[string]interface{})["test"])
			// log.DefaultLogger.Info("new field: ",calor-demais/devices/61b0b02e95fd466888055ca4/datadashboard")

			err = wsp.sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
			frame.Fields = make([]*data.Field, 0)
		}
	}
}

// encodeURl is hard coded with some variables like scheme and x-api-key but will be definetly refactored after changes in the config editor
func encodeURL(req *backend.RunStreamRequest) (string, error) {
	var reqJsonData map[string]interface{}

	if err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &reqJsonData); err != nil {
		return "", fmt.Errorf("failed reading JSON Data Source Instance Settings: %s", err.Error())
	}

	host := reqJsonData["host"].(string)
	// with url.Parse it's possible to set Host as "scheme://host/prefixPath" in the Config Editor (more flexibility)
	// u := url.URL{Scheme: "wss", Host: host, Path: req.Path}
	wsUrl, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("failed parsing host string from Config Editor: %s", err.Error())
	}

	wsUrl.Path = concatWithoutCharDuplicity(wsUrl.Path, req.Path, "/")

	apiKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]
	queryParams := url.Values{}
	queryParams.Add("x-api-key", apiKey)
	wsUrl.RawQuery = queryParams.Encode()

	return wsUrl.String(), nil
}

func concatWithoutCharDuplicity(str1, str2, char string) string {
	if str1 != "" && str2 != "" {
		str1LastChar := string(str1[len(str1)-1])
		str2FirstChar := string(str2[0])

		if str1LastChar == char && str2FirstChar == char {
			return str1 + str2[1:]
		}

		if str1LastChar != char && str2FirstChar != char {
			return str1 + char + str2
		}
	}

	return str1 + str2
}

func wsConnect(url string) (*websocket.Conn, error) {
	log.DefaultLogger.Info("connecting to", "url", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	log.DefaultLogger.Info("websocket connected to", "url", url)

	return c, nil
}
