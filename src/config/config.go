package config

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init(conf string) {
	//TODO implement viper.SetEnvPrefix()
	viper.AutomaticEnv()
	viper.SetConfigName(conf)
	viper.AddConfigPath("$HOME/.skywire-uptime-system/")
	viper.AddConfigPath(".")
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed %s", e.Name)
	})
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading configuration %v", err)
	}
	log.Debug("Configuration initialized")
}
