# Development documentation

## Getting started

 * [Development environment setup]
 * [Running the server]
 * [Loading database seeds]

## Design concepts

General:

 * [Technology stack]
 * [System architecture]
 * [Directory structure](directory-structure.md)
 * [Access control & roles](access-control-and-roles.md)

Data model domains:

 * [Organizations, users & service accounts](organizations-users-service-accounts.md)
 * [Applications, releases & release events](applications-releases-and-related.md)
 * [Approval rulesets, ruleset bindings, rules & rule outcomes](approval-rulesets-and-related.md)
 * [Creation audit records]

Data model core rules:

 * [Data model naming & casing conventions](data-model-naming-and-casing-conventions.md)
 * [Implicit multitenancy fields & relations](implicit-multitenancy-fields-and-relations.md)
 * [Non-nullable by default](non-nullable-by-default.md)
 * [Default foreign key constraints](default-foreign-key-constraints.md)

Data model core concepts:

 * [Multitenancy strategy](multitenancy-strategy.md)
 * [Multitenancy & security: foreign key constraints](multitenancy-security-foreign-key-constraints.md)
 * [Inheritance & polymorphism](inheritance-and-polymorphism.md)
 * [Undeletable resources] TODO
 * [Disableable resources] TODO
 * [The Immutability concept](immutability.md)
 * [The Reviewable concept: versioning, auditing & reviewing changes](reviewable-concept.md)

Database:

 * [Database migrations]()
 * [Enum types]()

## Processes & tasks

Coding:

 * [Changing the data model or the database schema]
 * [Adding a new Reviewable resource]

Collaboration:

 * [Code review checklist](code-review-checklist.md)
 * [Editing diagrams](editing-diagrams.md)

## Team handbook

Only applicable to core team members.

 * [Way of working]
