# Database development seed data organization members

The [development seed data](db-dev-seeds-load.md) contains two organizations:

 * `org1` — This is where most of the seed data belongs to and what you should normally use when developing.
 * `org2` — Contains data for benchmarking queries.

Org1 contains two service accounts organization members:

 * Name: `admin_sa`
    - This is the service account you should normally use when developing.
    - Password: 123456
    - Role: admin
 * Name: `org_admin_sa`
    - Password: 123456
    - Role: org admin
