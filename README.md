
# Multi-Target Oracle Exporter
## Notes:
An Oracle SQL database exporter written in Go designed for use in Kubernetes

Currently only gauge metrics are used.

## How it works
Utilizes the JDBC connection method in Go for enabling users to pass in TNS Connection details to Oracle servers.

## Port
By default, runs on port :9101

## Pre-Reqs
The example kube files assume you have the following:
- A secret called "oracle-config"
- A configmap called "oracle-configmap"

### Config file

The config file for a given connection requires you to pass in the details of the TNS connection. 
```
(DESCRIPTION=(CONNECT_TIMEOUT=5)(TRANSPORT_CONNECT_TIMEOUT=3)(RETRY_COUNT=3)(ADDRESS_LIST=(LOAD_BALANCE=on)(ADDRESS=(PROTOCOL=TCP)(HOST=myListener)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=myService)))
```

If you have multiple connections for handling failover or general resiliency, ensure you do not have "DESCRIPTION_LIST" in the "dsn" key and instead just add multiple connection details like so:
```   
(DESCRIPTION=(CONNECT_TIMEOUT=5)(TRANSPORT_CONNECT_TIMEOUT=3)(RETRY_COUNT=3)(ADDRESS_LIST=(LOAD_BALANCE=on)(ADDRESS=(PROTOCOL=TCP)(HOST=myListener)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=myService)))(DESCRIPTION=(CONNECT_TIMEOUT=5)(TRANSPORT_CONNECT_TIMEOUT=3)(RETRY_COUNT=3)(ADDRESS_LIST=(LOAD_BALANCE=on)(ADDRESS=(PROTOCOL=TCP)(HOST=myListener2)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=myService2))) 
```


Example config:

```yaml
configs:
  - connection: Prod
    dsn: | 
      (DESCRIPTION=(CONNECT_TIMEOUT=5)(TRANSPORT_CONNECT_TIMEOUT=3)(RETRY_COUNT=3)(ADDRESS_LIST=(LOAD_BALANCE=on)(ADDRESS=(PROTOCOL=TCP)(HOST=myListener)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=myService)))(DESCRIPTION=(CONNECT_TIMEOUT=5)(TRANSPORT_CONNECT_TIMEOUT=3)(RETRY_COUNT=3)(ADDRESS_LIST=(LOAD_BALANCE=on)(ADDRESS=(PROTOCOL=TCP)(HOST=myListener2)(PORT=1521)))(CONNECT_DATA=(SERVICE_NAME=myService2)))
    username: userName
    password: convertThisFileToAKubeSecret
    metric_files:
    - default_metrics.yaml
    - custom-metrics-MyProdDB.yaml
```
In the event the first connection fails, it will attempt to retry with any other connections found.

You can follow the pattern in the [config.yaml](./kube/oracle_secret.yaml). Its location can be specified with the ```--config``` flag or it will default to ```$PWD/config.yaml```. The connection name can be anything you would like to be passed in as the target paramater in the web request (i.e. ```localhost:9101/probe?target=connection_name```). The metric files will be searched for in the ```metrics-folder```. It defaults to ```$PWD/metrics```. Port is not required if your connection uses the default port.


### Metric files

If you will need to specify which database you need to connect to as part of the statement. The labels will be collected from the column of the same name. You will also want to specify the name of the column used to collect the value. ```value: value_column```. If the metric is incorrectly made, the logs should provide more info. All other metrics should continue to work as expected, but I do not claim to have made a safety catch for every crazy combination, and it may cause an http panic.

Example Metric:

```yaml
metrics:
- name: sessions
  help: Gauge metric with count of sessions by status and type. (value)
  value: value
  labels:
  - status
  - type
  statement: SELECT status, type, COUNT(*) as value FROM gv$session GROUP BY status,
    type
```


### Build ConfigMap

```bash
kubectl create configmap oracle-configmap --from-file metrics/ --namespace prometheus
```

### How to use

```go
go build multi-target-oracle-exporter

multi-target-oracle-exporter --config=/path/to/config.yaml --metrics-folder=/path/to/metrics

```
Or you can use the docker container by mounting the volume
```docker
docker run -v $PWD/kube/oracle_secret.yaml:/config.yaml -v $PWD/metrics:/metrics -p 9101:9101 6fears7/multi-target-oracle-exporter:0.1.0
````

Access at [http://localhost:9101](localhost:9101)
