package main

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	FromDir             string   `yaml:"fromDir"`
	ToDir               string   `yaml:"toDir"`
	TargetExts          string   `yaml:"targetExts"`
	TargetDocumentsExts []string `yaml:"targetDocumentsExts"`
	TargetPicturesExts  []string `yaml:"targetPicturesExts"`
	TargetMusicsExts    []string `yaml:"targetMusicsExts"`
	TargetMoviesExts    []string `yaml:"targetMoviesExts"`
	Rename              bool     `yaml:"rename"`
	Operation           int      `yaml:"operation"`
}

func getConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
