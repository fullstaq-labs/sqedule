package cli

import "github.com/spf13/cobra"

func DefineServerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("server-base-url", "S", "", "server base URL (required). Example: https://sqedule.mydomain.net")
}
