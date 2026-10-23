# DRCP example

This example shows how to use Oracle Database Resident Connection Pooling (DRCP) with the Go driver. 
It demonstrates both ordinary `database/sql` usage and explicit attachment to and detachment from the DRCP pool.

## Prerequisites.

- An Oracle Database service configured to use DRCP.
- A database username and password with permission to connect to that service.

### Enable DRCP on the remote server. 

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

To run the _customPool_ example you need to set up the pool MYPOOL using the following commands 

logon as sysdba on the remote server and issue the following commands (logged as SYSDBA)
```sql

exec dbms_connection_pool.add_pool('mypool')
exec dbms_connection_pool.configure_pool(pool_name => 'mypool',minsize => 4,maxsize => 40,incrsize => 2,session_cached_cursors => 20,inactivity_timeout => 300,max_think_time => 600,max_use_session => 500000,max_lifetime_session => 86400)
exec dbms_connection_pool.start_pool('mypool')
```

## Configure the connection.

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

## Run the example.

From this directory, run:

```sh
go run .
```

The program prints the connected database user if the connection and query succeed. It exits with an error if a required environment variable is missing or a database operation fails.

## What the program does?

The program creates a driver configuration with DRCP enabled by setting the server type to `pooled`. It also sets the connection class to `MyDRCPClass`; change this value if your application uses a different class.

It then runs two examples:

- **Basic usage:** Creates a connector, opens a `database/sql` pool with `sql.OpenDB`, and queries `SELECT USER FROM DUAL`.
- **Custom pool:** Sets the DRCP pool name  to `MyPool`, obtains a connection to perform operations.


### Monitor DRCP pool activity.

The following views can be used to see current DRCP statistics

```sql
SELECT cclass_name, pool_name, purity, tag, connection_status FROM v$cpool_conn_info;
SELECT pool_name, num_requests FROM V$CPOOL_STATS;
```

## References.

- [DRCP support](../documentation/drcp.md)

