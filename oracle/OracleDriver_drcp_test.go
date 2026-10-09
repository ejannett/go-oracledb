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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/oracle/go-oracledb/v26/internal/common"
	oracleErrors "github.com/oracle/go-oracledb/v26/oracle/errors"
)

// TestDriver_SimpleConnection executes a simple connection.
func TestDriver_DRCPSimpleConnection(t *testing.T) {
	t.Parallel()
	if TestingConfig == nil {
		t.Skip("No configuration available")
	}
	fmt.Printf("connecting to %q\n", TestingConfig.GetConnectionString())
	db, err := openTestDBWithConfig(TestingConfig)
	if err != nil {
		t.Fatalf("failed to open connection : %v", err)
	}
	_, err = db.Query("SELECT 1 FROM DUAL")
	if err != nil {
		return
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("failed to open connection : %v", err)
	}
}

// TestDriver_DRCP_SelectDual executes a trivial SELECT to validate DRCP query flow.
func TestDriver_DRCP_SelectDual(t *testing.T) {
	t.Parallel()
	if TestingConfig == nil {
		t.Skip("No configuration available")
	}

	config := NewOracleDriverConfig()

	config.Credentials.LogonMode = TestingConfig.Credentials.LogonMode

	//config.Credentials.User = TestingConfig.Credentials.Username
	//config.Credentials.Password = TestingConfig.Credentials.Password

	config.ConnectionProperties.ServerType = common.ServerTypePooled
	config.ConnectionProperties.Drcp.Class = "TestDriver_DRCP_SelectDual"
	//config.ConnectionProperties.Drcp.Tag = "MANU"
	//config.ConnectionProperties.Drcp.Boundary = "STATEMENT"

	//config.ConnectDescriptor = TestingConfig.GetConnectionDSN()
	config.ConnectDescriptor = TestingConfig.Credentials.Username + "/" + TestingConfig.Credentials.Password + "@" + TestingConfig.Database.Host + ":" + strconv.Itoa(int(TestingConfig.Database.Port)) + "/" + TestingConfig.Database.ServiceName
	connector, err := NewOracleConnector(config)
	if err != nil {
		t.Fatalf("create connector failed: %v", err)
	}

	// conf to be used for request that valid current attachement
	// We can use the same as it confuses DRCP stats
	privConfig := config.Clone()
	privConfig.ConnectionProperties.ServerType = common.ServerTypeDedicated
	privConfig.ConnectionProperties.Drcp.Class = ""

	privConnector, err := NewOracleConnector(privConfig)
	if err != nil {
		t.Fatalf("create priv connector failed: %v", err)
	}

	db := sql.OpenDB(connector)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}()
	err = db.Ping()
	if err != nil {
		t.Fatalf("ping connection failed: %v", err)
	}

	privdb := sql.OpenDB(privConnector)
	defer func() {
		if err := privdb.Close(); err != nil {
			t.Errorf("close privdb: %v", err)
		}
	}()

	cnx, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("get dedicated sql connection failed: %v", err)
	}
	defer cnx.Close()

	oracleConnection, err := NewConnectionWrapper(cnx)
	if err != nil {
		t.Fatalf("create connection wrapper failed: %v", err)
	}

	// here we are already attached as going out of th pool attaches us.
	if !isDrcpAttached(t, privdb, config.ConnectionProperties.Drcp.Class) {
		t.Fatalf("Should be attached now")
	}

	rows, err := cnx.QueryContext(context.Background(), "SELECT 1 FROM DUAL")
	if err != nil {
		t.Fatalf("select from DUAL failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("no row returned from DUAL")
	}

	err = oracleConnection.DetachFromDrcpPool(context.Background())
	if err != nil {
		t.Fatalf("Can't detach from DRCP: %v", err)
	}
	// Have to wait a little bit , be sure that remote view is updated
	time.Sleep(1 * time.Second)
	if isDrcpAttached(t, privdb, config.ConnectionProperties.Drcp.Class) {
		t.Fatalf("Should be detached now")
	}
}

func isDrcpAttached(t *testing.T, db *sql.DB, expectedClass string) bool {
	t.Helper()

	var attachedClass sql.NullString
	var purity sql.NullString
	var tag sql.NullString
	var connection_status sql.NullString

	rows := db.QueryRowContext(
		context.Background(),
		`SELECT cclass_name, purity, tag, connection_status
		   FROM v$cpool_conn_info
		  WHERE UPPER(cclass_name) LIKE '%.' || UPPER(:1)`,
		expectedClass,
	)
	if rows.Err() != nil {
		if rows.Err().(oracleErrors.SQLError).ErrorCode() == string(oracleErrors.TableOrViewNotFound) {
			// requires "Grant SELECT ON SYS.V_$CPOOL_CONN_INFO"
			t.Skipf("cannot verify DRCP attachment (missing privileges?): %v", rows.Err())
		}
		if rows.Err() == sql.ErrNoRows {
			t.Logf("DRCP connection class %q was not found in v$cpool_conn_info", expectedClass)
			return false
		}
		t.Fatalf("query v$cpool_conn_info for attached DRCP class %q failed: %v", expectedClass, rows.Err())
	}

	err := rows.Scan(&attachedClass, &purity, &tag, &connection_status)
	if err != nil {
		t.Fatalf("scan v$cpool_conn_info for attached DRCP class failed: %v",err)
	}

	if connection_status.String != "ACTIVE" {
		return false
	}
	t.Logf("DRCP conn info class %q attached, purity %q, tag %q, status %q",
		attachedClass.String, purity.String, tag.String, connection_status.String)
	return true
}
