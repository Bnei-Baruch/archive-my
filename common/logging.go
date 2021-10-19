package common

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	zerolog.CallerFieldName = "line"
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		rel := strings.Split(file, "archive-my/")
		return fmt.Sprintf("%s:%d", rel[1], line)
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFieldName = "timestamp"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano}
	log.Logger = log.Output(output).With().Caller().Stack().Logger()
}
