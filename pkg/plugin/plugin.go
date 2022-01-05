package plugin

import (
	"context"
	"encoding/json"
	"math/rand"

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
	WithStreaming bool   `json:"withStreaming"`
	WsPath        string `json:"path"`
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
	status := backend.SubscribeStreamStatusOK

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *WebSocketDataSource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	wsDataProxy, err := NewWsDataProxy(req, sender)
	if err != nil {
		log.DefaultLogger.Error("error instantiating new webSocket data proxy", "error", err.Error())
		return err
	}
	defer wsDataProxy.wsConn.Close()

	// go wsDataProxy.startDataProxy()

	go wsDataProxy.proxyMessage()

	go wsDataProxy.readMessage()

	<-ctx.Done()

	wsDataProxy.done <- true

	log.DefaultLogger.Info("Closing Channel", "channel", req.Path)

	return nil
}

// PublishStream is called when a client sends a message to the stream.
func (d *WebSocketDataSource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
