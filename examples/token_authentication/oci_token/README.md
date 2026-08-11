# OCI IAM token authentication example

This directory keeps the runnable Go sample in [main.go](C:/work/driver/go-driver/go-oracledb/examples/token_authentication/oci_token/main.go) and replaces the old setup scripts with a short guide for configuring Autonomous Database for OCI IAM authentication.

The Go driver expects:

- `TokenAuthentication = "OCI_TOKEN"`
- `TokenLocation` pointing to a directory that contains an OCI database token bundle
- a TCPS connect descriptor for the target Autonomous Database

By default, OCI CLI writes the token bundle to `~/.oci/db-token`. The driver looks for:

- `token`
- `oci_db_key.pem`

## 1. Prerequisites

You need:

- an Autonomous Database
- an OCI IAM user that should be allowed to connect
- OCI CLI installed and authenticated
- ADMIN access to the target database

You also need these identifiers:

- tenancy OCID
- compartment OCID
- Autonomous Database OCID
- IAM user name or IAM user OCID

## 2. Create IAM policy and group membership

The IAM user must be in a group that is allowed to connect to the Autonomous Database.

A narrow database-scoped policy looks like this:

```text
Allow group ADB_IAM_DB_USERS to use autonomous-database-family in compartment id <COMPARTMENT_OCID> where target.id = '<AUTONOMOUS_DATABASE_OCID>'
Allow group ADB_IAM_DB_USERS to use database-connections in compartment id <COMPARTMENT_OCID> where target.id = '<AUTONOMOUS_DATABASE_OCID>'
```

You can widen the scope to the whole compartment or tenancy if that is what your environment needs, but database scope is the safest default.

Then add the IAM user to that group.

## 3. Enable OCI IAM authentication in the database

Connect as `ADMIN` and enable OCI IAM external authentication:

```sql
BEGIN
  DBMS_CLOUD_ADMIN.ENABLE_EXTERNAL_AUTHENTICATION(
    type => 'OCI_IAM'
  );
END;
/
```

If another external authentication provider is already enabled and you intentionally want to replace it:

```sql
BEGIN
  DBMS_CLOUD_ADMIN.ENABLE_EXTERNAL_AUTHENTICATION(
    type  => 'OCI_IAM',
    force => TRUE
  );
END;
/
```

You can verify the setting with:

```sql
SELECT name, value
FROM v$parameter
WHERE name = 'identity_provider_type';
```

Expected value:

```text
OCI_IAM
```

## 4. Map the IAM principal to a database user

Create a globally identified database user for the IAM principal. For a direct user mapping:

```sql
CREATE USER IAM_USER IDENTIFIED GLOBALLY AS 'IAM_PRINCIPAL_NAME=<principal>';
GRANT CREATE SESSION TO IAM_USER;
```

If the user already exists:

```sql
ALTER USER IAM_USER IDENTIFIED GLOBALLY AS 'IAM_PRINCIPAL_NAME=<principal>';
GRANT CREATE SESSION TO IAM_USER;
```

The `<principal>` value is usually one of:

- `<user>`
- `<domain>/<user>` for non-default identity domains
- `<tenancy_ocid>:<domain>/<user>` for cross-tenancy cases

If you prefer group-based mapping instead of user-based mapping, use:

```sql
CREATE USER IAM_GROUP_USER IDENTIFIED GLOBALLY AS 'IAM_GROUP_NAME=<group>';
GRANT CREATE SESSION TO IAM_GROUP_USER;
```

## 5. Generate or refresh the database token

Use OCI CLI to generate a fresh database token bundle:

```bash
oci iam db-token get --db-token-location ~/.oci/db-token
```

On Windows:

```powershell
oci iam db-token get --db-token-location $env:USERPROFILE\.oci\db-token
```

This command writes the token and key files into the target directory. Re-run the same command whenever you want to refresh the token bundle.

If you want to limit the token to a specific database, compartment, or tenancy, pass `--scope`. For example:

```bash
oci iam db-token get \
  --db-token-location ~/.oci/db-token \
  --scope "urn:oracle:db::id::<COMPARTMENT_OR_DATABASE_SCOPE>"
```

## 6. Run the Go sample

The sample reads its configuration from these environment variables:

- `ORACLE_GO_OCI_TOKEN_CONNECT_DESCRIPTOR`
- `ORACLE_GO_OCI_TOKEN_LOCATION`

Set them before running [main.go](C:/work/driver/go-driver/go-oracledb/examples/token_authentication/oci_token/main.go).

PowerShell:

```powershell
$env:ORACLE_GO_OCI_TOKEN_CONNECT_DESCRIPTOR = "(description=(address=(protocol=tcps)(port=1522)(host=<adb-host>))(connect_data=(service_name=<service-name>))(security=(ssl_server_dn_match=yes)))"
$env:ORACLE_GO_OCI_TOKEN_LOCATION = "$env:USERPROFILE\.oci\db-token"
go run ./examples/token_authentication/oci_token
```

Bash:

```bash
export ORACLE_GO_OCI_TOKEN_CONNECT_DESCRIPTOR="(description=(address=(protocol=tcps)(port=1522)(host=<adb-host>))(connect_data=(service_name=<service-name>))(security=(ssl_server_dn_match=yes)))"
export ORACLE_GO_OCI_TOKEN_LOCATION="$HOME/.oci/db-token"
go run ./examples/token_authentication/oci_token
```

If the sample connects successfully, it prints:

```text
Query result: OK
```

## 7. Common checks when login fails

- Confirm the IAM user is in the expected IAM group.
- Confirm the policy includes `database-connections`.
- Confirm `identity_provider_type` is `OCI_IAM`.
- Confirm the database user is mapped with the correct `IAM_PRINCIPAL_NAME` or `IAM_GROUP_NAME`.
- Confirm the token bundle directory contains a fresh `token` and `oci_db_key.pem`.
- Confirm the connect descriptor uses TCPS and the correct ADB service name.

## References

This README is based on the OCI CLI and Autonomous Database documentation, especially:

- OCI CLI `oci iam db-token get`: <https://docs.oracle.com/en-us/iaas/tools/oci-cli/latest/oci_cli_docs/cmdref/iam/db-token/get.html>
- Enable IAM authentication on Autonomous Database: <https://docs.oracle.com/en-us/iaas/autonomous-database-shared/doc/enable-iam-authentication.html>
- `DBMS_CLOUD_ADMIN.ENABLE_EXTERNAL_AUTHENTICATION`: <https://docs.oracle.com/en-us/iaas/autonomous-database-serverless/doc/dbms-cloud-admin.html>
