package metric

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Collector struct {
	event *prometheus.CounterVec

	reg                *prometheus.Registry
	defaultHandlerFunc http.Handler
}

func NewCollector() *Collector {
	reg := prometheus.NewRegistry()

	c := &Collector{
		event: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "docker",
				Name:      "event",
				Help:      "Docker events.",
			},
			[]string{"containerName", "serviceName", "serviceID", "eventType"},
		),
		reg:                reg,
		defaultHandlerFunc: promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}),
	}

	reg.MustRegister(c.event)

	return c
}

func (c *Collector) Register(collector prometheus.Collector) {
	c.reg.MustRegister(collector)
}

func (c *Collector) RegisterEvent(containerName, serviceName, serviceID, eventType string) {
	c.event.With(
		prometheus.Labels{
			"containerName": containerName,
			"serviceName":   serviceName,
			"serviceID":     serviceID,
			"eventType":     eventType,
		},
	).Inc()
}

func (c *Collector) ExposeHTTP(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", c.defaultHandlerFunc)

	s := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		err := s.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error listen %s\n", err)
		}
	}()

	<-ctx.Done()
	err := s.Close()
	if err != nil {
		fmt.Printf("error close server %s\n", err)
	}
}
