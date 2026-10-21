

# DRCP Support

API for DRCP support in the Go driver.

Reference: https://www.oracle.com/a/tech/docs/oracle-drcp-technical-brief.pdf

DRCP (Database Resident Connection Pooling) is a connection pool built into Oracle Database.
It keeps a shared pool of database server processes and sessions that applications can use temporarily, then return to the pool when finished.
This helps applications reduce memory and process overhead and improves database scalability.
DRCP can also complement a connection pool in the application’s middle tier.

A client connects through Oracle’s connection broker. When it needs database access, the broker assigns an available pooled server, which handles the request.
When the client releases the connection, the server process returns to the pool for reuse. The session may remain intact but become temporarily disconnected from that server, allowing it to be reused according to the application’s configuration.

The connection class determines which requests can share pooled sessions. Purity indicates whether the client can reuse a session (SELF) or wants a new one (NEW).

## Configuration

DRCP relies on a few connection parameters. These parameters are available as part of the driver configuration.

### Pool Configuration

```go
type DatabaseResidentConnectionPooling struct {
 ClassName         string // sets the DRCP connection class, this is optional when not specified default SYS.SHARED is used
 Purity            string // sets the DRCP purity
 Boundary          string // sets the DRCP boundary
 Tag               string // sets the tag for pooled DRCP connection
 UseMultipleTag    bool   // Does the pooled DRCP connection use multiple tags
 FixUpCallBackName string // Name of the PL/SQL callback used to transform a session checked out from the pool to the desired state requested by the application
 PoolName          string // Name of the DRCP pool to be used.
}
```

**ClassName:** The DRCP connection class. The connection class defines a logical name for the type of connection an application wants to use and share across multiple application processes or other applications.

**Purity:** Session purity. This attribute controls the reusability of pooled sessions in DRCP. It tells DRCP whether to provide a fresh session (NEW) or allow reuse of an existing pooled session (SELF).
Reusing a session can be more efficient, but applications should account for its session state.

**Boundary:** Implicit connection pooling. There are two possible values:

- `STATEMENT` – DRCP performs an “implicit release” when the database session is stateless.
- `TRANSACTION` – DRCP performs an “implicit release” at commit or rollback, or when the database session is stateless.

**Tag:** Free-form tag for pooled connections.

**UseMultipleTag:** Indicates whether the client uses multiple tags.

**FixUpCallBackName:** Name of the PL/SQL callback used to transform a session checked out from the pool

**PoolName:** The name of the DRCP pool to be used. The default pool (SYS_DEFAULT_CONNECTION_POOL) is used when no pool name is specified.

### Server Type Configuration

The client can specify the type of remote server process to be used. The type can be shared, dedicated, or pooled.

For DRCP to be used, the type must be set to `POOLED`.

There will be a new property within `OracleConnectionProperties`:

```text
ServerType = POOLED | SHARED | DEDICATED
```

## APIs

### Pool Enablement

In order to enable the use of DRCP the server type set in `OracleConnectionProperties.ServerType` must
be set to `POOLED`.

### Attachment Points

The connection must be attached to the pool in order to perform operations and detached to release remote
resources when the connection becomes idle. 

There are two scenarii: when the connection is used directly by the client application and when it is used implicitly by the `sql.DB` object.

#### Implicit Usage

This is when the connection is handled by the Go connection pool, or when `ExecContext` and `QueryContext` are executed against the `sql.DB` object.

In such situations, the driver attaches and detaches the connection automatically when the connection is respectively borrowed from and returned to the pool.
When ExecerContext and QueryerContext methods are used on the DB object, the connection lifecycle is transparent to the end user 

#### Explicit Usage

This is when the connection is handled by the client application.
The connection wrapper is then used to attach and detach the connection to remote DRCP pool 

The wrapper implements the following interface.

```go

type DRCPAttachmentStatus int

// IsTagMismatch reports whether the acquired DRCP session did not fully match
// the requested tag. A tag mismatch does not necessarily mean attachment failed.
func (s DRCPAttachmentStatus) IsTagMismatch() bool 

// IsSessionChanged reports whether the database session associated with the
// DRCP connection changed. Client-side state associated with the prior session
// may need to be invalidated.
func (s DRCPAttachmentStatus) IsSessionChanged() bool

// IsPurityNew reports whether DRCP created a new database session instead of
// reusing an existing session.
func (s DRCPAttachmentStatus) IsPurityNew() bool

// IsStmtCacheDestroy reports whether DRCP indicates that the statement cache
// should be purged. When this status is set with IsSessionChanged, cached
// cursor IDs belong to the prior session and should be discarded.
func (s DRCPAttachmentStatus) IsStmtCacheDestroy() bool

type DRCPUser interface {
	// Attach the connection to the remote pool.
	// Returns attachment status.
	// Errors: OGD-00180 / OGD-00181 (see below )
	AttachToDrcpPool(ctx context.Context) (DRCPAttachmentStatus, error)
	
	// Detach the connection from the remote pool. 
	//Errors: OGD-00180 / OGD-00181 
    DetachFromDrcpPool() error

	// Checks that current underlying connection is attached
    IsAttached() bool
}
```

Example usage:

```go
db  := sql.OpenDB(NewOracleConnector(cfg))
// ...
conn := db.Conn(context.Background())
// ...
connectionWrapper, err := oracle.NewConnectionWrapper(conn)
// ...
if !connectionWrapper.IsIsAttached() {
   status, err := connectionWrapper.AttachToResidentPool(context.Background())
   /...
}
conn.QueryContext(...)
err := connectionWrapper.DetachFromResidentPool()
```


### Errors

`OGD-00180` – DRCP is not enabled: an attempt was made to perform an operation that requires DRCP to be enabled.

Example: calling the attach or detach method.

`OGD-00181` – Invalid state: an attempt was made to perform an operation that requires the connection to be attached or detached.

Example: calling `QueryContext` on a detached connection.

