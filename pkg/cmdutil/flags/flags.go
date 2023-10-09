/*
Copyright 2023 The Keyfactor Command Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flags

import (
	"github.com/spf13/pflag"
)

// Flags represents an instance of a flag set that can be added to a command.
// This is the core interface that all flags must implement.
//
// Optionally, the following interface can be implemented:
//   - ToOptions(): ToOptions is called when the flag set must be converted to a set of
//     options that can be used to modify execution of a command.
type Flags interface {
	// AddFlags is called when the flag set must be added to a command.
	AddFlags(flags *pflag.FlagSet)
}

type Options interface {
	Validate() error
}
