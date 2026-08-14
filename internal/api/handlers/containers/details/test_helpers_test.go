package details

import (
	"github.com/rs/zerolog"

	"github.com/sidneyojr/watchtower/internal/logging"
)

func testLogger() *zerolog.Logger { return logging.NopLogger() }
