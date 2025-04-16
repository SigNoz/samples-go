package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/helloworld"
	"github.com/temporalio/samples-go/helloworld/instrument"
)

func main() {
	ctx := context.Background()

	// Create a new Zerolog adapter.
	logger := instrument.NewZerologAdapter()

	tp, mp, err := instrument.InitializeGlobalTelemetryProvider(ctx)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Unable to create a global trace provider: %v", err.Error()))
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Error().Msg(fmt.Sprintf("Error shutting down trace provider: %v", err.Error()))
		}
		if err := mp.Shutdown(ctx); err != nil {
			log.Error().Msg(fmt.Sprintf("Error shutting down meter provider: %v", err.Error()))
		}
	}()

	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Unable to create interceptor: %v", err.Error()))
	}

	// Create metrics handler
	metricsHandler := instrument.NewOpenTelemetryMetricsHandler()

	// The client is a heavyweight object that should be created once per process.
	c, err := helloworld.NewClient(ctx, tracingInterceptor, metricsHandler, logger)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to create client: %v", err.Error()))
	}
	defer c.Close()

	w := worker.New(c, "hello-world", worker.Options{
		// Create interceptor that will put started time on the logger
		Interceptors: []interceptor.WorkerInterceptor{tracingInterceptor},
	})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	// Start listening to the Task Queue.
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to start worker: %v", err.Error()))
	}
}
