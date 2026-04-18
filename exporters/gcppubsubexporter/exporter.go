// Package gcppubsubexporter implements an OpenTelemetry Collector exporter
// for Google Cloud Pub/Sub. Supports traces, metrics, and logs pipelines.
package gcppubsubexporter

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type gcpPubSubExporter struct {
	config   *Config
	client   *pubsub.Client
	topic    *pubsub.Topic
	logger   *zap.Logger
	settings exporter.CreateSettings
}

type Config struct {
	ProjectID   string `mapstructure:"project"`
	TopicID     string `mapstructure:"topic"`
	Compression string `mapstructure:"compression"`
	Encoding    string `mapstructure:"encoding"`
}

func newGCPPubSubExporter(cfg *Config, set exporter.CreateSettings) *gcpPubSubExporter {
	return &gcpPubSubExporter{
		config:   cfg,
		logger:   set.Logger,
		settings: set,
	}
}

func (e *gcpPubSubExporter) Start(ctx context.Context, _ component.Host) error {
	client, err := pubsub.NewClient(ctx, e.config.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to create GCP Pub/Sub client: %w", err)
	}
	e.client = client
	e.topic = client.Topic(e.config.TopicID)
	e.topic.PublishSettings.CountThreshold = 1000
	e.topic.PublishSettings.DelayThreshold = 10e9

	e.logger.Info("GCP Pub/Sub exporter started",
		zap.String("project", e.config.ProjectID),
		zap.String("topic", e.config.TopicID),
	)
	return nil
}

func (e *gcpPubSubExporter) Shutdown(_ context.Context) error {
	if e.topic != nil {
		e.topic.Stop()
	}
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

func (e *gcpPubSubExporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	payload, err := json.Marshal(map[string]interface{}{
		"type":       "traces",
		"span_count": td.SpanCount(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal traces: %w", err)
	}
	result := e.topic.Publish(ctx, &pubsub.Message{
		Data:       payload,
		Attributes: map[string]string{"signal_type": "traces"},
	})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("failed to publish traces: %w", err)
	}
	e.logger.Debug("Published traces", zap.Int("span_count", td.SpanCount()))
	return nil
}

func (e *gcpPubSubExporter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	payload, err := json.Marshal(map[string]interface{}{
		"type":         "metrics",
		"metric_count": md.MetricCount(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	result := e.topic.Publish(ctx, &pubsub.Message{
		Data:       payload,
		Attributes: map[string]string{"signal_type": "metrics"},
	})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("failed to publish metrics: %w", err)
	}
	return nil
}

func (e *gcpPubSubExporter) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "logs",
		"log_count": ld.LogRecordCount(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}
	result := e.topic.Publish(ctx, &pubsub.Message{
		Data:       payload,
		Attributes: map[string]string{"signal_type": "logs"},
	})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("failed to publish logs: %w", err)
	}
	return nil
}
