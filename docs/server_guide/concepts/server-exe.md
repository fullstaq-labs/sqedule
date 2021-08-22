# The Sqedule server executable

The Sqedule server comes with an executable named `sqedule-server`. This executable supports multiple subcommands, similar to how Git works. The most often used subcommand is `run`, which runs the HTTP server.

You normally don't need to interact with this executable directly, because you're either using this executable through the SystemD service file, or through a container. But sometimes, such as when [manually migrating the database schema](../tasks/manual-database-schema-migration.md), you do need to interact with it.

## Invoking the Sqedule server executable

The way to invoke the Sqedule server executable depends on the installation method.

 - When [installed directly as a binary (without containerization)](../installation/binary.md), invoke `sqedule-server` directly from your shell. Example:

   ~~~bash
   sqedule-server --help
   ~~~

   If `sqedule-server` is not in PATH, then invoke its full path: `/path-to/sqedule-server`

 - When [using containerization](../installation/container.md), invoke it through a container. The image's entrypoint is the Sqedule server executable. Example:

   ~~~bash
   docker run -ti --rm ghcr.io/fullstaq-labs/sqedule-server --help
   ~~~

 - When [using Kubernetes](../installation/kubernetes.md), obtain a shell in a Sqedule pod:

   ~~~bash
   kubectl exec -ti deploy/sqedule -- sh
   ~~~

   The `sqedule-server` executable is in PATH so you can invoke from this shell:

   ~~~bash
   sqedule-server --help
   ~~~

   !!! note
       The container environment is Alpine.

## Subcommands

To see what subcommands `sqedule-server` supports, run it with `--help`. You should see something like this:

~~~
$ sqedule-server --help
Sqedule server

Usage:
  sqedule-server [command]

Available Commands:
  db          Manage the database
  help        Help about any command
  run         Run the Sqedule API server

Flags:
      --config string      config file (default $HOME/.sqedule-server.yaml)
  -h, --help               help for sqedule-server
      --log-level string   log level, one of: error,warn,info,silent (default "info")

Use "sqedule-server [command] --help" for more information about a command.
~~~
