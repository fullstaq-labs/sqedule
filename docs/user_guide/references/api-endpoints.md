# API endpoints

!!! info "See also"
    Concept: [API](../concepts/api.md)

## Base URL

All URL endpoints in this document are to be prefixed with the Sqedule server's base URL, plus the versioning prefix `/v1`. For example if this document speaks of `/releases`, then the API endpoint is `https://your-sqedule-server.com/v1/releases`.

## Authentication

There is currently no authentication because the server [does not support that yet](../../server_guide/concepts/security.md). But the server administrator is allowed to put the server behind a reverse proxy with HTTP Basic Authentication.

## Common error codes

 * 400 Bad Request — A path parameter or the input body has a syntax error.
 * 401 Unauthorized — Either an authentication failure, or the authenticated organization member is not authorized to perform this action.
 * 404 Not Found — Resource not found.
 * 500 Internal Server Error — Generic internal error.

## Types

Most types are self-explanatory, but these types deserve further explanation:

 * Timestamp — String in the format of "2021-05-07T12:44:59.105536+02:00".

## Releases

### Create release

~~~
POST /applications/:application_id/releases
~~~

Path parameters:

 * `application_id` — ID of the application to create a release for.

Input body:

~~~javascript
{
  /****** Optional fields ******/

  // Arbitrary string that identifies the source location of this application,
  // for example the Git repo URL. Currently only used for display purposes,
  // but planned to have a semantic meaning in the future.
  "source_identity": string,

  // Arbitrary metadata to include in this release.
  "metadata": object,

  // Arbitrary comments to include in this release.
  "comments": string,
}
~~~

Response codes:

 * 201 Created — Creation success.

### List releases

~~~
GET /applications/:application_id/releases
~~~

Path parameters:

 * `application_id` — ID of the application to list releases for.

### Get a release

~~~
GET /applications/:application_id/releases/:id
~~~

Path parameters:

 * `application_id` — ID of the application to read a release for.
 * `id` — ID of the release to read.

Output body:

~~~json
{
  "id": number,
  "state": "pending" | "approved" | "rejected",
  "source_identity": string | null,
  "metadata": object,
  "comments": string | null,
  "created_at": timestamp,
  "updated_at": timestamp,
  "finalized_at": timestamp | null,
  "approval_ruleset_bindings": [array of Release Approval Ruleset Bindings]
}
~~~
