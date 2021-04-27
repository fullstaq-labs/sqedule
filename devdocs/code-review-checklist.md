# Code review checklist

 * [ ] Documentation updated
 * [ ] Test suite updated
 * [ ] If database models changed:
    - [ ] PR includes a corresponding database migration
 * [ ] If database migrations changed:
    - [ ] Migrations do not import any other Sqedule packages
    - [ ] Migrations do not destroy data
    - [ ] Primary keys include OrganizationID as first column
    - [ ] Secondary indexes and unique constraints include OrganizationID as first column
    - [ ] Foreign key constraints follow the documented rules for `ON UPDATE`/`ON CASCADE`
    - [ ] Diagrams updated
