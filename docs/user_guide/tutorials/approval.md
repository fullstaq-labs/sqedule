# Tutorial 2: release approval

In the [tutorial 1](release-logging.md) we integrated a CD pipeline with Sqedule so that all releases are logged in Sqedule. In this tutorial we'll not only use Sqedule to log releases, but also to approve releases. That is: the pipeline won't actually deploy until Sqedule has approved the release. To this end, we will also define [approval rules](../concepts/approval-rules.md).

<figure>
  <img src="../../concepts/applications-releases.drawio.svg" alt="Diagram of a CD pipeline asking Sqedule for release approval">
  <figcaption>The CD pipeline will not only log releases in Sqedule, but will also wait for Sqedule's approval before deploying.</figcaption>
</figure>

## Before you begin

Please make sure you've done the following before proceeding with this tutorial:

 * [Setup a Sqedule server](../../server_guide/index.md)
 * [Install the CLI](../tasks/install-cli.md)
 * [Follow tutorial 1](release-logging.md)

## 1 Define approval rules

Important concepts to know when defining approval rules:

 * Individual rules are grouped in a _ruleset_.
 * A ruleset must be _bound_ to an application. New releases for that application will be evaluated against all bound rulesets.

At present, Sqedule only supports one type of rule: the so-called "schedule rule", which defines a time-of-day window in which releases are allowed. More rule types will be supported in the future as Sqedule receives further development.

In this tutorial we'll define a ruleset named "only afternoon", which only allows releases to occur from 12:00 to 18:00.

### 1 Create ruleset

Let's create a ruleset using the Sqedule CLI. This requires two parameters:

 * An ID meant for machines. It will also be used in URLs. Example: `only_afternoon`
 * A human-readable display name. Example: "Only afternoon"

=== "Unix"

    ~~~bash
    sqedule approval-ruleset create \
      --id only_afternoon \
      --display-name 'Only afternoon'
    ~~~

=== "Windows (cmd)"

    ~~~bash
    sqedule approval-ruleset create ^
      --id only_afternoon ^
      --display-name "Only afternoon"
    ~~~

Unlike tutorial 1, we don't specify `--proposal-state final`. That's because we aren't done editing this ruleset yet: we still need to add a rule to this ruleset. We'll finalize the proposal after we're done adding a rule.

The CLI confirms that creating this ruleset was successful:

~~~json
{
    "application_approval_ruleset_bindings": [],
    "created_at": "2021-08-22T17:17:48.14245+02:00",
    "id": "only_afternoon",
    "updated_at": "2021-08-22T17:17:48.14245+02:00",
    "version": {
        "adjustment_state": "draft",
        "approval_rules": [],
        "approved_at": null,
        "created_at": "2021-08-22T17:17:48.148982+02:00",
        "description": "",
        "display_name": "Only afternoon",
        "enabled": true,
        "globally_applicable": false,
        "id": 5,
        "release_approval_ruleset_bindings": [],
        "updated_at": "2021-08-22T17:17:48.155585+02:00",
        "version_number": null,
        "version_state": "proposal"
    }
}
--------------------
🎉 Approval ruleset 'only_afternoon' created!
💡 It is still a proposal. To view it, use `sqedule approval-ruleset proposal list`
⚠️  It has no rules yet. To add rules, use `sqedule approval-ruleset proposal rule create-...`
~~~

According to `version.version_state`, this ruleset has one version, which is a proposal.

Take note of the proposal ID! You can find that in `version.id`, which in the above example is 5. We'll need it for adding rules to this proposal.

### 2 Add rule

Let's add a schedule rule to this ruleset proposal. This requires the following parameters:

 * The ruleset's ID. Example: `only_afternoon`
 * The proposal's ID. In the above example, it's 5.
 * Schedule rule parameters.

=== "Unix"

    ~~~bash
    sqedule approval-ruleset proposal rule create-schedule \
      --approval-ruleset-id only_afternoon \
      --proposal-id 5 \
      --begin-time '12:00' \
      --end-time '18:00'
    ~~~

=== "Windows (cmd)"

    ~~~batch
    sqedule approval-ruleset proposal rule create-schedule ^
      --approval-ruleset-id only_afternoon ^
      --proposal-id 5 ^
      --begin-time 12:00 ^
      --end-time 18:00
    ~~~

The CLI confirms that the rule is successfully created:

~~~json
[
    {
        "begin_time": "12:00",
        "created_at": "2021-08-22T17:25:36.952222+02:00",
        "days_of_month": null,
        "days_of_week": null,
        "enabled": true,
        "end_time": "18:00",
        "id": 5,
        "months_of_year": null,
        "type": "schedule"
    }
]
--------------------
🎉 Rule created!
~~~

### 3 Finalize ruleset proposal

We're done editing this ruleset proposal so let's finalize it.

This requires the following parameters:

 * The ruleset's ID. Example: `only_afternoon`
 * The proposal's ID. In this example, it's 5.

=== "Unix"

    ~~~bash
    sqedule approval-ruleset proposal update \
      --approval-ruleset-id only_afternoon \
      --id 5 \
      --proposal-state final
    ~~~

=== "Windows (cmd)"

    ~~~batch
    sqedule approval-ruleset proposal update ^
      --approval-ruleset-id only_afternoon ^
      --id 5 ^
      --proposal-state final
    ~~~

The CLI confirms that finalization is successful (`version.state` is `approved`):

~~~json
{
    "application_approval_ruleset_bindings": [],
    "created_at": "2021-08-22T17:17:48.14245+02:00",
    "id": "only_afternoon",
    "updated_at": "2021-08-22T17:17:48.14245+02:00",
    "version": {
        "adjustment_state": "approved",
        "approval_rules": [
            {
                "begin_time": "12:00",
                "created_at": "2021-08-22T17:25:36.952222+02:00",
                "days_of_month": null,
                "days_of_week": null,
                "enabled": true,
                "end_time": "18:00",
                "id": 6,
                "months_of_year": null,
                "type": "schedule"
            }
        ],
        "approved_at": "2021-08-22T17:29:24.235901+02:00",
        "created_at": "2021-08-22T17:17:48.148982+02:00",
        "description": "",
        "display_name": "Only afternoon",
        "enabled": true,
        "globally_applicable": false,
        "id": 5,
        "release_approval_ruleset_bindings": [],
        "updated_at": "2021-08-22T17:29:24.252494+02:00",
        "version_number": 1,
        "version_state": "approved"
    }
}
--------------------
🎉 Proposal updated!
~~~

### 4 Bind ruleset to application

Now that the ruleset is finished, we must bind it to an application. Only after binding will new releases be evaluated against this ruleset.

To bind, we create an _application approval ruleset binding_. This requires the following parameters:

 * The ID of the application to bind against. Example: `shopping_cart_service`
 * The ID of the ruleset to bind against. Example: `only_afternoon`

=== "Unix"

    ~~~bash
    sqedule application-approval-ruleset-binding create \
      --application-id shopping_cart_service \
      --approval-ruleset-id only_afternoon \
      --proposal-state final
    ~~~

=== "Windows (cmd)"

    ~~~batch
    sqedule application-approval-ruleset-binding create ^
      --application-id shopping_cart_service ^
      --approval-ruleset-id only_afternoon ^
      --proposal-state final
    ~~~

The CLI confirms that the binding is created and that its version is finalized:

~~~json
{
    "application": {
        "created_at": "2021-08-22T14:44:14.115866+02:00",
        "id": "shopping_cart_service",
        "latest_approved_version": {
            "adjustment_state": "approved",
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
    },
    "approval_ruleset": {
        "created_at": "2021-08-22T17:17:48.14245+02:00",
        "id": "only_afternoon",
        "latest_approved_version": {
            "adjustment_state": "approved",
            "approved_at": "2021-08-22T17:29:24.235901+02:00",
            "created_at": "2021-08-22T17:17:48.148982+02:00",
            "description": "",
            "display_name": "Only afternoon",
            "enabled": true,
            "globally_applicable": false,
            "id": 5,
            "updated_at": "2021-08-22T17:29:24.252494+02:00",
            "version_number": 1,
            "version_state": "approved"
        },
        "updated_at": "2021-08-22T17:17:48.14245+02:00"
    },
    "created_at": "2021-08-22T20:13:44.264493+02:00",
    "updated_at": "2021-08-22T20:13:44.264493+02:00",
    "version": {
        "adjustment_state": "approved",
        "approved_at": "2021-08-22T20:13:44.264346+02:00",
        "created_at": "2021-08-22T20:13:44.267858+02:00",
        "id": 7,
        "mode": "enforcing",
        "updated_at": "2021-08-22T20:13:44.273691+02:00",
        "version_number": 1,
        "version_state": "approved"
    }
}
--------------------
🎉 Application approval ruleset binding binding created!
💡 It has been auto-approved by the system. To view it, use `sqedule application-approval-ruleset-binding describe`
~~~

## 2 Modify pipeline

Open `.github/workflows/cd.yml`. Modify the "Register release" step and add the environment variable `SQEDULE_WAIT: true`. This parameter tells `sqedule release create` to not only register a release, but also to wait until it's either approved or rejected.

~~~yaml
name: CD

on:
  push: {}

jobs:
  deploy:
    runs-on: ubuntu-20.04
    steps:
      - uses: actions/checkout@v2

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
          ### Begin modification
          SQEDULE_WAIT: true
          ### End modification

      - run: echo I have deployed
~~~

Commit and push the changes:

~~~bash
git commit -a -m "Approve releases via Sqedule"
git push
~~~

Then wait until the CD pipeline finishes. You should see that it waits until the Sqedule server approves or rejects the release:

~~~
Checking release state...
Current state: pending
Checking release state...
Current state: pending
...
Checking release state...
Current state: approved
~~~

## About polling in the pipeline

At present, deploying only after a release is approved or rejected by Sqedule, is implemented through polling in the pipeline. As you can imagine, this is suboptimal: it is inefficient and also ties up a pipeline executor or concurrency slot.

At present, polling is the only release gate strategy that Sqedule supports because it's easy to implement and works with every pipeline. We plan to support additional release gate strategies in the future that don't involve polling (e.g. the ability to trigger Github/Gitlab pipelines from Sqedule after a release is approved/rejected).

## Conclusion

Congratulations, you've implemented a pipeline that deploys only when the Sqedule server approves the release!

We recommend that you read the Concepts section in order to familiarize yourself with Sqedule.
