package optn

import "github.com/spf13/viper"

type DatabaseOptions struct {
  ConnectionString string
}

func NewDatabaseOptions(cfg *viper.Viper) *DatabaseOptions {
  return &DatabaseOptions{
    ConnectionString: cfg.GetString("DatabaseOptions.ConnectionString"),
  }
}
