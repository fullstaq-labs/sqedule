# Installation with container

Use the image `ghcr.io/fullstaq-labs/sqedule-server`.

 * You must pass configuration via environment variables. Learn more in [Configuration](../config/index.md). At minimum you need to configure the database type and credentials.
 * Inside the container, the Sqedule server listens on port 3001.
 * You don't need to manually setup database schemas. The Sqedule server takes care of that automatically during startup.

Example:

~~~bash
docker run --rm \
  -p 3001:3001 \
  -e SQEDULE_DB_TYPE=postgresql \
  -e SQEDULE_DB_CONNECTION='dbname=sqedule user=sqedule password=something host=localhost port=5432' \
  ghcr.io/fullstaq-labs/sqedule-server
~~~

Try it out, you should see [the web interface's](../../user_guide/concepts/web-interface.md) HTML:

~~~bash
curl localhost:3001
~~~

## Next up

Now that it's installed, please be aware of the [security considerations](../concepts/security.md).
