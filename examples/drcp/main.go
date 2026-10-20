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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/oracle/go-oracledb/v26/internal/common"
	"github.com/oracle/go-oracledb/v26/oracle"
	oracleconfig "github.com/oracle/go-oracledb/v26/oracle/config"
)

// Sample application that uses DRCP connections.
func main() {

	connectDescriptor := os.Getenv("ORACLE_DSN")
	user := os.Getenv("ORACLE_USER")
	password := os.Getenv("ORACLE_PASSWORD")
	if connectDescriptor == "" || user == "" || password == "" {
		log.Fatal("set ORACLE_DSN, ORACLE_USER, and ORACLE_PASSWORD")
	}

	// Build an explicit driver configuration for NewOracleConnector.
	cfg := oracle.NewOracleDriverConfig()
	cfg.ConnectDescriptor = connectDescriptor
	cfg.Credentials.User = user
	cfg.Credentials.Password = password

	// Set DRCP related properties
	//  Enable DRCP
	cfg.ConnectionProperties.ServerType = common.ServerTypePooled
	//  Sets the connection class
	cfg.ConnectionProperties.Drcp.Class = "MyDRCPClass"

	// Runs the basic example
	simpleRun(cfg)

	// Runs explicit attachment with tagging
	simpleExplicit(cfg)
}

func simpleExplicit(cfg *oracleconfig.OracleDriverConfig) {

	//  Uses tagging.
	cfg.ConnectionProperties.Drcp.Tag = "MyTag"

	// sql.OpenDB accepts the connector returned by the driver API.
	connector, err := oracle.NewOracleConnector(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Creates the connection pool
	db := sql.OpenDB(connector)
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// At this stage the connection is attached.

	// Run a minimal query using the established connection pool.
	//  All connections used by DB Apis will be attached and detached to and form the remote DRCP pool.
	var currentUser string
	if err := db.QueryRowContext(context.Background(),
		"select user from dual").Scan(&currentUser); err != nil {
		log.Fatal(err)
	}

	// Grab a connection wrapper to use the DRCP APIs
	wrapper, err := oracle.NewConnectionWrapper(conn)
	if err != nil {
		log.Fatal(err)
	}

	// Explicitly asked for attachment
	attachStatus, err := wrapper.AttachToDrcpPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if attachStatus.IsTagMismatch() {
		// The remote DRCP pool was not able to fulfill our request for "MyTag"
		// ...
	}

	// Explicitly asked for connection detachment
	err = wrapper.DetachFromDrcpPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Rest of application goes here.
	// ...
}

func simpleRun(config *oracleconfig.OracleDriverConfig) {

	// sql.OpenDB accepts the connector returned by the driver API.
	connector, err := oracle.NewOracleConnector(config)
	if err != nil {
		log.Fatal(err)
	}

	// Creates the connection pool
	db := sql.OpenDB(connector)
	defer db.Close()

	// Run a minimal query using the established connection pool.
	//  All connections used by DB Apis will be attached and detached to and form the remote DRCP pool.
	var currentUser string
	if err := db.QueryRowContext(context.Background(),
		"select user from dual").Scan(&currentUser); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to Oracle as %s\n", currentUser)
}
