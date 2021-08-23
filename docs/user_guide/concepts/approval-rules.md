# Approval rules

When a [release](applications-releases.md) is created, Sqedule evaluates rules on that release. Depending on the outcomes of these evaluations, Sqedule may approve or reject a release.

Rules are grouped in **rulesets**.

Rulesets must be **bound** to applications. When a release is created, Sqedule only approves the rules for rulesets bound to the corresponding application.

A ruleset binding can be in one of two modes:

 - **Enforcing** — If a rule evaluates as failed, then the release is rejected.
 - **Permissive** — If a rule evaluates as failed, then this failure is logged, but the release is not rejected.

Sqedule rejects a release if at least one rule evaluates as failed, and that rule is bound in enforcing mode. Otherwise, the release is approved. This means that if there are no rules bound to an application, then new releases are always approved.

<figure>
  <img src="../approval-rules.drawio.svg" alt="Diagram of how applications are bound to rulesets, and how rulesets group multiple rules">
  <figcaption>Rulesets group multiple rules. Applications are bound to rulesets via bindings. A binding can be in <em>enforcing</em> or <em>permissive</em> mode.</figcaption>
</figure>

## Rule types

At present, Sqedule only supports one type of rule: the so-called "schedule rule". More rule types will be supported in the future as Sqedule receives further development.

## Schedule rules

A Schedule rule defines a time-of-day window in which releases are allowed.
