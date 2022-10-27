package appinsightstrace

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"

	"go.uber.org/zap"
)

type AppInsightsCore struct {
	Client         appinsights.TelemetryClient
	traceExtractor ITraceExtractor
	ServName       string
}

// Constructs an instance of AppInsightsCore with defaults, including the default
// ITraceExtractor which does not provide any tracing information from the context
// it is recommended to use the non context dependent functions (functions that end
// with "WithIds") to take advangate of the tracing if you use this constructor
func NewBasic(
	instrumentationKey string,
	serviceName string,
) (*AppInsightsCore, error) {
	lgr, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	client := appinsights.NewTelemetryClient(instrumentationKey)
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		lgr.Info(msg)
		return nil
	})
	return &AppInsightsCore{
		Client:         client,
		ServName:       serviceName,
		traceExtractor: &DefaultTraceExtractor{},
	}, nil
}

// Constructs an instance of AppInsightsCore using the provided zap logger and 
// the default ITraceExtractor which does not provide any tracing information
// from the context it is recommended to use the non context dependent functions
// (functions that end with "WithIds") to take advangate of the tracing if you
// use this constructor
func NewBasicWithLogger(
	instrumentationKey string,
	serviceName string,
	lgr zap.Logger,
) (*AppInsightsCore, error) {	
	client := appinsights.NewTelemetryClient(instrumentationKey)
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		lgr.Info(msg)
		return nil
	})
	return &AppInsightsCore{
		Client:         client,
		ServName:       serviceName,
		traceExtractor: &DefaultTraceExtractor{},
	}, nil
}

// Constructs an instance of AppInsightsCore using the provided zap logger and 
// a custom trace extractor, it's recommended to provide a custom trace
// extractor that will extract the w3c trace information from the context and
// take advantage of the context dependent trace functions, check documentation
// of ITraceExtractor for more information
func NewAppInsightsCore(
	optn *AppInsightsOptions,
	traceExtractor ITraceExtractor,
	lgr *zap.Logger,
) *AppInsightsCore {
	client := appinsights.NewTelemetryClient(optn.InstrumentationKey)
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		lgr.Info(msg)
		return nil
	})
	return &AppInsightsCore{
		Client:         client,
		ServName:       optn.ServiceName,
		traceExtractor: traceExtractor,
	}
}

// Same as NewAppInsightsCorenbut without options, instead just taking values
// as strings
func NewAppInsightsCoreFlatOptions(
	instrumentationKey string,
	serviceName string,
	traceExtractor ITraceExtractor,
	lgr *zap.Logger,
) *AppInsightsCore {
	client := appinsights.NewTelemetryClient(instrumentationKey)
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		lgr.Info(msg)
		return nil
	})
	return &AppInsightsCore{
		Client:         client,
		ServName:       serviceName,
		traceExtractor: traceExtractor,
	}
}

func (insights *AppInsightsCore) Close() {
	select {
	case <-insights.Client.Channel().Close(10 * time.Second):
	case <-time.After(30 * time.Second):
	}
}

func (ins *AppInsightsCore) ExtractTraceInfo(
	ctx context.Context,
) (ver, tid, pid, rid, flg string) {
	return ins.traceExtractor.ExtractTraceInfo(ctx)
}

// - Context dependent


// Transmits a new Request telemtery, this should be used to trace incoming
// requests (from a middleware for example).
//
// ctx: the current context of the execution, the ITraceExtractor.ExtractTraceInfo
//   function will be utilized to extract traceId parentId and current requestId
//   from the context the default implementation of ITraceExtractor provided in
//   this package will leave these fields empty
// method: the http request method (https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
// path: the path for the request (ex: /api/v1/weather)
// query: any query paramters provided with the request
// statusCode: the response http status code (https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
// bodySize: the size of the response body 
// ip: ip of the client making the request
// userAgent: the user agent of the client making the request (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent)
// startTimestamp: timestamp of when the request was received by this service
// eventTimestamp: timestamp of when the request has completed processing by this
//   service
func (ins *AppInsightsCore) TraceRequest(
	ctx context.Context,
	method string,
	path string,
	query string,
	statusCode int,
	bodySize int,
	ip string,
	userAgent string,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	_, tid, pid, rid, _ := ins.traceExtractor.ExtractTraceInfo(ctx)

	props := fields
	props["bodySize"] = strconv.Itoa(bodySize)
	props["ip"] = ip
	props["userAgent"] = userAgent
	name := fmt.Sprintf("%s %s", method, path)
	tele := appinsights.RequestTelemetry{
		Name:         name,
		Url:          fmt.Sprintf("%s%s", path, query),
		Id:           rid,
		Duration:     eventTimestamp.Sub(startTimestamp),
		ResponseCode: strconv.Itoa(statusCode),
		Success:      statusCode > 99 && statusCode < 300,
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  startTimestamp,
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}

	tele.Tags.Cloud().SetRole(ins.ServName)
	tele.Tags.Operation().SetId(tid)
	tele.Tags.Operation().SetParentId(pid)
	tele.Tags.Operation().SetName(name)

	ins.Client.Track(&tele)
}



// Transmits a new Request telemtery for events, this should be used to trace incoming
// events (from a middleware for example).
//
// ctx: the current context of the execution, the ITraceExtractor.ExtractTraceInfo
//   function will be utilized to extract traceId parentId and current requestId
//   from the context the default implementation of ITraceExtractor provided in
//   this package will leave these fields empty
// name: the custom name of the trace
// key: the routing key of the event
// statusCode: the response http status code (https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
// startTimestamp: timestamp of when the event was received by this service
// eventTimestamp: timestamp of when the event has completed processing by this
//   service
func (ins *AppInsightsCore) TraceEvent(
	ctx context.Context,
	name string,
	key string,
	statusCode int,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	_, tid, pid, rid, _ := ins.traceExtractor.ExtractTraceInfo(ctx)

	props := fields
	name := fmt.Sprintf("%s %s", name, key)
	tele := appinsights.RequestTelemetry{
		Name:         name,
		Url:          key,
		Id:           rid,
		Duration:     eventTimestamp.Sub(startTimestamp),
		ResponseCode: strconv.Itoa(statusCode),
		Success:      statusCode > 99 && statusCode < 300,
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  startTimestamp,
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}

	tele.Tags.Cloud().SetRole(ins.ServName)
	tele.Tags.Operation().SetId(tid)
	tele.Tags.Operation().SetParentId(pid)
	tele.Tags.Operation().SetName(name)

	ins.Client.Track(&tele)
}



// Transmits a new Dependency telemtery, this should be used to trace outgoing
// requests, database calls etc.
//
// ctx: the current context of the execution, the ITraceExtractor.ExtractTraceInfo
//   function will be utilized to extract traceId parentId and current requestId
//   from the context the default implementation of ITraceExtractor provided in
//   this package will leave these fields empty
// spanId: optional id that can be provided, recommended to leave empty for most
//   cases except if the dependency also follows w3c tracing (outgoing http calls)
//   just helps in forming an accurate trace tree
// dependencyType: the type of the dependency, for example postgres, rabbitmq etc
// serviceName: unique name for the dependency (database address for example)
// commandName: name for the action being performed by the dependency, for example
//   the method and path for an outgoing http request (GET /api/v1/weather)
// success: whether the requet was successful or not
// startTimestamp: timestamp of when the dependency has been invoked by this
//   service
// eventTimestamp: timestamp of when the dependency has been completed
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceDependency(
	ctx context.Context,
	spanId string,
	dependencyType string,
	serviceName string,
	commandName string,
	success bool,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	_, tid, _, rid, _ := ins.traceExtractor.ExtractTraceInfo(ctx)

	props := fields
	tele := &appinsights.RemoteDependencyTelemetry{
		Id:       spanId,
		Name:     commandName,
		Type:     dependencyType,
		Target:   serviceName,
		Success:  success,
		Duration: eventTimestamp.Sub(startTimestamp),
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  startTimestamp,
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}
	tele.Tags.Operation().SetId(tid)
	tele.Tags.Operation().SetParentId(rid)
	tele.Tags.Operation().SetName(commandName)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}


// Transmits a new trace log telemetry.
//
// ctx: the current context of the execution, the ITraceExtractor.ExtractTraceInfo
//   function will be utilized to extract traceId parentId and current requestId
//   from the context the default implementation of ITraceExtractor provided in
//   this package will leave these fields empty
// message: the message that is to be traced
// severityLevel: the severity level (Critical, Error, Warning, Information,
//   Verbose), it's recommended to use the constants of the same name provided
//   in this package
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceLog(
	ctx context.Context,
	message string,
	severityLevel SeverityLevel,
	fields map[string]string,
) {
	_, tid, _, rid, _ := ins.traceExtractor.ExtractTraceInfo(ctx)

	props := fields
	tele := &appinsights.TraceTelemetry{
		Message: message,
		SeverityLevel: contracts.SeverityLevel(severityLevel),
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  time.Now(),
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},	
	}
	tele.Tags.Operation().SetId(tid)
	tele.Tags.Operation().SetParentId(rid)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}


// Transmits a new exception telemetry, should be used to track unexpected errors
// or panics.
//
// ctx: the current context of the execution, the ITraceExtractor.ExtractTraceInfo
//   function will be utilized to extract traceId parentId and current requestId
//   from the context the default implementation of ITraceExtractor provided in
//   this package will leave these fields empty
// err: the unexpected error object
// skip: the number of levels to skip on the call stack (set to 0 if unsure)
// severityLevel: the severity level (Critical, Error, Warning, Information,
//   Verbose), it's recommended to use the constants of the same name provided
//   in this package
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceException(
	ctx context.Context,
	err interface{},
	skip int,
	fields map[string]string,
) {
	_, tid, _, rid, _ := ins.traceExtractor.ExtractTraceInfo(ctx)

	props := fields
	tele := &appinsights.ExceptionTelemetry{
		Error:         err,
		Frames:        appinsights.GetCallstack(2 + skip),
		SeverityLevel: appinsights.Error,
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  time.Now(),
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}
	tele.Tags.Operation().SetId(tid)
	tele.Tags.Operation().SetParentId(rid)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}

// - Context Independent


// Transmits a new Request telemtery, this should be used to trace incoming
// requests (from a middleware for example). it uses the provied traceId, parentId
// and requestId instead of trying to extract the same from the context.
// The said IDs follow the W3C trace-context spec
// (https://www.w3.org/TR/trace-context/).
//
// traceId: is a common identifier for all dependents and dependencies for the
//   request, can be left as an empty string but this would reduce tracablility
// parentId is and id unique to the current request's direct dependent, can be left
//   as an empty string but this would reduce tracablility
// requestId is the unique Id for the request, generated at the start of the
//   request, can be left as an empty string but this would reduce tracablility
// method: the http request method (https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
// path: the path for the request (ex: /api/v1/weather)
// query: any query paramters provided with the request
// statusCode: the response http status code (https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
// bodySize: the size of the response body 
// ip: ip of the client making the request
// userAgent: the user agent of the client making the request (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent)
// startTimestamp: timestamp of when the request was received by this service
// eventTimestamp: timestamp of when the request has completed processing by this
//   service
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceRequestWithIds(
	traceId string,
	parentId string,
	requestId string,
	method string,
	path string,
	query string,
	statusCode int,
	bodySize int,
	ip string,
	userAgent string,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {

	props := fields
	props["bodySize"] = strconv.Itoa(bodySize)
	props["ip"] = ip
	props["userAgent"] = userAgent
	name := fmt.Sprintf("%s %s", method, path)
	tele := appinsights.RequestTelemetry{
		Name:         name,
		Url:          fmt.Sprintf("%s%s", path, query),
		Id:           requestId,
		Duration:     eventTimestamp.Sub(startTimestamp),
		ResponseCode: strconv.Itoa(statusCode),
		Success:      statusCode > 99 && statusCode < 300,
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  startTimestamp,
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}

	tele.Tags.Cloud().SetRole(ins.ServName)
	tele.Tags.Operation().SetId(traceId)
	tele.Tags.Operation().SetParentId(parentId)
	tele.Tags.Operation().SetName(name)

	ins.Client.Track(&tele)
}


// Transmits a new Dependency telemtery, this should be used to trace outgoing
// requests, database calls etc. it uses the provied traceId and requestId instead
// of trying to extract the same from the context.
//
// traceId: is a common identifier for all dependents and dependencies for the
//   request, can be left as an empty string but this would reduce tracablility
// requestId is the unique Id for the request, generated at the start of the
//   request, can be left as an empty string but this would reduce tracablility
// spanId: optional id that can be provided, recommended to leave empty for most
//   cases except if the dependency also follows w3c tracing (outgoing http calls)
//   just helps in forming an accurate trace tree
// dependencyType: the type of the dependency, for example postgres, rabbitmq etc
// serviceName: unique name for the dependency (database address for example)
// commandName: name for the action being performed by the dependency, for example
//   the method and path for an outgoing http request (GET /api/v1/weather)
// success: whether the requet was successful or not
// startTimestamp: timestamp of when the dependency has been invoked by this
//   service
// eventTimestamp: timestamp of when the dependency has been completed
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceDependencyWithIds(
	traceId string,
	requestId string,
	spanId string,
	dependencyType string,
	serviceName string,
	commandName string,
	success bool,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {

	props := fields
	tele := &appinsights.RemoteDependencyTelemetry{
		Id:       spanId,
		Name:     commandName,
		Type:     dependencyType,
		Target:   serviceName,
		Success:  success,
		Duration: eventTimestamp.Sub(startTimestamp),
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  startTimestamp,
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}
	tele.Tags.Operation().SetId(traceId)
	tele.Tags.Operation().SetParentId(requestId)
	tele.Tags.Operation().SetName(commandName)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}


// Transmits a new trace log telemetry. It uses the provied traceId and requestId
// instead of trying to extract the same from the context.
//
// traceId: is a common identifier for all dependents and dependencies for the
//   request, can be left as an empty string but this would reduce tracablility
// requestId is the unique Id for the request, generated at the start of the
//   request, can be left as an empty string but this would reduce tracablility
// message: the message that is to be traced
// severityLevel: the severity level (Critical, Error, Warning, Information,
//   Verbose), it's recommended to use the constants of the same name provided
//   in this package
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceLogWithIds(
	traceId string,
	requestId string,
	message string,
	severityLevel contracts.SeverityLevel,
	timestamp time.Time,
	fields map[string]string,
) {

	props := fields
	tele := &appinsights.TraceTelemetry{
		Message: message,
		SeverityLevel: contracts.SeverityLevel(severityLevel),
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  timestamp,
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},	
	}
	tele.Tags.Operation().SetId(traceId)
	tele.Tags.Operation().SetParentId(requestId)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}


// Transmits a new exception telemetry, should be used to track unexpected errors
// or panics. It uses the provied traceId and requestId instead of trying to
// extract the same from the context.
//
// traceId: is a common identifier for all dependents and dependencies for the
//   request, can be left as an empty string but this would reduce tracablility
// requestId is the unique Id for the request, generated at the start of the
//   request, can be left as an empty string but this would reduce tracablility
// err: the unexpected error object
// skip: the number of levels to skip on the call stack (set to 0 if unsure)
// severityLevel: the severity level (Critical, Error, Warning, Information,
//   Verbose), it's recommended to use the constants of the same name provided
//   in this package
// fields: additional custom values to include in the telemetry
func (ins *AppInsightsCore) TraceExceptionWithIds(
  traceId string,
  requestId string,
	err interface{},
	skip int,
	fields map[string]string,
) {

	props := fields
	tele := &appinsights.ExceptionTelemetry{
		Error:         err,
		Frames:        appinsights.GetCallstack(2 + skip),
		SeverityLevel: appinsights.Error,
		BaseTelemetry: appinsights.BaseTelemetry{
			Timestamp:  time.Now(),
			Tags:       make(contracts.ContextTags),
			Properties: props,
		},
		BaseTelemetryMeasurements: appinsights.BaseTelemetryMeasurements{
			Measurements: make(map[string]float64),
		},
	}
	tele.Tags.Operation().SetId(traceId)
	tele.Tags.Operation().SetParentId(requestId)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}
