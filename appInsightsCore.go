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

func (ins *AppInsightsCore) TraceDependencyWithIds(
	tid string,
	rid string,
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
	tele.Tags.Operation().SetId(tid)
	tele.Tags.Operation().SetParentId(rid)
	tele.Tags.Operation().SetName(commandName)
	tele.Tags.Cloud().SetRole(ins.ServName)
	ins.Client.Track(tele)
}
