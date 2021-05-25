# Reviewable resources REST API pattern

This document describes a pattern for REST API endpoints for CRUD operations on [Reviewable resources](reviewable-concept.md). All Reviewable resources' REST API follow this pattern. These API endpoints allow performing all operations in the [change process diagram](reviewable-concept.md#the-version-creation-process).

## Operations on resources

### POST /resources

Creates a new Resource with one version. You may specify whether to immediately submit this version for approval, or to keep it as a proposal. The resulting resource version could either be a proposal or an approved version.

#### Input

The input is a `ResourceInput`:

~~~javascript
{
    // Optional:
    // Specify "proposal" to keep this version as a proposal,
    // "final" to submit it for approval, or "abandon" to
    // create a proposal in the abandoned state.
    "proposal_state": "proposal" (default) | "final" | "abandon" | null,

    // ...non-versioned resource-specific fields here...

    // A ResourceVersionInput object
    "version": {
        "comments": "..." | null,

        // ...versioned resource-specific fields here...
    }
}
~~~

#### Response

Response codes:

 * 201 Created

The response is a `ResourceWithVersion`:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "version": {
        "id": 123456,
        "version_state": "proposal" | "approved",
        "version_number": 1 | null,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": "2021-05-25T18:00:00+02:00" | null,

        // ...versioned resource-specific fields here...
    }
}
~~~

### GET /resources

Retrieves all resources and their latest approved versions.

Response codes:

 * 200 OK

~~~javascript
{
    "items": [
        // A ResourceWithLatestApprovedVersion object:
        {
            // ...non-versioned resource-specific fields here....

            "created_at": "2021-05-25T18:00:00+02:00",
            "updated_at": "2021-05-25T18:00:00+02:00",

            // A ResourceVersion object
            "latest_approved_version": null | {
                "id": 123456,
                "version_state": "approved",
                "version_number": 123456,
                "created_at": "2021-05-25T18:00:00+02:00",
                "updated_at": "2021-05-25T18:00:00+02:00",
                "approved_at": "2021-05-25T18:00:00+02:00",

                // ...versioned resource-specific fields here...
            }
        }
    ]
}
~~~

### GET /resources/:id

Retrieves the resource and its latest approved version.

Response codes:

 * 200 OK

The response is a `ResourceWithLatestApprovedVersion`:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "latest_approved_version": null | {
        "id": 123456,
        "version_state": "approved",
        "version_number": 123456,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": "2021-05-25T18:00:00+02:00",

        // ...versioned resource-specific fields here...
    }
}
~~~

### PATCH /resources/:id

Patches a resource. One may specify non-versioned fields to patch, or versioned fields to patch, or both.

In case versioned fields are specified, this endpoint creates a new version on top of the latest approved version, applying the patch. The resulting version could either be a proposal or an approved version.

If you want to modify a proposal, use `PATCH /resources/:id/proposals/:version_id` instead.

#### Input

The input is a `ResourceInput`:

~~~javascript
{
    // Optional:
    // Specify "proposal" to keep the patched version as a proposal,
    // or "final" to submit it for approval.
    "proposal_state": "proposal" (default) | "final" | null,

    // Optional:
    // ...non-versioned resource-specific fields here...

    // Optional:
    // A ResourceVersionInput object
    "version": null | {
        "comments": "..." | null,

        // ...versioned resource-specific fields here...
    }
}
~~~

#### Response

Response codes:

 * 200 OK

The response is a `ResourceWithVersion`:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "version": {
        "id": 123456,
        "version_state": "proposal" | "approved",
        "version_number": 1 | null,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": "2021-05-25T18:00:00+02:00" | null,

        // ...versioned resource-specific fields here...
    }
}
~~~

### DELETE /resources/:id

Deletes a resource. Only exists/allowed if the resource is deletable.

## Operations on approved versions

### GET /resources/:id/versions

Retrieves all versions of a resource, and their associated latest adjustments.

~~~javascript
{
    "items": [
        // A ResourceWithVersion object
        {
            // ...non-versioned resource-specific fields here....

            "created_at": "2021-05-25T18:00:00+02:00",
            "updated_at": "2021-05-25T18:00:00+02:00",

            // A ResourceVersion object
            "version": {
                "id": 123456,
                "version_state": "approved",
                "version_number": 1234,
                "created_at": "2021-05-25T18:00:00+02:00",
                "updated_at": "2021-05-25T18:00:00+02:00",
                "approved_at": "2021-05-25T18:00:00+02:00",

                // ...versioned resource-specific fields here...
            }
        }
    ]
}
~~~

### GET /resources/:id/versions/:number

Retrieves a specific version (and its associated latest adjustment) of a resource.

The response is a `ResourceWithVersion`:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "version": {
        "id": 123456,
        "version_state": "approved",
        "version_number": 1234,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": "2021-05-25T18:00:00+02:00",

        // ...versioned resource-specific fields here...
    }
}
~~~

## Operations on proposals

### GET /resources/:id/proposals

Retrieves all proposals for a given resource.

~~~javascript
{
    "items": [
        // A ResourceWithVersion object
        {
            // ...non-versioned resource-specific fields here....

            "created_at": "2021-05-25T18:00:00+02:00",
            "updated_at": "2021-05-25T18:00:00+02:00",

            // A ResourceVersion object
            "version": {
                "id": 123456,
                "version_state": "proposal",
                "version_number": null,
                "created_at": "2021-05-25T18:00:00+02:00",
                "updated_at": "2021-05-25T18:00:00+02:00",
                "approved_at": null,

                // ...versioned resource-specific fields here...
            }
        }
    }
}
~~~

### GET /resources/:id/proposals/:version_id

Retrieves a specific proposal for a given resource.

The response is a ResourceWithVersion:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "version": {
        "id": 123456,
        "version_state": "proposal",
        "version_number": null,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": null,

        // ...versioned resource-specific fields here...
    }
}
~~~

### PATCH /resources/:id/proposals/:version_id

Adjusts a proposal. You may specify whether to immediately submit this proposal for approval, or to keep it as a draft, or to abandon it. The resulting proposal could either still be a proposal, or become an approved version.

#### Input

The input is a `ResourceVersionInputWithState`:

~~~javascript
{
    // Optional:
    // Specify "proposal" to keep the patched proposal in the draft state,
    // "final" to submit it for approval, or "abandon" to abandon
    // the proposal.
    "proposal_state": "proposal" (default) | "final" | "abandon" | null,

    "comments": "..." | null,

    // ...versioned resource-specific fields here...
}
~~~

#### Response: 200 OK

Adjustment succeeded.

The response is a `ResourceWithVersion`:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "version": {
        "id": 123456,
        "version_state": "proposal" | "approved",
        "version_number": 123456 | null,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": "2021-05-25T18:00:00+02:00" | null,

        // ...versioned resource-specific fields here...
    }
}
~~~

#### Response: 422 Unprocessable entity

Adjustment is not allowed in the proposal's current state.

~~~javascript
{
    "error": "An error message"
}
~~~

### PUT /resources/:id/proposals/:version_id/review-state

Approves or reject a proposal's latest adjustment. This only works if it's in the "awaiting approval" state.

#### Input

~~~javascript
{
    "state": "approved" | "rejected"
}
~~~

#### Response: 200 OK

The response is a `ResourceWithVersion`:

~~~javascript
{
    // ...non-versioned resource-specific fields here....

    "created_at": "2021-05-25T18:00:00+02:00",
    "updated_at": "2021-05-25T18:00:00+02:00",

    // A ResourceVersion object
    "version": {
        "id": 123456,
        "version_state": "proposal" | "approved",
        "version_number": 123456 | null,
        "created_at": "2021-05-25T18:00:00+02:00",
        "updated_at": "2021-05-25T18:00:00+02:00",
        "approved_at": "2021-05-25T18:00:00+02:00" | null,

        // ...versioned resource-specific fields here...
    }
}
~~~

#### Response: 422 Unprocessable entity

Approving or rejecting the proposal is not allowed in the proposal's current state.

~~~javascript
{
    "error": "An error message"
}
~~~

### DELETE /resources/:id/proposals/:version_id

Deletes a proposal.
