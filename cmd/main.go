package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/alexeynavarkin/docker_exporter/internal/event"
	"github.com/alexeynavarkin/docker_exporter/internal/metric"
	"github.com/alexeynavarkin/docker_exporter/internal/stat"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	metricCollector := metric.NewCollector()

	metricCollector.Register(stat.NewGatherer(cli))

	evHandler := event.NewHandler(cli, metricCollector)

	go evHandler.HandleEvents(ctx)
	go metricCollector.ExposeHTTP(ctx)

	<-ctx.Done()
}
