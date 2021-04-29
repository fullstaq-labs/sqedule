# Access control and roles

We control access through the use of predefined roles. Each [organization member](organizations-users-service-accounts.md) has exactly one role. The roles and their capabilities are as follows:

| Capability                            | Org admin   | Admin | Change manager | Technician | Viewer    |
|---------------------------------------|-------------|-------|----------------|------------|-----------|
| Create organization                   | ✔︎           |       |                |            |           |
| Read organization                     | ✔︎           | Self  | Self           | Self       | Self      |
| Edit organization                     | ✔︎           | Self  |                |            |           |
| Delete organization                   | ✔︎           |       |                |            |           |
| Create organization member            | (1)         | ✔︎     |                |            |           |
| Edit organization member              |             | ✔︎     | Self           | Self       | Self      |
| Delete organization member            |             | ✔︎     |                |            |           |
| Propose application & ruleset changes |             | ✔︎     | ✔︎              | ✔︎          |           |
| Approve application & ruleset changes |             | ✔︎     | ✔︎              |            |           |
| Create release                        |             | ✔︎     | ✔︎              | ✔︎          |           |
| Cancel release                        |             | ✔︎     | ✔︎              | ✔︎          |           |
| Approving "manual approval" rule      |             | ✔︎     | ✔︎              |            |           |

 * (1) = Only if organization has no admins
