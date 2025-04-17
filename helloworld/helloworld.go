package helloworld

import (
	"context"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, "Activity").Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	logger.Info("Workflow completed", "result", result)
	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity started", "name", name)

	url := "https://signoz.io"
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	logger.Info("Sending request to Signoz", "url", url)
	res, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request", "error", err)
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return "", err
	}

	logger.Info("Received response from Signoz", "bytes", len(body))
	return "Hello " + name + "!", nil
}
