# The Undeletable concept

Certain resources in the data model must not be deletable, because doing so may destroy history and thus inferere with audits. We say that these resources are Undeletable. This document describes what the behavioral implications are of this concept.

## Scope & real deletion

One must specify over what _scope_ a resource is undeletable. This means that real deletion is only allowed when the scope is deleted. This is the only exception under which Undeletable resources can be truly deleted.

Unless specified otherwise, the scope is the entire Organization that the resource belongs to.

## Foreign key constraints

Resources that reference an Undeletable resource, use the foreign key constraint `ON DELETE RESTRICT` to prevent accidental data deletion.

## See also

 * [The Disableable concept](disableable-concept.md)
