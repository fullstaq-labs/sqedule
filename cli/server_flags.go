package cli

import "github.com/spf13/pflag"

func DefineServerFlags(flags *pflag.FlagSet) {
	flags.StringP("server-base-url", "S", "", "server base URL (required). Example: https://sqedule.mydomain.net")
}
