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

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler = (*SampleDatasource)(nil)
	// _ backend.CheckHealthHandler    = (*SampleDatasource)(nil)
	_ backend.StreamHandler         = (*SampleDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SampleDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &SampleDatasource{}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SampleDatasource struct {
	teste string
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SampleDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// var jsonData map[string]interface{}
	// json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &jsonData)
	// jsonData["testeKey"] = "testando........"
	// req.PluginContext.DataSourceInstanceSettings.JSONData, _ = json.Marshal(jsonData)
	// log.DefaultLogger.Info("QueryData called", "jsonData", jsonData)
	// log.DefaultLogger.Info("QueryData called", "request after modification", jsonData)

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

func (d *SampleDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
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

	// a principio para graficos de streaming, nao preciso retornar um frame inicial....bobeira
	// add fields.
	// frame.Fields = append(frame.Fields,
	// 	data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
	// 	data.NewField("values", nil, []int64{10, 20}),
	// )

	log.DefaultLogger.Info("query called", "acessando o conteudo de JSON da query para pegar a constante vindo do forntend", query.JSON)

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      qm.WsPath, /* path, */ //"stream",

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
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
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
func (d *SampleDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	// status := backend.SubscribeStreamStatusPermissionDenied
	// if req.Path == "deviceState" || req.Path == "timeSeries" || req.Path == "logs" {
	// Allow subscribing only on expected path.
	status := backend.SubscribeStreamStatusOK
	// }
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {

	log.DefaultLogger.Info("RunStream called", "request", req)
	log.DefaultLogger.Info("RunStream called", "path to listen to", req.Path)
	log.DefaultLogger.Info("RunStream called", "acessando variavel do sampledatasource d.teste:", d.teste)

	apiKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["apiKey"]

	var reqJsonData map[string]interface{}
	json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &reqJsonData)

	host := reqJsonData["host"].(string)

	// connect to ws to rcv data
	u := url.URL{Scheme: "wss", Host: host /* "api.golioth.net" */ /* "localhost:9080" */, Path: req.Path}
	params := url.Values{}
	params.Add("x-api-key", apiKey)
	u.RawQuery = params.Encode()

	log.DefaultLogger.Info("golioth websocket connecting to ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// log.DefaultLogger.Info(" golioth websocket dial c", "info", c)
	if err != nil {
		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	}
	defer c.Close()

	log.DefaultLogger.Info("golioth websocket connected to ", u.String())

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")
	// Add fields (matching the same schema used in QueryData).
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		// data.NewField("type", nil, make([]string, 1)),
		// data.NewField("level", nil, make([]string, 1)),
		// data.NewField("module", nil, make([]string, 1)),
		// data.NewField("message", nil, make([]string, 1)),
		// data.NewField("deviceId", nil, make([]string, 1)),
		data.NewField("data", nil, make([]string, 1)),
	)

	done := make(chan struct{})

	go func() {
		m := make(map[string]interface{})
		// var m interface{}
		defer close(done)
		for {
			_, message, _ := c.ReadMessage()
			// if err != nil {
			// 	log.DefaultLogger.Error("golioth websocket read:", "error", err)
			// 	// return
			// 	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			// 	if err != nil {
			// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
			// 	}
			// 	log.DefaultLogger.Info("golioth websocket reconnected")
			// }
			log.DefaultLogger.Info("Leu Logs? %+v", m)
			json.Unmarshal(message, &m)
			logData := m["result"].(map[string]interface{})["data"].(map[string]interface{})
			log.DefaultLogger.Info("Reading Logs %+v", m)
			frame.Fields[0].Set(0, time.Now())
			// frame.Fields[1].Set(0, logData["type"].(string))
			// frame.Fields[2].Set(0, logData["level"].(string))
			// frame.Fields[3].Set(0, logData["module"].(string))
			// frame.Fields[4].Set(0, logData["message"].(string))
			// frame.Fields[5].Set(0, logData["deviceId"].(string))

			l, _ := json.Marshal(logData)
			frame.Fields[1].Set(0, string(l))
			log.DefaultLogger.Info("Sending Time Series")

			err = sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}()

	<-ctx.Done()

	log.DefaultLogger.Info("encerrando esse runstream ")
	// if req.Path == "deviceState" {
	// 	// Create the same data frame as for query data.
	// 	frame := data.NewFrame("response")

	// 	// Add fields (matching the same schema used in QueryData).
	// 	frame.Fields = append(frame.Fields,
	// 		data.NewField("time", nil, make([]time.Time, 1)),
	// 		data.NewField("counter", nil, make([]int64, 1)),
	// 		data.NewField("func", nil, make([]int64, 1)),
	// 	)

	// 	// connect to ws to rcv data
	// 	u := url.URL{Scheme: "wss", Host: "api.golioth.net" /* "localhost:9080" */, Path: "/v1/ws/projects/calor-demais/devices/61b0b02e95fd466888055ca4/data"}

	// 	params := url.Values{}
	// 	params.Add("x-api-key", "KeY1MgDcENYZuFIjsi9pKRm03O0QGMvO")
	// 	u.RawQuery = params.Encode()

	// 	// u := url.URL{Scheme: "wss", Host: "localhost:9080", Path: "/v1/ws/projects/local-teste/devices/61a4cfa0b2b45578105aeca1/data"}

	// 	// u := url.URL{Scheme: "ws", Host: "localhost:9080", Path: "/v1/ws/projects/local-teste/stream"}

	// 	log.DefaultLogger.Info("golioth websocket connecting to ", u.String())

	// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// 	// log.DefaultLogger.Info(" golioth websocket dial c", "info", c)
	// 	if err != nil {
	// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	// 	}
	// 	defer c.Close()

	// 	done := make(chan struct{})

	// 	go func() {
	// 		m := make(map[string]interface{})
	// 		defer close(done)
	// 		for {
	// 			_, message, _ := c.ReadMessage()
	// 			// if err != nil {
	// 			// 	log.DefaultLogger.Error("golioth websocket read:", "error", err)
	// 			// 	// return
	// 			// 	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	// 			// 	if err != nil {
	// 			// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	// 			// 	}
	// 			// 	log.DefaultLogger.Info("golioth websocket reconnected")
	// 			// }
	// 			json.Unmarshal(message, &m)
	// 			// log.DefaultLogger.Info("Reading Device State Changes %+v", m)

	// 			frame.Fields[0].Set(0, time.Now())
	// 			frame.Fields[1].Set(0, int64(m["result"].(map[string]interface{})["data"].(map[string]interface{})["counter"].(float64)))
	// 			frame.Fields[2].Set(0, int64(m["result"].(map[string]interface{})["data"].(map[string]interface{})["env"].(map[string]interface{})["test"].(float64)))

	// 			// log.DefaultLogger.Info("Sending Device State")

	// 			err = sender.SendFrame(frame, data.IncludeAll)
	// 			if err != nil {
	// 				log.DefaultLogger.Error("Error sending frame", "error", err)
	// 				continue
	// 			}
	// 		}
	// 	}()
	// } else if req.Path == "timeSeries" {
	// 	// Create the same data frame as for query data.
	// 	frame := data.NewFrame("response")

	// 	// Add fields (matching the same schema used in QueryData).
	// 	frame.Fields = append(frame.Fields,
	// 		data.NewField("time", nil, make([]time.Time, 1)),
	// 		data.NewField("temp", nil, make([]float64, 1)),
	// 		// data.NewField("func", nil, make([]int64, 1)),
	// 	)

	// 	u := url.URL{Scheme: "ws", Host: "localhost:9080", Path: "/v1/ws/projects/local-teste/stream"}

	// 	log.DefaultLogger.Info("golioth websocket connecting to ", u.String())

	// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// 	// log.DefaultLogger.Info(" golioth websocket dial c", "info", c)
	// 	if err != nil {
	// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	// 	}
	// 	defer c.Close()<-ctx.Done()

	// 	done := make(chan struct{})

	// 	go func() {
	// 		m := make(map[string]interface{})
	// 		defer close(done)
	// 		for {
	// 			_, message, _ := c.ReadMessage()
	// 			// if err != nil {
	// 			// 	log.DefaultLogger.Error("golioth websocket read:", "error", err)
	// 			// 	// return
	// 			// 	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	// 			// 	if err != nil {
	// 			// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	// 			// 	}
	// 			// 	log.DefaultLogger.Info("golioth websocket reconnected")
	// 			// }
	// 			json.Unmarshal(message, &m)
	// 			log.DefaultLogger.Info("Reading Time Series %+v", m)
	// 			frame.Fields[0].Set(0, time.Now())
	// 			frame.Fields[1].Set(0, float64(m["result"].(map[string]interface{})["data"].(map[string]interface{})["data"].(map[string]interface{})["temp"].(float64)))

	// 			log.DefaultLogger.Info("Sending Time Series")

	// 			err = sender.SendFrame(frame, data.IncludeAll)
	// 			if err != nil {
	// 				log.DefaultLogger.Error("Error sending frame", "error", err)
	// 				continue
	// 			}
	// 		}
	// 	}()

	// } else if req.Path == "logs" {
	// 	// Create the same data frame as for query data.
	// 	frame := data.NewFrame("response")
	// 	log.DefaultLogger.Info("Logs 1.............")
	// 	// Add fields (matching the same schema used in QueryData).
	// 	frame.Fields = append(frame.Fields,
	// 		data.NewField("time", nil, make([]time.Time, 1)),
	// 		// data.NewField("type", nil, make([]string, 1)),
	// 		// data.NewField("level", nil, make([]string, 1)),
	// 		// data.NewField("module", nil, make([]string, 1)),
	// 		// data.NewField("message", nil, make([]string, 1)),
	// 		// data.NewField("deviceId", nil, make([]string, 1)),
	// 		data.NewField("log", nil, make([]string, 1)),
	// 	)

	// 	u := url.URL{Scheme: "ws", Host: "localhost:9080", Path: "/v1/ws/projects/local-teste/logs"}
	// 	log.DefaultLogger.Info("Logs 2.............")
	// 	log.DefaultLogger.Info("golioth websocket connecting to ", u.String())

	// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// 	// log.DefaultLogger.Info(" golioth websocket dial c", "info", c)
	// 	if err != nil {
	// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	// 	}
	// 	defer c.Close()

	// 	done := make(chan struct{})

	// 	go func() {
	// 		m := make(map[string]interface{})
	// 		defer close(done)
	// 		for {
	// 			_, message, _ := c.ReadMessage()
	// 			// if err != nil {
	// 			// 	log.DefaultLogger.Error("golioth websocket read:", "error", err)
	// 			// 	// return
	// 			// 	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	// 			// 	if err != nil {
	// 			// 		log.DefaultLogger.Error("error golioth websocket dial", "error", err)
	// 			// 	}
	// 			// 	log.DefaultLogger.Info("golioth websocket reconnected")
	// 			// }
	// 			log.DefaultLogger.Info("Leu Logs? %+v", m)
	// 			json.Unmarshal(message, &m)
	// 			logData := m["result"].(map[string]interface{})["data"].(map[string]interface{})
	// 			log.DefaultLogger.Info("Reading Logs %+v", m)
	// 			frame.Fields[0].Set(0, time.Now())
	// 			// frame.Fields[1].Set(0, logData["type"].(string))
	// 			// frame.Fields[2].Set(0, logData["level"].(string))
	// 			// frame.Fields[3].Set(0, logData["module"].(string))
	// 			// frame.Fields[4].Set(0, logData["message"].(string))
	// 			// frame.Fields[5].Set(0, logData["deviceId"].(string))

	// 			l, _ := json.Marshal(logData)
	// 			frame.Fields[1].Set(0, string(l))
	// 			log.DefaultLogger.Info("Sending Time Series")

	// 			err = sender.SendFrame(frame, data.IncludeAll)
	// 			if err != nil {
	// 				log.DefaultLogger.Error("Error sending frame", "error", err)
	// 				continue
	// 			}
	// 		}
	// 	}()

	// }

	return nil
}

// PublishStream is called when a client sends a message to the stream.
func (d *SampleDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
