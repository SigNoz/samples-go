package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/client"

	"github.com/rs/zerolog/log"
	"github.com/temporalio/samples-go/helloworld"
	"github.com/temporalio/samples-go/helloworld/instrument"
)

func main() {
	logger := instrument.NewZerologAdapter()

	ctx := context.Background()
	tp, mp, err := instrument.InitializeGlobalTelemetryProvider(ctx)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to create a global trace provider: %v", err.Error()))
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle shutdown
	go func() {
		<-sigChan
		if err := tp.Shutdown(ctx); err != nil {
			log.Error().Msg(fmt.Sprintf("Error shutting down trace provider: %v", err.Error()))
		}
		if err := mp.Shutdown(ctx); err != nil {
			log.Error().Msg(fmt.Sprintf("Error shutting down meter provider: %v", err.Error()))
		}
		os.Exit(0)
	}()

	tracingInterceptor, err := instrument.NewTracingInterceptor(instrument.TracerOptions{})
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to create interceptor: %v", err.Error()))
	}

	metricsHandler := instrument.NewOpenTelemetryMetricsHandler()

	// The client is a heavyweight object that should be created once per process.
	c, err := helloworld.NewClient(ctx, tracingInterceptor, metricsHandler, logger)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to create client: %v", err.Error()))
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "hello-world",
	}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, helloworld.Workflow, "Workflow Name 2")
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to execute workflow: %v", err.Error()))
	}

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("Unable to get workflow result: %v", err.Error()))
	}
	log.Info().Msg(fmt.Sprintf("Workflow result: %v", result))

}
