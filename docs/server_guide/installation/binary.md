# Installation with the binary (without containerization)

## 1. Create user account

Create a user account to run the Sqedule server in. For example:

~~~basH
sudo addgroup --gid 3420 sqedule-server
sudo adduser --uid 3420 --gid 3420 --disabled-password --gecos 'Sqedule Server' sqedule-server
~~~

## 2. Download

[Download a Sqedule server binary tarball](https://github.com/fullstaq-labs/sqedule/releases) (`sqedule-server-XXX-linux-amd64.tar.gz`).

Extract the tarball. There's a [`sqedule-server` executable](../concepts/server-exe.md) inside. Check whether it works:

~~~bash
/path-to/sqedule-server --help
~~~

## 3. Create config file

Create a Sqedule server configuration file `/etc/sqedule-server.yml`. Learn more in [Configuration](../config/index.md).

At minimum you need to configure the database type and credentials. Example:

~~~yaml
db-type: postgresql
db-connection: 'dbname=sqedule user=sqedule password=something host=localhost port=5432'
~~~

Be sure to give the file the right permissions so that the database password cannot be read by others:

~~~bash
sudo chown sqedule-server: /etc/sqedule-server.yml
sudo chmod 600 /etc/sqedule-server.yml
~~~

## 4. Install SystemD service

Install a SystemD service file. Create /etc/systemd/system/sqedule-server.service:

~~~systemd
[Unit]
Description=Sqedule Server

[Service]
ExecStart=/path-to/sqedule-server run --config=/etc/sqedule-server.yml
User=sqedule-server
PrivateTmp=true

[Install]
WantedBy=multi-user.target
~~~

!!! note
    Be sure to replace `/path-to/sqedule-server`.

Then:

~~~bash
sudo systemctl daemon-reload
~~~

## 5. Start Sqedule server

Start the Sqedule server:

~~~bash
sudo systemctl start sqedule-server
~~~

!!! note
    You don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](../concepts/database-schema-migration.md).

It listens on localhost port 3001 by default. Try it out, you should see [the web interface's](../../user_guide/web_interface.md) HTML:

~~~bash
curl localhost:3001
~~~

## Next up

Now that it's installed, please be aware of the [security considerations](../concepts/security.md).
