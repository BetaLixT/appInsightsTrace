package optn

import "github.com/spf13/viper"

type AppInsightsOptions struct {
	InstrumentationKey string
	ServiceName        string
}

func NewAppInsightsOptions(cfg *viper.Viper) *AppInsightsOptions {
	return &AppInsightsOptions{
		InstrumentationKey: cfg.GetString("AppInsightsOptions.InstrumentationKey"),
		ServiceName:        cfg.GetString("AppInsightsOptions.ServiceName"),
	}
}
