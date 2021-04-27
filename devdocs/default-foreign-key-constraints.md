# Default foreign key constraints

Unless specified otherwise:

 * All foreign keys are set to `ON UPDATE CASCADE`.
 * Most foreign keys are set to `ON DELETE CASCADE`.
    - If the referenced resource is [disableable](disableable-trait.md) or [undeletable](undeletable-trait.md), then the foreign key is set to `ON DELETE RESTRICT`.
