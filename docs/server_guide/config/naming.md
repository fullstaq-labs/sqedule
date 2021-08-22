# Configuration naming format

Configuration options have a **canonical name**, for example `db-type`.

 - When using the configuration file, specify configuration options using their canonical names. Example: `db-type: something`
 - When using environment variables, specify configuration options with the `SQEDULE_` prefix, in uppercase, with dashes (`-`) replaced by underscores (`_`). Example: `SQEDULE_DB_TYPE=something`
 - When using CLI parameters, specify two dashes (`--`) followed by the canonical name. Example: `--db-type=something`

   !!! warning "Caveat"
       When setting a boolean option using a CLI parameter, be sure to specify the value after a `=`, like this: `--something=true_or_false`. Specifying the value after a space, like `--something true_or_false` doesn't work.

!!! note
    The rest of the Server Guide only specifies configuration names in their canonical format.
