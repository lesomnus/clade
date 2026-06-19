package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/lesomnus/clade/cmd/version"
	"github.com/lesomnus/mkot"
	"github.com/lesomnus/mkot/pretty"
	"github.com/lesomnus/otx"
	"go.opentelemetry.io/otel/attribute"
)

type OtelConfig struct {
	mkot.Config `yaml:",inline"`
}

func (c *OtelConfig) Build(ctx context.Context) (context.Context, *otx.Otx, error) {
	otc := mkot.NewConfig()
	if c != nil {
		otc = &c.Config
	}

	if otc.Processors == nil {
		otc.Processors = map[mkot.Id]mkot.ProcessorConfig{}
	}
	if otc.Exporters == nil {
		otc.Exporters = map[mkot.Id]mkot.ExporterConfig{}
	}
	if otc.Providers == nil {
		otc.Providers = map[mkot.Id]*mkot.ProviderConfig{}
	}

	const ServiceResourceId mkot.Id = "resource/clade"
	if _, ok := otc.Processors[ServiceResourceId]; !ok {
		otc.Processors[ServiceResourceId] = &mkot.Resource{
			Attributes: []mkot.Attr{
				{Key: "service.name", Value: attribute.StringValue("clade")},
				{Key: "service.version", Value: attribute.StringValue(version.Get().Version)},
			},
		}
	}
	if _, ok := otc.Exporters["pretty"]; !ok {
		otc.Exporters["pretty"] = pretty.ExporterConfig{}
	}
	if _, ok := otc.Providers["tracer"]; !ok {
		otc.Providers["tracer"] = &mkot.ProviderConfig{
			Processors: []mkot.Id{ServiceResourceId},
		}
	}
	if _, ok := otc.Providers["meter"]; !ok {
		otc.Providers["meter"] = &mkot.ProviderConfig{
			Processors: []mkot.Id{ServiceResourceId},
		}
	}
	if _, ok := otc.Providers["logger"]; !ok {
		otc.Providers["logger"] = &mkot.ProviderConfig{
			Exporters: []mkot.Id{"pretty"},
		}
	}

	resolver := mkot.Make(ctx, otc)

	tracer_provider, err := resolver.Tracer(ctx, "")
	if err != nil && !errors.Is(err, mkot.ErrNotExist) {
		return nil, nil, fmt.Errorf("resolve tracer provider: %w", err)
	}

	meter_provider, err := resolver.Meter(ctx, "")
	if err != nil && !errors.Is(err, mkot.ErrNotExist) {
		return nil, nil, fmt.Errorf("resolve meter provider: %w", err)
	}

	logger_provider, err := resolver.Logger(ctx, "")
	if err != nil && !errors.Is(err, mkot.ErrNotExist) {
		return nil, nil, fmt.Errorf("resolve logger provider: %w", err)
	}

	o := otx.New(
		otx.WithController(resolver),
		otx.WithTracerProvider(tracer_provider),
		otx.WithMeterProvider(meter_provider),
		otx.WithLoggerProvider(logger_provider),
	)
	return otx.Into(ctx, o), o, nil
}
