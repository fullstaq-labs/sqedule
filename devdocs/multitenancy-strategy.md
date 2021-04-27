# Multitenancy strategy

Multitenancy is the ability to support multiple "organizations", "customers", "namespaces" or "tenants". We represent each tenant by an [Organization](organizations-users-service-accounts.md). Almost all resources in the data model live in the context of exactly one Organization.

We implement multitenancy as follows:

 * We serve multiple tenants from a single database.
 * All multitenant resources have an `OrganizationID` field.
 * All multitenant resources have a composite primary key, which includes `OrganizationID` as the first field.
 * All multitenant resources' secondary indexes and unique constraints are also composite, and include `OrganizationID` as the first field.

See also:

 * [Implicit multitenancy fields & relations](implicit-multitenancy-fields-and-relations.md)
 * [Multitenancy & security: foreign key constraints](multitenancy-security-foreign-key-constraints.md)

## Non-multitenant resources

Some resources do not live inside the context of an Organization. For example Organization itself. These resources are explicitly marked as such through the "[non-multitenant]" attribute.

