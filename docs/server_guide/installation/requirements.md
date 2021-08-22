# Installation requirements

The Sqedule server requires PostgreSQL >= 13. Other databases may be supported in the future depending on user demand.

!!! note
    You don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](../concepts/database-schema-migration.md).

In theory, the server can be run on all operating systems that the Go programming language supports. But we've only tested on macOS and Linux, and we only provide precompiled binaries for Linux (x86\_64).
