FROM golang:1.18-alpine3.15 as build

RUN mkdir -p /app
COPY go.mod go.sum /app/
COPY multi-target-oracle-exporter/*.go /app/multi-target-oracle-exporter/
WORKDIR /app/multi-target-oracle-exporter
RUN go build -o multi-target-oracle-exporter
RUN chmod a+x ./multi-target-oracle-exporter

FROM alpine
COPY --from=build /app/multi-target-oracle-exporter/multi-target-oracle-exporter /app/multi-target-oracle-exporter
WORKDIR /app/
EXPOSE 9101
ENTRYPOINT [ "/app/multi-target-oracle-exporter"]