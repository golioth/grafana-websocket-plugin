package plugin

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

// Make sure WebSocketDataSource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler = (*WebSocketDataSource)(nil)
	// _ backend.CheckHealthHandler    = (*WebSocketDataSource)(nil)
	_ backend.StreamHandler         = (*WebSocketDataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*WebSocketDataSource)(nil)
)

// NewWebSocketDataSource creates a new datasource instance.
func NewWebSocketDataSource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &WebSocketDataSource{}, nil
}

// WebSocketDataSource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type WebSocketDataSource struct {
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewWebSocketDataSource factory function.
func (d *WebSocketDataSource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *WebSocketDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	log.DefaultLogger.Info("QueryData called", "response", response)
	return response, nil
}

type queryModel struct {
	WithStreaming bool    `json:"withStreaming"`
	Constant      float64 `json:"constant"`
	WsPath        string  `json:"path"`
}

func (d *WebSocketDataSource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	log.DefaultLogger.Info("query called", "pluginCtx", pCtx)
	log.DefaultLogger.Info("query called", "query", query)
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response.
	frame := data.NewFrame("response")

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      qm.WsPath,
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}
	// add the frames to the response.
	response.Frames = append(response.Frames, frame)
	log.DefaultLogger.Info("query called", "response", response)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *WebSocketDataSource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if rand.Int()%2 == 0 {
		status = backend.HealthStatusError
		message = "randomized error"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *WebSocketDataSource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	// status := backend.SubscribeStreamStatusPermissionDenied
	// if req.Path == "deviceState" || req.Path == "timeSeries" || req.Path == "logs" {
	// Allow subscribing only on expected path.
	status := backend.SubscribeStreamStatusOK

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *WebSocketDataSource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	url := encodeURL(req)

	c, err := wsConnect(url)
	if err != nil {
		return err
	}
	defer c.Close()

	done := make(chan struct{})

	go func(done chan struct{}) {
		// Create the same data frame as for query data.
		frame := data.NewFrame("response")

		m := make(map[string]interface{})

		defer close(done)

		for {
			select {
			case <-done:
				return
			default:
				_, message, _ := c.ReadMessage()
				if err != nil {
					c, err := wsConnect(url)
					if err != nil {
						return
					}
					defer c.Close()
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

				err = sender.SendFrame(frame, data.IncludeAll)
				if err != nil {
					log.DefaultLogger.Error("Error sending frame", "error", err)
					continue
				}
				frame.Fields = make([]*data.Field, 0)
			}
		}
	}(done)

	<-ctx.Done()

	done <- struct{}{}

	log.DefaultLogger.Info("Closing Channel", "channel", req.Path)

	return nil
}

func encodeURL(req *backend.RunStreamRequest) string {
	apiKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]
	var reqJsonData map[string]interface{}
	json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &reqJsonData)
	host := reqJsonData["host"].(string)
	params := url.Values{}
	params.Add("x-api-key", apiKey)
	u := url.URL{Scheme: "wss", Host: host, Path: req.Path}
	u.RawQuery = params.Encode()

	return u.String()
}

func wsConnect(url string) (*websocket.Conn, error) {
	log.DefaultLogger.Info("connecting to", "url", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.DefaultLogger.Error("webSocket Connection Error", "error", err.Error())
		return nil, err
	}
	log.DefaultLogger.Info("websocket connected to", "url", url)

	return c, nil
}

// PublishStream is called when a client sends a message to the stream.
func (d *WebSocketDataSource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
