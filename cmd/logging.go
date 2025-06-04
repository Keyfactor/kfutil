package cmd

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// zerologWriter implements io.Writer and forwards standard log output to zerolog
type zerologWriter struct{}

func (w zerologWriter) Write(p []byte) (n int, err error) {
	// Clean up the log message (remove timestamp, etc.)
	msg := string(p)
	msg = strings.TrimSpace(msg)

	// Check if it's a debug message
	if strings.Contains(msg, "[DEBUG]") {
		msg = strings.Replace(msg, "[DEBUG]", "", 1)
		log.Debug().Msg(strings.TrimSpace(msg))
	} else if strings.Contains(msg, "[ERROR]") {
		msg = strings.Replace(msg, "[ERROR]", "", 1)
		log.Error().Msg(strings.TrimSpace(msg))
	} else if strings.Contains(msg, "[INFO]") {
		msg = strings.Replace(msg, "[INFO]", "", 1)
		log.Info().Msg(strings.TrimSpace(msg))

	} else if strings.Contains(msg, "[WARN]") {
		msg = strings.Replace(msg, "[WARN]", "", 1)
		log.Warn().Msg(strings.TrimSpace(msg))

	} else if strings.Contains(msg, "[FATAL]") {
		msg = strings.Replace(msg, "[FATAL]", "", 1)
		log.Fatal().Msg(strings.TrimSpace(msg))

	} else if strings.Contains(msg, "[TRACE]") {
		msg = strings.Replace(msg, "[TRACE]", "", 1)
		log.Debug().Msg(strings.TrimSpace(msg)) // Converting TRACE to DEBUG for zerolog
	} else {
		// Default to info level
		log.Info().Msg(msg)
	}

	return len(p), nil
}
