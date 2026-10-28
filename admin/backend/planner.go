package main

import (
	"context"

	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/controlplane"
	"github.com/dirmich/marubot/pkg/orchestration"
	"github.com/dirmich/marubot/pkg/providers"
)

func loadPlanner(path string) (func(context.Context, string, []orchestration.Node) (orchestration.Run, error), error) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		return nil, err
	}
	return controlplane.Planner(provider, cfg.Agents.Defaults.Model), nil
}
