package auth

import (
	"github.com/rs/zerolog"

	"github.com/sidneyojr/watchtower/internal/logging"
)

func testLog() *zerolog.Logger {
	return logging.NopLogger()
}
