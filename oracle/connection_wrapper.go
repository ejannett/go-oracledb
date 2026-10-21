/*
** Copyright (c) 2026 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively the "Software"), free of charge and under any and all copyright
** rights in the Software, and any and all patent rights owned or freely
** licensable by each licensor hereunder covering either (i) the unmodified
** Software as contributed to or provided by such licensor, or (ii) the Larger
** Works (as defined below), to deal in both
**
** (a) the Software, and
** (b) any piece of software and/or hardware listed in the lrgrwrks.txt file if
** one is included with the Software (each a "Larger Work" to which the Software
** is contributed by such licensors),
**
** without restriction, including without limitation the rights to copy, create
** derivative works of, display, perform, and distribute the Software and make,
** use, sell, offer for sale, import, export, have made, and have sold the
** Software and the Larger Work(s), and to sublicense the foregoing rights on
** either these or other terms.
**
** This license is subject to the following condition:
** The above copyright notice and either this complete permission notice or at
** a minimum a reference to the UPL must be included in all copies or
** substantial portions of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
 */

package oracle

import (
	"context"
	"database/sql"

	"github.com/oracle/go-oracledb/v26/internal/common"
	driverCommon "github.com/oracle/go-oracledb/v26/internal/driver/common"
	oracleErrors "github.com/oracle/go-oracledb/v26/oracle/errors"
)

// connectionWrapper provides Oracle specific operations for a dedicated
// database/sql connection.
//
// The wrapped connection must be a connection returned by this driver.
type connectionWrapper struct {
	connection *sql.Conn
}

// NewConnectionWrapper validates and wraps a dedicated database/sql connection
// for Oracle specific operations.
//
// Parameters:
//   - connection: Dedicated database/sql connection to wrap.
//
// Returns:
//   - *connectionWrapper: Wrapper for the supplied connection.
//   - error: Error if the underlying driver connection type is not supported.
func NewConnectionWrapper(connection *sql.Conn) (*connectionWrapper, error) {
	var wrapper *connectionWrapper
	err := connection.Raw(func(c any) error {
		// Include here all functions/interfaces we want a connection to implement in
		// order to be wrapped by this wrapper
		type canBeWrapped interface {
			driverCommon.DRCPUser
		}
		_, ok := c.(canBeWrapped)
		if !ok {
			return common.NewOracleError(oracleErrors.InvalidConnection, nil)
		}
		wrapper = &connectionWrapper{connection: connection}
		return nil
	})
	return wrapper, err
}

type DRCPAttachmentStatus int

const (
	// tagMismatch the acquired session did not fully match the requested tag.
	//   This is a result status, not necessarily an acquisition failure:
	//   the server can return a new or otherwise acceptable session and still report the mismatch.
	tagMismatch = 0x00000001
	// purityNew the server created a new database session for this acquire instead of reusing the previous one
	purityNew = 0x00000002
	// sessionChanged tells the client that the server-side database session associated with its DRCP connection
	//   has changed—or may be different on its next checkout—so client state tied to
	//  the prior session must be invalidated. In particular, the server source describes
	//  clearing the client’s statement cache to avoid using stale cursors.
	sessionChanged = 0x00000004
	// stmtCacheDestroy when paired with SESSGET_SESSION_CHANGED,
	//   it tells OCI to purge the statement cache and discard cached cursor IDs
	//   because they belong to the prior session
	stmtCacheDestroy = 0x00000008
)

// IsTagMismatch reports whether the acquired DRCP session did not fully match
// the requested tag. A tag mismatch does not necessarily mean attachment failed.
func (s DRCPAttachmentStatus) IsTagMismatch() bool {
	return (s & tagMismatch) > 0
}

// IsSessionChanged reports whether the database session associated with the
// DRCP connection changed. Client-side state associated with the prior session
// may need to be invalidated.
func (s DRCPAttachmentStatus) IsSessionChanged() bool {
	return (s & sessionChanged) > 0
}

// IsPurityNew reports whether DRCP created a new database session instead of
// reusing an existing session.
func (s DRCPAttachmentStatus) IsPurityNew() bool {
	return (s & purityNew) > 0
}

// IsStmtCacheDestroy reports whether DRCP indicates that the statement cache
// should be purged. When this status is set with IsSessionChanged, cached
// cursor IDs belong to the prior session and should be discarded.
func (s DRCPAttachmentStatus) IsStmtCacheDestroy() bool {
	return (s & stmtCacheDestroy) > 0
}

// AttachToDrcpPool attaches the connection to its configured DRCP pool,
// using ctx for the operation. The returned status describes the attachment
// outcome, including tag mismatch and session changes.
//
// It returns an error if DRCP is disabled, implicit pooling is enabled,
// the connection is already attached, or attachment fails.
func (c *connectionWrapper) AttachToDrcpPool(ctx context.Context) (DRCPAttachmentStatus, error) {
	var status DRCPAttachmentStatus = 0
	error := c.connection.Raw(func(c any) error {
		oc, _ := c.(driverCommon.DRCPUser)
		s, err := oc.AttachToResidentPool(ctx)
		status = DRCPAttachmentStatus(s)
		return err
	})
	return status, error
}

// DetachFromDrcpPool releases the attached connection to the DRCP pool,
// returns an error if DRCP is disabled,
// the connection is already detached, or release fails.
func (c *connectionWrapper) DetachFromDrcpPool(ctx context.Context) error {
	return c.connection.Raw(func(c any) error {
		oc, _ := c.(driverCommon.DRCPUser)
		return oc.DetachFromResidentPool(ctx)
	})

}

// IsAttached reports whether the connection is attached to a DRCP session.
// It returns true if the underlying connection is attached false otherwise.
func (c *connectionWrapper) IsAttached() bool {
	var attachmentStatus bool = false
	c.connection.Raw(func(c any) error {
		oc, _ := c.(driverCommon.DRCPUser)
		b := oc.IsAttachedToResidentPool()
		attachmentStatus = b
		return nil
	})
	return attachmentStatus
}
