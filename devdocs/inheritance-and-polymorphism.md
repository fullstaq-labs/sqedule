# Inheritance & polymorphism

## Inheritance via one-table-per-child

We implement inheritance in the database by using multiple tables, one per child concrete class. This is as opposed to single-table-inheritance. The reasons for choosing this strategy are:

 * This allows different child classes to have a different type of primary key. For example, User's primary key is Email, while ServiceAccount's primary key is Name.
 * Using multiple tables allows better use of foreign key constraints to guarantee data integrity. For example, suppose that a resource has a relation with a concrete subclass of another resource type. In single-table-inheritance, it's not possible to constrain the relationship to a specific subclass.

In diagrams, we mark multi-table inheritance with the annotation "[table per child]".

## Polymorphic associations

We implement polymorphic associations through the use of multiple foreign keys, one per possible concrete class. Each foreign key is nullable, and we specify a database CHECK constraint to ensure at most one (or exactly one, depending on the association's multiplicity) foreign key is set.

For example, let's take a look at CreationAuditRecord, which has a 0..1 reference to OrganizationMember (which can be either User or ServiceAccount):

~~~sql
CREATE TABLE creation_audit_records (
  ...

  -- Foreign key for User association
  user_email TEXT
    REFERENCES users (email),

  -- Foreign key for ServiceAccount association
  service_account_name TEXT
    REFERENCES service_accounts (name),

  -- Check constraint: there's either a User association, or ServiceAccount association, or either
  CHECK (
    (CASE WHEN user_email IS NULL THEN 0 ELSE 1 END)
    + (CASE WHEN service_account_name IS NULL THEN 0 ELSE 1 END)
    <= 1
  )
);
~~~
