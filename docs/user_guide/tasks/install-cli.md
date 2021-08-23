# Installing the CLI

!!! info "See also"
    Concept: [About the CLI](../concepts/cli.md)

[Download a Sqedule CLI binary tarball](https://github.com/fullstaq-labs/sqedule/releases) (`sqedule-cli-XXX-linux-x86_64.tar.gz`).

Extract this tarball somewhere. For example:

~~~bash
cd /usr/local
tar xzf sqedule-cli-XXX.tar.gz
~~~

The Sqedule CLI is now available as `/usr/local/sqedule-cli-XXX/sqedule`. We recommend adding the directory to PATH.

Try it out:

~~~bash
export PATH=/usr/local/sqedule-cli-XXX:$PATH
sqedule version
~~~

## Next up

You must [setup the CLI's initial configuration](initial-cli-setup.md) before the CLI can be used.
