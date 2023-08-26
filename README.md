# pg-zerotrust

`pg-zerotrust` is a proxy that sits in front of your PostgresSQL server and enforces a zero-trust access model. Specifically, `pg-zerotrust` performs the following activities:

- Provisions just-in-time users based on claims from a configured Identity Provider (IdP)
- Dynamically assigns privileges based on data contained inside IdP claims
- Cleans up just-in-time users once IdP-provided credentials expire

> [!IMPORTANT]
> `pg-zerotrust` is intended for use by humans accessing a database, not for programmatic access by applications.

## Why `pg-zerotrust`?

PostgreSQL provides many different authentication options that remove the need to set static passwords for user (e.g., LDAP, RADIUS). However, these options fall short in two ways:

1. **The user must be present in the PostgreSQL database.** Individual users must still be manually created inside the PostgreSQL user database before they can access the database, and they must be manually removed when the user should no longer have access.
2. **Users are assigned static privileges.** User privileges are assigned in the PostegreSQL database, which gives the user standing privileges to access the database. This prevents advanced access scenarios such as policy-based auth
