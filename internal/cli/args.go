package cli

import (
	"fmt"
	"strings"
)

// args is a command's parsed command line. Flags may come before, between
// or after positional arguments, as agents write them in any order.
type args struct {
	positional []string
	bools      map[string]bool
	values     map[string][]string
}

// parseArgs parses raw against the command's boolean and value flags,
// given without their leading dashes.
func parseArgs(raw []string, boolFlags, valueFlags []string) (*args, error) {
	a := &args{bools: map[string]bool{}, values: map[string][]string{}}
	isBool := map[string]bool{}
	for _, f := range boolFlags {
		isBool[f] = true
	}
	isValue := map[string]bool{}
	for _, f := range valueFlags {
		isValue[f] = true
	}
	for i := 0; i < len(raw); i++ {
		arg := raw[i]
		if !strings.HasPrefix(arg, "--") {
			a.positional = append(a.positional, arg)
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimPrefix(arg, "--"), "=")
		switch {
		case isBool[name] && !hasValue:
			a.bools[name] = true
		case isValue[name] && hasValue:
			a.values[name] = append(a.values[name], value)
		case isValue[name] && i+1 < len(raw):
			i++
			a.values[name] = append(a.values[name], raw[i])
		case isValue[name]:
			return nil, fmt.Errorf("--%s needs a value", name)
		default:
			return nil, fmt.Errorf("unknown flag --%s", name)
		}
	}
	return a, nil
}

// value returns the last value given for a flag, or "".
func (a *args) value(name string) string {
	v := a.values[name]
	if len(v) == 0 {
		return ""
	}
	return v[len(v)-1]
}
