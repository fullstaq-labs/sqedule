# Versioning, proposals & reviews

Many resources in Sqedule are **versioned**. This has several reasons:

 * Versioning allows preserving an audit trail of changes.
 * Versioning prevents data modifications from interfering with on-going, long-running processes. Such processes use the data versions from when the processes begun, not the latest versions.

   For example, [approval rulesets](approval-rules.md) are versioned. But when Sqedule starts evaluating rules on a [release](applications-releases.md) (which can take a while), concurrent modifications on rulesets have no effect on the rule evaluation process: that process use the ruleset versions that existed when evaluation started.

Versioned resources version _most_ of their data, but not all. For example, resource IDs are never versioned.

## Lifecycle

A version starts as a **proposal** (meaning it's unapproved) and may later become **approved**, **rejected** or **abandoned**.

<figure>
  <img src="../versioning.drawio.svg" alt="Proposal lifecycle diagram">
  <figcaption>Proposal lifecycle. For brevity, the "rejected" version state is omitted.</figcaption>
</figure>

Proposals have no effect and can be modified at will. A proposal starts as a **draft**. Once the author(s) of a proposal are satisfied, they can **finalize** the proposal.

Upon finalization, Sqedule determines whether the proposal is eligible for automatic approval. Automatic approval applies when Sqedule determines that approving the proposal is safe because it does not change system behavior. If the proposal is not eligible for automatic approval, then Sqedule forwards the proposal to change managers to **manually review**.

A rejected proposal may be modified and refinalized. An unapproved proposal may be abandoned at any time. An abandoned proposal may be reopened at any time. Only the "approved" state is final.

## Relationship with JSON API output

To understand the relationship between the versioning concept and the [JSON API](api.md) output (which is also outputted by the [CLI](cli.md)), let's take a look at the following example which shows the JSON representation of an [application](applications-releases.md).

~~~json
{
    "approval_ruleset_bindings": [],
    "id": "shopping_cart_service",                     // (1)
    "created_at": "2021-08-22T14:44:14.115866+02:00",  // (1)
    "updated_at": "2021-08-22T14:44:14.115866+02:00",  // (1)
    "latest_approved_version": {            // (2)
        "adjustment_state": "approved",     // (3)
        "approved_at": "2021-08-22T14:44:14.115686+02:00",
        "created_at": "2021-08-22T14:44:14.11635+02:00",
        "display_name": "Shopping Cart Service",  // (4)
        "enabled": true,                          // (4)
        "id": 738,
        "updated_at": "2021-08-22T14:44:14.11722+02:00",
        "version_number": 1,
        "version_state": "approved"         // (3)
    }
}
~~~

 1. Non-versioned data is located in the top-level JSON object.
 2. Versioned data, as well as versioning state, is located in `latest_approved_version` (or in some other JSON outputs, `version`).
 3. The version state and proposal state are shown here.
 4. These are the versioned data.
