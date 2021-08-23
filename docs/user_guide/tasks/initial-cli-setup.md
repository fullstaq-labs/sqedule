# Initial CLI setup

Before we can use the Sqedule CLI, we must configure it to tell it which Sqedule server to use. Edit the Sqedule configuration file:

 * Unix: `~/.sqedule-cli.yaml`
 * Windows: `C:\Users\<Username>\.sqedule-cli.yaml`

In this file:

~~~yaml
server-base-url: https://your-sqedule-server.com
~~~

If the Sqedule server is behind a reverse proxy with HTTP Basic Authentication, then also specify the Basic Authentication credentials:

~~~yaml
basic-auth-user: <username here>
basic-auth-password: <password here>
~~~

On Unix, be sure to restrict the file's permissions so that the credentials can't be read by other users:

~~~bash
chmod 600 ~/.sqedule-cli.yaml
~~~
