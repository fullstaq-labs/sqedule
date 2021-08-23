# Applications & releases

One defines **applications** in Sqedule. Sqedule then tracks the **releases** that have occurred for those applications.

 * An "application" in Sqedule represents an entity for which releases can exist. So it doesn't have to be an actual application: it could be a part of an application, it could be a logical group of applications, or it could be anything else as long as it's releasable.
 * A "release" tracks a single release. When a releaser (for example a CD pipeline) wants start a release process, it first registers a new release in Sqedule. This release starts in the _"pending"_ state. Sqedule then optionally evaluates [approval rules](approval-rules.md) on this release. Depending on the outcomes of the rule evaluations, Sqedule may _approve or reject_ this release. When the releaser notices this state change, it either aborts or proceeds with the release process.

<figure>
  <img src="../applications-releases.drawio.svg" alt="Diagrams showing the relationship between applications, releases and the CD pipeline">
  <figcaption>An application groups multiple releases. Each release may be in the "pending", "approved" or "rejected" state. The CD pipeline registers a release, and waits until it's approved or rejected before proceeding.</figcaption>
</figure>
