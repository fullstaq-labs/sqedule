# Implicit multitenancy fields & relations

According to our [multitenancy strategy](multitenancy-strategy.md), almost all resources live in the context of an [Organization](organizations-users-service-accounts.md). These resources have an **implicit relation to Organization**, through an `OrganizationID` field. This implicit relation and this field are not shown in diagrams, and are not described in the development documentation besides the current document.

Usually, this `OrganizationID` field is composed into the resource's primary key as the first column. Example: the Application resource is documented in the diagrams as having the primary key `ID`. In actuality, its **primary key is a composite**: `(OrganizationID, ID)`.

Any **secondary indexes and unique constraints** are also within the context of an Organization. Example: ApprovalRulesetMajorVersion is documented in the diagrams as having a unique constraint on `(ApprovalRulesetID, VersionNumber)`. In actuality, its unique constraint is `(OrganizationID, ApprovalRulesetID, VersionNumber)`.
