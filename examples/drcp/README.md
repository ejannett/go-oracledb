# DRCP example

This example shows how to use Oracle Database Resident Connection Pooling (DRCP) with the Go driver. 
It demonstrates both ordinary `database/sql` usage and explicit attachment to and detachment from the DRCP pool.

## Prerequisites

- An Oracle Database service configured to use DRCP.
- A database username and password with permission to connect to that service.

### Enable DRCP on the remote server 

At least one DRCP pool should be started on the remote server.

Start the default DRCP pool (SYS_DEFAULT_CONNECTION_POOL)
```sql
sqlplus /nolog
SQL> connect / as sysdba
SQL> execute dbms_connection_pool.start_pool()
```

Start a named DRCP pool (SYS_DEFAULT_CONNECTION_POOL)

```sql
sqlplus /nolog
SQL> connect / as sysdba
SQL> execute dbms_connection_pool.start_pool('my_pool')
```

## Configure the connection

Set these environment variables:

- `ORACLE_DSN`: an Oracle connect descriptor or Easy Connect string for the DRCP-enabled service.
- `ORACLE_USER`: the database username.
- `ORACLE_PASSWORD`: the database password.

For example, in macOS or Linux:

```sh
export ORACLE_DSN='tcp://127.0.0.1:1521/freepdb1'
export ORACLE_USER='myuser'
export ORACLE_PASSWORD='mypassword'
```

Use values for your database. Keep credentials out of source control.

## Run the example

From this directory, run:

```sh
go run .
```

The program prints the connected database user if the connection and query succeed. It exits with an error if a required environment variable is missing or a database operation fails.

## What the program does

The program creates a driver configuration with DRCP enabled by setting the server type to `pooled`. It also sets the connection class to `MyDRCPClass`; change this value if your application uses a different class.

It then runs two examples:

- **Basic usage:** Creates a connector, opens a `database/sql` pool with `sql.OpenDB`, and queries `SELECT USER FROM DUAL`.
- **Explicit attachment:** Sets the DRCP tag to `MyTag`, obtains a connection, and uses `oracle.NewConnectionWrapper` to explicitly attach to and detach from the DRCP pool.

The class and tag values are sample values. Adjust them to suit your application and DRCP configuration.

## Reference

- [Oracle Database Resident Connection Pooling](https://www.oracle.com/a/tech/docs/drcp-technical-brief.pdf)
