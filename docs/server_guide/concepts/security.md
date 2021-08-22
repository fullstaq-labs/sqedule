# Security

The Sqedule HTTP server is **unprotected**. There is currently no built-in support for authentication. Therefore you should not expose it to the Internet directly. Instead, you should put it behind an HTTP reverse proxy that supports authentication, for example Nginx with HTTP basic authentication enabled.

Support for user accounts, and thus authentication, is planned for a future release.
