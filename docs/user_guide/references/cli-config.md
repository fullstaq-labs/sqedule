# CLI configuration reference

!!! info "See also"
    Concept: [CLI](../concepts/cli.md)

## Ways to pass configuration

There are 3 ways to pass configuration to the CLI.

 1. **Configuration file**. It's located in `~/.sqedule-cli.yaml` by default. You can customize the location by running the CLI with `--config /path-to-actual-file.yml`.
 2. **Environment variables**. In some CI/CD pipelines, it's more convenient to set environment variables than to set CLI parameters. The [tutorials](../tutorials/release-logging.md) demonstrate passing configuration via environment variables in the CI/CD pipeline.
 3. **CLI parameters**.

## Naming format

Configuration options have a **canonical name**, for example `server-base-url`.

 - When using the configuration file, specify configuration options using their canonical names. Example: `server-base-url: https://your-sqedule-server.com`
 - When using environment variables, specify configuration options with the `SQEDULE_` prefix, in uppercase, with dashes (`-`) replaced by underscores (`_`). Example: `SQEDULE_SERVER_BASE_URL: https://your-sqedule-server.com`
 - When using CLI parameters, specify two dashes (`--`) followed by the canonical name. Example: `--server-base-url=something`

!!! warning "Caveat"
    When setting a boolean option using a CLI parameter, be sure to specify the value after a `=`, like this: `--something=true_or_false`. Specifying the value after a space, like `--something true_or_false` doesn't work.

!!! note
    The rest of the User Guide only specifies configuration names in their canonical format.

## Precedence

Configuration is loaded in the given order (least to most important):

 1. Config file
 2. Environment variables
 3. CLI parameters

CLI parameters override environment variables. Environment variables override the configuration file.

## Global options

These options apply to all subcommands.

 * `server-base-url` (string, required) — The base URL of the Sqedule server to use. Example: `https://your-sqedule-server.com`
 * `basic-auth-user` (string) — If the Sqedule server is protected by HTTP basic authentication, then specify the username here.
 * `basic-auth-password` (string) — If the Sqedule server is protected by HTTP basic authentication, then specify the password here.
 * `debug` (boolean, default: false)

## Subcommand-specific options

To learn about subcommand-specific options, please run the subcommand with `--help`. The supported CLI parameters are all the configuration options that that subcommand supports.
