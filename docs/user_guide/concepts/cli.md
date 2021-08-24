# About the CLI

We provide a CLI for viewing and manipulating resources in the Sqedule server. It is an alternative to the [web interface](web-interface.md). Under the hood, the CLI makes use the of [API](api.md), so if you feel adventurous enough you can also just use the API directly.

The CLI supports all operating systems that the Go programming language supports, but we only provide precompiled binaries for macOS (x86\_64, M1) Windows (x86\_64) and Linux (x86\_64).

## Subcommands

The CLI supports multiple subcommands, similar to how Git works. You can view a list of all subcommands by running `sqedule --help`. It should look similar to this:

~~~
Sqedule client CLI

Usage:
  sqedule [command]

Available Commands:
  application                          Manage applications
  application-approval-ruleset-binding Manage application approval ruleset bindings
  approval-ruleset                     Manage approval rulesets
  help                                 Help about any command
  release                              Manage releases
  version                              Show CLI version

Flags:
      --config string   config file (default ~/.config/sqedule-cli/config.yml)
      --debug           show API requests/responses
  -h, --help            help for sqedule

Use "sqedule [command] --help" for more information about a command.
~~~

## See also

 * [The tutorials](../tutorials/release-logging.md), which demonstrate the use of the CLI
 * [Installing the CLI](../tasks/install-cli.md)
 * [Initial CLI setup](../tasks/initial-cli-setup.md)
 * [CLI configuration reference](../references/cli-config.md)
