package flags

import (
	"github.com/spf13/cobra"

	"github.com/sidneyojr/watchtower/internal/flags/api"
	"github.com/sidneyojr/watchtower/internal/flags/client"
	"github.com/sidneyojr/watchtower/internal/flags/compat"
	"github.com/sidneyojr/watchtower/internal/flags/docker"
	"github.com/sidneyojr/watchtower/internal/flags/filter"
	"github.com/sidneyojr/watchtower/internal/flags/lifecycle"
	"github.com/sidneyojr/watchtower/internal/flags/logging"
	"github.com/sidneyojr/watchtower/internal/flags/mode"
	"github.com/sidneyojr/watchtower/internal/flags/notify"
	"github.com/sidneyojr/watchtower/internal/flags/registry"
	"github.com/sidneyojr/watchtower/internal/flags/schedule"
	"github.com/sidneyojr/watchtower/internal/flags/spec"
	"github.com/sidneyojr/watchtower/internal/flags/update"
)

// RegisterAll registers every domain's flags on the root command.
//
// Domain packages match the config taxonomy: docker, client, schedule, mode,
// update, lifecycle, filter, registry, compat, api, notify, logging.
//
// Parameters:
//   - rootCmd: Root Cobra command.
func RegisterAll(rootCmd *cobra.Command) {
	docker.Register(rootCmd)
	client.Register(rootCmd)
	schedule.Register(rootCmd)
	mode.Register(rootCmd)
	update.Register(rootCmd)
	lifecycle.Register(rootCmd)
	filter.Register(rootCmd)
	registry.Register(rootCmd)
	compat.Register(rootCmd)
	api.Register(rootCmd)
	notify.Register(rootCmd)
	logging.Register(rootCmd)
}

// AllSpecs returns aggregated FlagSpec rows from every domain.
//
// Returns:
//   - []spec.FlagSpec: Flag metadata for BindAll.
func AllSpecs() []spec.FlagSpec {
	return CollectSpecs(
		docker.Specs(),
		client.Specs(),
		schedule.Specs(),
		mode.Specs(),
		update.Specs(),
		lifecycle.Specs(),
		filter.Specs(),
		registry.Specs(),
		compat.Specs(),
		api.Specs(),
		notify.Specs(),
		logging.Specs(),
	)
}
