package optn

import "github.com/spf13/viper"

type FileServiceClientOptions struct {
  BaseUrl string
}

func NewFileServiceClientOptions(cfg *viper.Viper) *FileServiceClientOptions {
  return &FileServiceClientOptions{
    BaseUrl: cfg.GetString("FileServiceClientOptions.BaseUrl"),
  }
}
