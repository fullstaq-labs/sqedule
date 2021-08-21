# Sqedule — a release auditing & approval platform

Sqedule is an application release auditing & approval platform.

 - **Auditing**: Sqedule allows teams to have a central audit log of all released applications and their versions. This provides valuable information when troubleshooting complex application architectures that may involve many microservices or components.
 - **Approval**: Sqedule helps organizations that traditionally used formal (ITSM-based) change & release management processes, and are transitioning to adopt more DevOps-style continuous release processes.

   Sqedule helps such organizations implement a more restrictive CI/CD. Sqedule allows change & release managers to define release approval rules. Some of these rules are fully automated (for example: "only allow releasing in this time window"), others involve manual approvals from specific people or teams.

Sqedule works by integrating with CI/CD pipelines, so that all releases are logged into Sqedule. If rules are defined, then the CD pipeline only proceeds with releasing when all rules allow so.

Sqedule consists of:

 - An HTTP server with a JSON API.
 - A web interface (part of the HTTP server).
 - A CLI for interacting with the HTTP server.

## Why Sqedule as an approval platform

The transition of a formal (ITSM based) change & release management process towards a DevOps-style software delivery is very challenging for some organizations. We go from a carefully planned deployment every couple of weeks, to an automated process that deploys updated software multiple times per day without human intervention. Change and release managers tend to feel powerless and out-of-control during these transitions.

With Sqedule we are trying to bridge the gap between ITIL-style change/release management and CI/CD/DevOps processes, by automating the change approval processes centrally and allowing change & release managers to collaborate smoothly.

## Installation

 - [Installing the Sqedule server](SERVER-ADMIN-GUIDE.md#installation)
 - [Installing the Sqedule CLI](USERS-GUIDE.md#installating-the-cli)

## Documentation

 - [Sqedule server administration guide](SERVER-ADMIN-GUIDE.md)
 - [Sqedule users guide](USERS-GUIDE.md)

## Development & contribution

 * [Development & contribution guide](CONTRIBUTING.md) — helps you get started.
 * [Development documentation](devdocs/README.md) — documents design, architecture, processes and more.