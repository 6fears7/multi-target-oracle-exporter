package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type error interface {
	Error() string
}

var sqlServerUp = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "oracle_server_up",
		Help: "If the connection to the server was successful",
	},
	[]string{"error"},
)

func do_stuff(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	configs := get_metric_info(target)
	registry := prometheus.NewRegistry()
	registry.MustRegister(sqlServerUp)
	for _, metric := range configs.Metrics {
		makeGauges(registry, metric)
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
func main() {
	logrus.Info("Starting exporter")
	mainPage()
	probePage()
	logrus.Fatal(http.ListenAndServe(":9101", nil))
}

func makeGauges(reg *prometheus.Registry, metric Metric) {
	var label_vals []string
	any_errs := false

	newValues := make(map[string]interface{})
	for k, v := range metric.Values {
		newValues[strings.ToLower(k)] = v
	}

	// Update the original map with the new map
	metric.Values = newValues

	for _, labelKey := range metric.Labels {
		value, ok := metric.Values[labelKey]
		if !ok {
			log.Printf("Metric data collection failed. Key not found in metric values: %s\n", labelKey)
			any_errs = true
			continue // Use continue instead of break to skip the current iteration
		}

		// Append the value to label_vals
		label_vals = append(label_vals, value.(string))
	}
	if !any_errs {
		x := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: metric.Name,
			Help: metric.Help,
		}, metric.Labels)
		x, err := set_values(x, label_vals, metric)
		if err == nil {
			reg.MustRegister(x)
		}
	}
}
func set_values(x *prometheus.GaugeVec, label_vals []string, metric Metric) (*prometheus.GaugeVec, error) {
	switch metric.Values[metric.Value].(type) {
	case float64:
		x.WithLabelValues(label_vals...).Set(float64(metric.Values[metric.Value].(float64)))
	case int64:
		x.WithLabelValues(label_vals...).Set(float64(metric.Values[metric.Value].(int64)))
	case string:
		float, err := strconv.ParseFloat(metric.Values[metric.Value].(string), 64)
		if err != nil {
			logrus.Error("Could not convert string to float64: ", err)
			break
		}
		x.WithLabelValues(label_vals...).Set(float)
	case nil:
		return nil, errors.New("Metric data collection failed. Unable to add metric: " + metric.Name)
	}
	return x, nil
}
func mainPage() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		data := targetList()
		sort.Strings(data)
		const tmpl = `<html>
        <head>
        <title>Multi-Target Oracle Exporter</title>
        <style>
        label{
        display:inline-block;
        width:75px;
        }
        form label {
        margin: 10px;
        }
        form input {
        margin: 10px;
        }
        </style>
        </head>
        <body>
        <h1>Multi-Target Oracle Exporter</h1>
        <form action="/probe">
        <label>Targets:</label>
		{{ range $i := .}}
		<p><a href="/probe?target={{$i}}">{{$i}}<br/>
		{{end}}</a></p>
        </form>
        </body>
        </html>`

		t, err := template.New("webpage").Parse(tmpl)
		if err != nil {
			log.Fatal(err)
		}
		t.Execute(w, data)
		if err != nil {
			logrus.Fatal(err)
		}

	})
}

func probePage() {
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		do_stuff(w, r)
	})
}

func targetList() []string {
	result := Get_Conns()
	var targets []string
	for _, val := range result.Configs {
		targets = append(targets, val.Connection)
	}
	return targets
}
