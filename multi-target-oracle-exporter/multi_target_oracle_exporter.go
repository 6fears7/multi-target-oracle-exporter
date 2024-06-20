package main

import (
	"database/sql"
	"errors"
	"strings"
	"sync"

	go_ora "github.com/sijms/go-ora/v2"
	"github.com/sirupsen/logrus"
)

var waitGroup sync.WaitGroup

func sql_exporter(name string) Config {

	conf, err := get_config(name, Get_Conns().Configs)
	if err != nil {
		logrus.Error(err)

	}
	return conf
}

func get_config(name string, configs []Config) (Config, error) {

	for _, config := range configs {
		if config.Connection == name {
			return config, nil
		}
	}
	var empty_conf Config
	return empty_conf, errors.New("Could not find the configuration file with name: " + name)

}
func get_metric_info(server_con string) Config {
	conf := sql_exporter(server_con)
	db, err := connect(conf)
	if err == nil {
		waitGroup.Add(len(conf.Metrics))
		for i, metric := range conf.Metrics {
			go func(i int, metric Metric) {
				values, err := run_query(db, metric.Statement)
				if err != nil {
					logrus.Error("Error collecting metrics:\n\t", err)

				} else {
					conf.Metrics[i].Values = values
				}
				defer waitGroup.Done()
			}(i, metric)
		}
		waitGroup.Wait()
	} else {
		conf.Metrics = make([]Metric, 0)
	}
	db.Close()
	return conf
}

func connect(con Config) (*sql.DB, error) {

	originalDSN := con.DSN
	multiHostDSN := strings.Split(originalDSN, "(DESCRIPTION")
	var connectionArray []string

	for i, part := range multiHostDSN {
		if i > 0 {
			connectionArray = append(connectionArray, "(DESCRIPTION"+part)
		} else if part != "" {
			connectionArray = append(connectionArray, part)
		}
	}

	connect := go_ora.BuildJDBC(con.Username, con.Password, connectionArray[0], nil)

	db, _ := sql.Open("oracle", connect)
	sqlServerUp.Reset()
	Err := db.Ping()
	if Err != nil {

		if len(connectionArray) > 1 {
			logrus.Error("Connection failed, trying again with another connection string found")

			for i := range connectionArray {

				connect := go_ora.BuildJDBC(con.Username, con.Password, connectionArray[i], nil)

				db, _ := sql.Open("oracle", connect)
				sqlServerUp.Reset()
				Err := db.Ping()

				if Err != nil {
					sqlServerUp.WithLabelValues(Err.Error()).Set(float64(0))

				}
			}
		} else {

			sqlServerUp.WithLabelValues(Err.Error()).Set(float64(0))
		}
	} else {
		sqlServerUp.WithLabelValues("").Set(float64(1))
	}

	return db, Err
}

func run_query(db *sql.DB, statement string) (map[string]interface{}, error) {
	value := make(map[string]interface{})

	rows, err := db.Query(statement)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			logrus.Error(err)
		}

		for i, col := range columns {
			val := values[i]

			b, ok := val.([]byte)
			var v interface{}
			if ok {
				v = string(b)
			} else {
				v = val
			}

			value[col] = v

		}
	}
	return value, nil
}
