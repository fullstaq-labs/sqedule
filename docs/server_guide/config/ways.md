# Ways to pass configuration

There are 3 ways to pass configuration to the Sqedule server.

## 1. Config file

You can pass configuration from the configuration file. This is recommended when you've [installed the Sqedule server via the binary (without containerization)](../installation/binary.md).

The file is in YAML format. The Sqedule server looks for the configuration file in `~/.sqedule-server.yaml` by default. You can customize the location by [running the Sqedule server](../concepts/server-exe.md) with `--config /path-to-actual-file.yml`.

## 2. Environment variables

You can pass configuration from environment variables. This is recommended when you're using [containerization](../installation/container.md) or [Kubernetes](../installation/kubernetes.md).

## 3. CLI parameters

You can pass some configuration through CLI parameters when [running the Sqedule server](../concepts/server-exe.md). Which CLI parameters are accepted depends on the specific subcommand.

!!! warning "Caveat"

    There's a caveat with setting boolean options through CLI parameters. Learn more in [Configuration naming format](naming.md).

## Precedence

Configuration is loaded in the given order (least to most important):

 1. Config file
 2. Environment variables
 3. CLI parameters

CLI parameters override environment variables. Environment variables override the configuration file.
