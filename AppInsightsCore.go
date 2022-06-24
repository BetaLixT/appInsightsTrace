package gotrace

import (
	"time"

	"github.com/BetaLixT/appInsightsTrace/optn"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"

	"go.uber.org/zap"
)

type AppInsightsCore struct {
	Client   *appinsights.TelemetryClient
	ServName string
}

func (insights *AppInsightsCore) Close() {
	select {
	case <-(*insights.Client).Channel().Close(10 * time.Second):
	case <-time.After(30 * time.Second):
	}
}

func NewAppInsightsCore(
	optn *optn.AppInsightsOptions,
	lgr *zap.Logger,
) *AppInsightsCore {
	client := appinsights.NewTelemetryClient(optn.InstrumentationKey)
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		lgr.Info(msg)
		return nil
	})
	return &AppInsightsCore{
		Client:   &client,
		ServName: optn.ServiceName,
	}
}
