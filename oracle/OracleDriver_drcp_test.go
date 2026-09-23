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

	"github.com/oracle/go-oracledb/v26/internal/common"
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
	defer db.Close()
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
	config.ConnectionProperties.Drcp.Boundary = "STATEMENT"

	//config.ConnectDescriptor = TestingConfig.GetConnectionDSN()
	config.ConnectDescriptor = TestingConfig.Credentials.Username + "/" + TestingConfig.Credentials.Password + "@" + TestingConfig.Database.Host + ":" + strconv.Itoa(int(TestingConfig.Database.Port)) + "/" + TestingConfig.Database.ServiceName
	connector, err := NewOracleConnector(config)
	if err != nil {
		t.Fatalf("create compressed connection connector: %v", err)
	}
	db := sql.OpenDB(connector)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close compressed connection: %v", err)
		}
	}()
	err = db.Ping()
	if err != nil {
		t.Fatalf("cannot connect %v", err)
	}
	rows, err := db.QueryContext(context.Background(), "SELECT 1 FROM DUAL")
	if err != nil {
		t.Fatalf("select from DUAL failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatalf("no row returned from DUAL")
	}
	var val int
	if err := rows.Scan(&val); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if val != 1 {
		t.Fatalf("unexpected value from DUAL: got %d, want 1", val)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
}
