package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Metric struct {
	Name      string   `yaml:"name"`
	Help      string   `yaml:"help"`
	Value     string   `yaml:"value"`
	Labels    []string `yaml:"labels"`
	Statement string   `yaml:"statement"`
	Values    map[string]interface{}
}

type Metrics struct {
	Metric []Metric `yaml:"metrics"`
}

type Config struct {
	Connection  string   `yaml:"connection"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	DSN         string   `yaml:"dsn"`
	MetricFiles []string `yaml:"metric_files"`
	Metrics     []Metric
}

type Connections struct {
	Configs []Config `yaml:"configs"`
}

var location = flag.String("metrics-folder",
	os.Getenv("PWD")+"/metrics",
	"Location of folder containg metrics files: Defaults to your path's metric folder")

var config = flag.String("config",
	os.Getenv("PWD")+"/config.yaml",
	"Location of config file: Defaults to your path's config.yml")

func Get_Conns() Connections {
	flag.Parse()
	connections := get_conf_struct(*config)

	for i, config := range connections.Configs {
		connections.Configs[i].Metrics = get_metrics(config.MetricFiles)
	}
	return connections
}

func get_conf_struct(config string) Connections {
	yamlFile, err := os.ReadFile(config)
	if err != nil {
		logrus.Fatal("Error reading YAML file: \n\t", err, "\n")
		os.Exit(1)
	}
	var yamlConfig Connections
	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		logrus.Fatal("Error reading YAML file: \n\t", err, "\n")
		os.Exit(1)
	}
	return yamlConfig
}

func get_metrics(files []string) []Metric {
	var metrics []Metric
	for _, file := range files {
		yamlFile, err := os.ReadFile(*location + "/" + file)
		if err != nil {
			logrus.Fatal("Error reading YAML file: \n\t", err, "\n")
			os.Exit(1)
		}
		var met Metrics
		err = yaml.Unmarshal(yamlFile, &met)
		if err != nil {
			logrus.Fatal("Error reading YAML file: \n\t", err, "\n")
			os.Exit(1)
		}
		metrics = append(metrics, met.Metric...)
	}
	return metrics
}
