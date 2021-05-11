package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// This file contains functions for checking whether configuration options are set.

// ConfigRequirementSpec describes which configuration options must be set,
// and what "not set" exactly means.
type ConfigRequirementSpec struct {
	// StringNonEmpty contains a list of Viper config names, whose
	// String value must be non-empty.
	StringNonEmpty []string
}

// Names returns all configuration option names that appear in this spec.
func (spec ConfigRequirementSpec) Names() []string {
	result := make([]string, 0)
	for _, name := range spec.StringNonEmpty {
		result = append(result, name)
	}
	return result
}

// RequireConfigOptions checks whether the given configuration options
// are set. Returns an error if some are missing, nil if everything is
// present.
func RequireConfigOptions(viper *viper.Viper, spec ConfigRequirementSpec) error {
	missing, _ := checkConfigRequirements(viper, spec)
	if len(missing) > 0 {
		return fmt.Errorf("Configuration required: %s", strings.Join(missing, ", "))
	} else {
		return nil
	}
}

// RequireOneOfConfigOptions checks whether exactly one of the given configuration
// options is set. Returns an error if none are set or if more than one is set.
// Returns nil if exactly one is present.
func RequireOneOfConfigOptions(viper *viper.Viper, spec ConfigRequirementSpec) error {
	_, numSet := checkConfigRequirements(viper, spec)

	if numSet == 0 {
		return fmt.Errorf("One of these configurations is required: %s", strings.Join(spec.Names(), ", "))
	} else if numSet > 1 {
		return fmt.Errorf("Exactly one of these configurations may be set: %s", strings.Join(spec.Names(), ", "))
	} else {
		return nil
	}
}

func checkConfigRequirements(viper *viper.Viper, spec ConfigRequirementSpec) (missing []string, numSet uint) {
	missing = make([]string, 0)

	for _, name := range spec.StringNonEmpty {
		if len(viper.GetString(name)) == 0 {
			missing = append(missing, name)
		} else {
			numSet++
		}
	}

	return missing, numSet
}
