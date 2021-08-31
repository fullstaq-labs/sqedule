# Release process

 1. Bump the version number in:

     - `cli/version.go`
     - `server/version.go`
     - `mkdocs.yml`

 2. Ensure [the CI](https://github.com/fullstaq-labs/sqedule/actions) is successful.

 3. [Manually run the "CI/CD" workflow](https://github.com/fullstaq-labs/sqedule/actions/workflows/ci-cd.yml). Set the `create_release` parameter to `true`. Wait until it finishes. This creates a draft release.

 4. Edit [the draft release](https://github.com/fullstaq-labs/sqedule/releases)'s notes and finalize the release.
