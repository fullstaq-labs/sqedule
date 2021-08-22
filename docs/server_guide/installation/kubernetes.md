# Installation with Kubernetes

## 1. Create Secret

Create a secret containing the database location and credentials.

First, create a [database connection string](../config/postgresql.md) and encode it as Base64. For example:

~~~bash
echo -n 'dbname=sqedule user=sqedule password=something host=localhost port=5432' | base64
~~~

Then create a Kubernetes Secret:

~~~yaml
apiVersion: v1
kind: Secret
metadata:
  name: sqedule-db-connection
type: Opaque
data:
  db_connection: <BASE64 DATA HERE>
~~~

## 2. Create Deployment

Create a Kubernetes Deployment.

 * You must pass configuration via environment variables. Learn more in [Configuration](../config/index.md). At minimum you need to configure the database type and credentials.
 * You don't need to manually setup database schemas. The Sqedule server [takes care of that automatically during startup](../concepts/database-schema-migration.md).
 * Sqedule in its default configuration does not support running multiple instances. Therefore, unless you've taken steps to [make Sqedule multi-instance-safe](../concepts/multi-instance-safety.md), you must only run a single replica and you must only use the `recreate` update strategy.

Example:

~~~yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sqedule
  labels:
    app: sqedule
spec:
  selector:
    matchLabels:
      app: sqedule
  strategy:
    type: recreate
  template:
    metadata:
      labels:
        app: sqedule
    spec:
      containers:
        - name: sqedule
          image: ghcr.io/fullstaq-labs/sqedule-server
          ports:
            - containerPort: 3001
          env:
            - name: SQEDULE_DB_TYPE
              value: postgresql
            - name: SQEDULE_DB_CONNECTION
              valueFrom:
                secretKeyRef:
                  name: sqedule-db-connection
                  key: db_connection
~~~

## 3. Create Service

Create a Kubernetes Service so that the Sqedule server can be accessed through the network.

~~~yaml
apiVersion: v1
kind: Service
metadata:
  name: sqedule
spec:
  selector:
    app: sqedule
  ports:
    - protocol: TCP
      port: 80
      targetPort: 3001
~~~

## Next up

Now that it's installed, please be aware of the [security considerations](../concepts/security.md).
