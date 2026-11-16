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
	"strings"
	"testing"
)

// TestcheckDecimalScale reads a stored 12.34 back through several expressions. Oracle reports scale 0
// for an expression over a scaled NUMBER, and the driver then discards the fraction.
func TestCheckDecimalScale(t *testing.T) {
	t.Parallel()
	if TestingConfig == nil {
		t.Skip("No configuration available")
	}

	db, err := openTestDBWithConfig(TestingConfig)
	if err != nil {
		t.Fatalf("failed to open connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tableName := createObjectName("t_decimal_scale")

	// Create table with RAW column
	cols := map[string]string{
		"id":   "NUMBER(10) PRIMARY KEY",
		"amt": "NUMBER(22,2)",
	}
	if err := createTable(ctx, db, tableName, cols); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	defer dropTable(ctx, db, tableName)


	_, err = db.ExecContext(ctx, `INSERT INTO ` + tableName + ` VALUES (1, 12.34)`)
	if err != nil {
		t.Fatalf("failed to prepare insert: %v", err)
	}

	statements := []string{
		`SELECT amt FROM ` + tableName,
		`SELECT COALESCE(amt, 0) FROM ` + tableName,
		`SELECT CASE WHEN id = 1 THEN amt ELSE 0 END FROM ` + tableName,
		`SELECT amt + 0 FROM ` + tableName,
		`SELECT NULLIF(amt, 0) FROM ` + tableName,
		`SELECT CAST(COALESCE(amt, 0) AS NUMBER(22,2)) FROM ` + tableName,
	}

	var failures []string
	for _, stmt := range statements {
		p := probe(db,t, stmt)
		switch {
		case !p.ok():
			p.printFailure(t, "")
			failures = append(failures, stmt)
		case fmt.Sprintf("%v", p.value) != "12.34":
			p.printFailure(t,"want 12.34")
			failures = append(failures, stmt)
		}
	}
if len(failures) > 0 {
	t.Fatalf("failed to check decimal scale: %v", failures)
}

}
// printFailure prints one failing statement, its result, and the column metadata that selected the
// driver's decoder. want describes the correct result when the statement returned a wrong value
// rather than an error.
func (p probeResult) printFailure(t *testing.T , want string) {
	t.Logf("\n  FAIL  %s\n", p.sql)
	if p.scanned {
		detail := fmt.Sprintf("got %v (Go type %s)", p.value, p.goType)
		if want != "" {
			detail += ", " + want
		}
		t.Logf("        %s\n", detail)
	} else {
		t.Logf("        %s\n", p.err)
	}
	if p.hasMeta {
		t.Logf("        column reported as %s precision=%d scale=%d\n", p.typeName, p.precision, p.scale)
	}
}

func probe(db *sql.DB, t *testing.T , query string) probeResult {
	p := probeResult{sql: query}

	rows, err := db.Query(query)
	if err != nil {
		p.err = oneLine(err)
		return p
	}
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		p.err = "reading the column metadata failed: " + oneLine(err)
		return p
	}
	p.hasMeta = true
	p.typeName = colTypes[0].DatabaseTypeName()
	p.precision, p.scale, _ = colTypes[0].DecimalSize()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			p.err = oneLine(err)
			return p
		}
		p.err = "no rows returned"
		return p
	}

	var value any
	if err := rows.Scan(&value); err != nil {
		p.err = "scan failed: " + oneLine(err)
		return p
	}
	p.scanned = true
	p.value = value
	p.goType = fmt.Sprintf("%T", value)
	return p
}

// oneLine collapses a driver error onto a single line, because several of them end in a newline.
func oneLine(err error) string {
	return strings.Join(strings.Fields(err.Error()), " ")
}
// probeResult is one statement's outcome.
type probeResult struct {
	sql       string
	hasMeta   bool // false when the statement failed before any column metadata could be read
	typeName  string
	precision int64
	scale     int64
	scanned   bool
	value     any
	goType    string
	err       string
}
func (p probeResult) ok() bool { return p.scanned }

// checkNumberMantissa reads back NUMBER values from 15 to 38 digits. Oracle allows 38 significant
// digits; the driver decodes the mantissa into a uint64, which holds about 20.
func TestCheckNumberMantissa(t *testing.T)  {
	tableName := createObjectName("t_decimal_mantissa")
	const minDigits, maxDigits = 15, 38

	t.Logf("One unconstrained NUMBER column holding %d to %d nines, each read back on its own.\n",
		minDigits, maxDigits)
	t.Logf("Expected: every width returns its value, since Oracle NUMBER allows 38 digits.")

	t.Parallel()
	if TestingConfig == nil {
		t.Skip("No configuration available")
	}

	db, err := openTestDBWithConfig(TestingConfig)
	if err != nil {
		t.Fatalf("failed to open connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create table with RAW column
	cols := map[string]string{
		"digits":   "NUMBER(2)",
		"n": "NUMBER",
	}
	if err := createTable(ctx, db, tableName, cols); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	defer dropTable(ctx, db, tableName)

	for d := minDigits; d <= maxDigits; d++ {
		insert := fmt.Sprintf(`INSERT INTO %s (digits, n) VALUES (%d, %s)`, tableName, d, strings.Repeat("9", d))
		if _, err := db.Exec(insert); err != nil {
			t.Fatalf("insertion failed: %v", err)
		}
	}

	var failures []int
	for d := minDigits; d <= maxDigits; d++ {
		p := probe(db,t, fmt.Sprintf(`SELECT n FROM %s WHERE digits = %d`, tableName, d))
		if !p.ok() {
			// Print only the first failure. Once the mantissa overflows, every wider value fails the
			// same way, and 20 identical blocks say nothing the boundary does not.
			if len(failures) == 0 {
				p.printFailure(t, fmt.Sprintf("%d digit value", d))
			}
			failures = append(failures, d)
		}
	}

	if len(failures) > 0 {
		t.Fatalf("failed to check decimal mantissa: %v", failures)
	}
}
