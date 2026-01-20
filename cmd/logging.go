// Copyright 2026 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		log.Trace().Msg(strings.TrimSpace(msg))
	} else {
		// Default to info level
		log.Info().Msg(msg)
	}

	return len(p), nil
}
