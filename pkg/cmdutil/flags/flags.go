package flags

import "github.com/spf13/pflag"

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
