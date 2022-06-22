package optn

import (
	"github.com/spf13/viper"
)

type CorsOptions struct {
	AllowedOrigins []string
}

func NewCorsOptions(cfg *viper.Viper) *CorsOptions {
	// origins := []string{}
	// iter := 0
	// for true {
	// 	val := cfg.GetString(fmt.Sprintf(
	// 		"%s.%s.%d",
	// 		"CorsOptions",
	// 		"AllowedOrigins",
	// 		iter,
	// 	))
	// 	if val == "" {
	// 		break
	// 	} else {
	// 		origins = append(origins, val)
	// 		iter++
	// 	}
	// }
	// return &CorsOptions{
	// 	AllowedOrigins: origins,
	// }
	return &CorsOptions{
	  AllowedOrigins: cfg.GetStringSlice("CorsOptions.AllowedOrigins"),
	}
}
