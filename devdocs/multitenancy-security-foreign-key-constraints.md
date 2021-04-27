# Multitenancy & security: foreign key constraints

Our [multitenancy strategy](multitenancy-strategy.md) involves storing multiple tenants in a single database. There's a big security risk with this approach: one tenant could accidentally access the data of another tenant because of a bug in the application code.

In order to prevent this from happening, we ensure that all primary keys and foreign keys share the same `OrganizationID` field. This makes it impossible for resources belonging to one tenant, to reference resources belonging to another tenant.

For example, suppose we have a Users resource:

 * Resource: Users
    - Field: OrganizationID
    - Field: Email
    - Primary key: (OrganizationID, Email)

✅ Good use of foreign key constraints (only allows referencing Users from the same Organization):

 * Resource Applications
    - Field: OrganizationID
    - Field: ID
    - Field: OwnerEmail
    - Primary key: (OrganizationID, ID)
    - Foreign key: (OrganizationID, OwnerEmail) references Users (OrganizationID, Email)

❌ Bad use of foreign key constraints (allows referencing Users from another Organization):

* Resource Applications
    - Field: OrganizationID
    - Field: ID
    - Field: OwnerOrganizationID
    - Field: OwnerEmail
    - Primary key: (OrganizationID, ID)
    - Foreign key: (OwnerOrganizationID, OwnerEmail) references Users (OrganizationID, Email)
