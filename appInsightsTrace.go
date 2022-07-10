package appinsightstrace

import (
	"fmt"
	"strconv"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"
	//"github.com/BetaLixT/gottp"
)

type AppInsightsTrace struct {
	core *AppInsightsCore
	tid  string
	pid  string
	rid  string
}

//for testing purposes
//var _ interfaces.ITracer = (*AppInsightsTrace)(nil)

func NewAppInsightsTrace(core *AppInsightsCore, tid string, pid string, rid string) *AppInsightsTrace {
	if pid == "" {
		pid = tid
	}
	return &AppInsightsTrace{
		core: core,
		tid:  tid,
		pid:  pid,
		rid:  rid,
	}
}

func (tracer *AppInsightsTrace) TraceRequest(
	isParent bool,
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
	tele := appinsights.RequestTelemetry{
		Name:         fmt.Sprintf("%s %s", method, path),
		Url:          fmt.Sprintf("%s%s", path, query),
		Id:           tracer.rid,
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

	tele.Tags.Operation().SetId(tracer.tid)
	tele.Tags.Operation().SetParentId(tracer.pid)

	(*tracer.core.Client).Track(&tele)
}

func (tracer *AppInsightsTrace) TraceDependency(
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
	// props["bodySize"] = strconv.Itoa(bodySize)
	// props["ip"] = ip
	// props["userAgent"] = userAgent

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
	tele.Tags.Operation().SetId(tracer.tid)
	tele.Tags.Operation().SetParentId(tracer.rid)
	(*tracer.core.Client).Track(tele)
}
