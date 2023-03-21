package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type wsDataProxy struct {
	wsUrl         string
	wsConn        *websocket.Conn
	msgRead       chan []byte
	sender        *backend.StreamSender
	done          chan bool
	wsDataSource  *WebSocketDataSource
	readingErrors chan error
}

func NewWsDataProxy(req *backend.RunStreamRequest, sender *backend.StreamSender, ds *WebSocketDataSource) (*wsDataProxy, error) {
	wsDataProxy := &wsDataProxy{
		msgRead:       make(chan []byte),
		sender:        sender,
		done:          make(chan bool, 1),
		wsDataSource:  ds,
		readingErrors: make(chan error),
	}

	url, err := wsDataProxy.encodeURL(req)
	if err != nil {
		return nil, fmt.Errorf("encode URL Error: %s", err.Error())
	}
	wsDataProxy.wsUrl = url

	c, err := wsDataProxy.wsConnect()
	if err != nil {
		return nil, fmt.Errorf("connection Error: %s", err.Error())
	}
	wsDataProxy.wsConn = c

	return wsDataProxy, nil
}

func (wsdp *wsDataProxy) readMessage() {
	defer func() {
		wsdp.wsConn.Close()
		close(wsdp.msgRead)
		log.DefaultLogger.Info("Read Message routine", "detail", "closing websocket connection and msgRead channel")
	}()

	for {
		select {
		case <-wsdp.done:
			return
		default:
			_, message, err := wsdp.wsConn.ReadMessage()
			if err != nil {
				time.Sleep(3 * time.Second)
				wsdp.readingErrors <- fmt.Errorf("%s: %s", "Error reading the websocket", err.Error())
				return
			} else {
				wsdp.msgRead <- message
			}
		}
	}
}

func (wsdp *wsDataProxy) proxyMessage() {
	frame := data.NewFrame("response")
	m := make(map[string]interface{})

	for {
		message, ok := <-wsdp.msgRead
		// if channel was closed
		if !ok {
			return
		}

		json.Unmarshal(message, &m)

		frame.Fields = append(frame.Fields, data.NewField("data", nil, []string{string(message)}))

		err := wsdp.sender.SendFrame(frame, data.IncludeAll)
		if err != nil {
			log.DefaultLogger.Error("Failed to send frame", "error", err)
		}
		frame.Fields = make([]*data.Field, 0)
	}
}

// encodeURl is hard coded with some variables like scheme and x-api-key but will be definetly refactored after changes in the config editor
func (wsdp *wsDataProxy) encodeURL(req *backend.RunStreamRequest) (string, error) {
	var reqJsonData map[string]interface{}

	if err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &reqJsonData); err != nil {
		return "", fmt.Errorf("failed to read JSON Data Source Instance Settings: %s", err.Error())
	}

	host := reqJsonData["url"].(string)

	wsUrl, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("failed to parse host string from Config Editor: %s", err.Error())
	}

	wsUrl.Path = path.Join(wsUrl.Path, req.Path)

	queryParams := url.Values{}
	// add all query parameters to the URL
	for qpName, qpValue := range wsdp.wsDataSource.customQueryParameters {
		queryParams.Add(qpName, qpValue)
	}
	wsUrl.RawQuery = queryParams.Encode()

	return wsUrl.String(), nil
}

func (wsdp *wsDataProxy) wsConnect() (*websocket.Conn, error) {
	log.DefaultLogger.Info("Ws Connect", "connecting to", wsdp.wsUrl)

	customHeaders := http.Header{}
	for headerName, headerValue := range wsdp.wsDataSource.customHeaders {
		customHeaders.Add(headerName, headerValue)
	}

	c, _, err := websocket.DefaultDialer.Dial(wsdp.wsUrl, customHeaders)
	if err != nil {
		return nil, err
	}
	log.DefaultLogger.Info("Ws Connect", "connected to", wsdp.wsUrl)

	return c, nil
}

func sendErrorFrame(msg string, sender *backend.StreamSender) {
	frame := data.NewFrame("error")
	frame.Fields = append(frame.Fields, data.NewField("error", nil, []string{msg}))

	serr := sender.SendFrame(frame, data.IncludeAll)
	if serr != nil {
		log.DefaultLogger.Error("Failed to send error frame", "error", serr)
	}
}
