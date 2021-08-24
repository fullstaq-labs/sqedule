# Tutorial 1: release audit logging

In this tutorial, we will:

 1. Create a Github repository with a CD pipeline.
 2. Configure the [CLI](../concepts/cli.md).
 3. Register an [application](../concepts/applications-releases.md) in Sqedule using the CLI. We will touch upon the [Versioning](../concepts/versioning.md) concept.
 4. Integrate a CD pipeline with Sqedule, so that the pipeline registers a release in Sqedule. The pipeline will make use of the CLI.
 5. View a list of releases for the registered applications.

<figure>
  <img src="../release-logging.drawio.svg" alt="Diagram of a CD pipeline integrating with Sqedule">
  <figcaption>We will integrate a CD pipeline with Sqedule. The pipeline will register releases in the Sqedule server.</figcaption>
</figure>

## Before you begin

Please make sure you've done the following before proceeding with this tutorial:

 * [Setup a Sqedule server](../../server_guide/index.md)
 * [Install the CLI](../tasks/install-cli.md)

## 1 Create Github repo with pipeline

Create a new Git repo:

~~~bash
mkdir sqedule-tutorial-1
cd sqedule-tutorial-1
git init
~~~

Add a new Github workflow representing our CD pipeline. Since this is just a tutorial, this pipeline doesn't actually deploy anything and merely prints "I have deployed". In `.github/workflows/cd.yml`:

~~~yaml
name: CD

on:
  push: {}

jobs:
  deploy:
    runs-on: ubuntu-20.04
    steps:
      - uses: actions/checkout@v2
      - run: echo I have deployed
~~~

Commit the file:

~~~bash
git add .github/workflows/ci.yml
git commit -a -m "Initial commit"
~~~

Then [create a Github repo](https://github.com/new) and push this Git repo to Github.

## 2 Configure CLI

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

## 3 Register an application

Use the CLI to register an application. This requires two parameters:

 * An ID meant for machines. It will also be used in URLs. Example: `shopping_cart_service`
 * A human-readable display name. Example: "Shopping Cart Service"

=== "Unix"

    ~~~bash
    sqedule application create \
      --id shopping_cart_service \
      --display-name 'Shopping Cart Service' \
      --proposal-state final
    ~~~

=== "Windows (cmd)"

    ~~~batch
    sqedule application create ^
      --id shopping_cart_service ^
      --display-name "Shopping Cart Service" ^
      --proposal-state final
    ~~~

!!! info "Proposal state?"

    Many resources in Sqedule are versioned. There are two kinds of versions: proposals (unapproved versions), and approved versions. Proposals can be modified at will and have no effect. Once you are satisfied with the proposal, you finalize the proposal, after which it may become an approved version (approval is done either by change managers or automatically by the system).

    `sqedule xxx create` commands don't create finalized versions by default. Instead, they create draft (unfinalized) proposals. The `--proposal-state final` flag indicates that we want to finalize the proposal and submit it for approval.

    Learn more about the concept: [Versioning, proposals & reviews](../concepts/versioning.md).

The CLI tells you that the operation is successful, and outputs the properties of the registered application:

~~~json
{
    "created_at": "2021-08-22T14:44:14.115866+02:00",
    "id": "shopping_cart_service",
    "updated_at": "2021-08-22T14:44:14.115866+02:00",
    "version": {
        "proposal_state": "approved",
        "approved_at": "2021-08-22T14:44:14.115686+02:00",
        "created_at": "2021-08-22T14:44:14.11635+02:00",
        "display_name": "Shopping Cart Service",
        "enabled": true,
        "id": 738,
        "updated_at": "2021-08-22T14:44:14.11722+02:00",
        "version_number": 1,
        "version_state": "approved"
    }
}
--------------------
🎉 Application created!
💡 It has been auto-approved by the system. To view it, use `sqedule application describe`
~~~

### Viewing the registered application

#### With the CLI

Run:

~~~bash
sqedule application describe --id shopping_cart_service
~~~

The CLI outputs the properties of the application:

~~~json
{
    "approval_ruleset_bindings": [],
    "created_at": "2021-08-22T14:44:14.115866+02:00",
    "id": "shopping_cart_service",
    "latest_approved_version": {
        "proposal_state": "approved",
        "approved_at": "2021-08-22T14:44:14.115686+02:00",
        "created_at": "2021-08-22T14:44:14.11635+02:00",
        "display_name": "Shopping Cart Service",
        "enabled": true,
        "id": 738,
        "updated_at": "2021-08-22T14:44:14.11722+02:00",
        "version_number": 1,
        "version_state": "approved"
    },
    "updated_at": "2021-08-22T14:44:14.115866+02:00"
}
~~~

#### With the web interface

Visit your Sqedule server's base URL. That will open up the [web interface](../concepts/web-interface.md).

Click "Applications" and you should see your application.

## 4 Integrate with CD pipeline

Modify `.github/workflows/cd.yml` to:

 1. Download and extract the Sqedule CLI.
 2. Use the Sqedule CLI to register a release before deploying.

~~~yaml
name: CD

on:
  push: {}

jobs:
  deploy:
    runs-on: ubuntu-20.04
    steps:
      - uses: actions/checkout@v2

      ### Begin modification
      - name: Download Sqedule CLI
        run: curl -fsSLo sqedule-cli.tar.gz https://github.com/fullstaq-labs/sqedule/releases/download/v{{ latest_version }}/sqedule-cli-{{ latest_version }}-x86_64-linux.tar.gz
      - name: Extract Sqedule CLI
        run: tar -xzf sqedule-cli.tar.gz --strip-components=1
      - name: Register release
        run: ./sqedule release create
        env:
          SQEDULE_SERVER_BASE_URL: https://your-sqedule-server.com  # <--- change this!
          SQEDULE_APPLICATION_ID: shopping_cart_service
          SQEDULE_METADATA: |
            {
              "commit_sha": "${{ "{{ github.sha }}" }}",
              "run_number": ${{ "{{ github.run_number }}" }}
            }
      ### End modification

      - run: echo I have deployed
~~~

The most important addition is the `sqedule release create` invocation — this is the CLI command that registers a release. This command accepts the following parameters (which we pass via environment variables):

 * The Sqedule server URL.
 * The ID of the application for which we want to register a release.
 * Arbitrary JSON metadata to include in the release. In this case, we include the Git commit SHA and the Github Actions run number.

Commit and push the changes:

~~~bash
git commit -a -m "Log releases in Sqedule"
git push
~~~

Then wait until the CD pipeline finishes.

## 5 View release

Visit your Sqedule server's base URL. That will open up the [web interface](../concepts/web-interface.md).

Click "Releases" and you should see the release that just occurred.

## Conclusion

Congratulations, you've integrated a CD pipeline with Sqedule! Every time the pipeline deploys, it registers a new release in Sqedule.

In [the next tutorial](approval.md), we'll not only use Sqedule to log releases, but also to approve releases.
