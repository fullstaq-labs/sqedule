# Applications, releases & release events

![](applications-releases-and-related.drawio.svg)

## Release background jobs

A ReleaseBackgroundJob represents a Release which is in need of being processed by the Approval Rules Engine. A Release has a ReleaseBackgroundJob, if and only if that Release is in the `in_progress` state.

ReleaseBackgroundJob mainly exists to allow a Release to be locked during Approval Rules Engine processing. This locking prevents two concurrent instances of the Approval Rules Engine from processing the same Release.

We use [PostgreSQL advisory locks](https://www.postgresql.org/docs/13/explicit-locking.html). Advisory locks are identified by 64-bit unsigned IDs, and are unique on a per-database level. In order to give each Release a unique advisory lock ID, we generate and store a unique sub-ID in the corresponding ReleaseBackgroundJob. This sub-ID is 31-bit unsigned and is also unique on a database-wide level (not just on the Organization level). The final advisory lock ID is calculated as `ReleaseBackgroundJobPostgresLockNamespace + SubID`.

The usage of a 31-bit unsigned sub-ID means that, in the context of Release locks, we don't reserve of the PostgreSQL advisory lock ID namespace. Thus, there is still space to use PostgreSQL advisory locks for other purposes.

It also means that at most `2^31-1 = 2147483647` unprocessed Releases may exist in a database at any given time.
