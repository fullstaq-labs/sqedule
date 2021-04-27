# Data model naming & casing conventions

## Development docs and Go code

Resource type names (resource classes) and resource field names are in singular format, formatted with CamelCase. The first letter is capitalized. Examples:

 - User
 - ServiceAccount
 - OrganizationID
 - DisplayName

## Database tables

Table names are in plural format, formatted with snake\_case. Examples:

 - users
 - service\_accounts

Column names are formatted with snake\_case. Example:

 - organization\_id
 - display\_name
