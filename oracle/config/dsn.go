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

package config

import (
	"strings"

	"github.com/oracle/go-oracledb/v26/internal/common"
	driverCommon "github.com/oracle/go-oracledb/v26/internal/driver/common"
	oracleErrors "github.com/oracle/go-oracledb/v26/oracle/errors"
)

func _isLongFormat(dsn string) bool {
	return strings.HasPrefix(dsn, "(")
}

func HandleDsn(dsn string, drvConfig *OracleDriverConfig) error {
	cleanupDSN := driverCommon.StripSpacesOutsideQuotes(dsn)

	var effectiveDsn = cleanupDSN

	// first extract credentials if specified
	parts := strings.SplitN(dsn, "@", 2)
	effectiveDsn = parts[0]
	if len(parts) == 2 {
		// we have some
		creds := strings.SplitN(parts[0], "/", 2)
		if len(creds) != 2 {
			return common.NewOracleError(oracleErrors.NamingDSNInvalid, nil, dsn)
		}
		if len(creds[0]) == 0 {
			return common.NewOracleError(oracleErrors.NamingDSNInvalid, nil, dsn)
		}
		// Credentials in config is not authorized when also specified
		// in dsn
		if drvConfig.Credentials.User != "" || drvConfig.Credentials.Password != "" {
			return common.NewOracleError(oracleErrors.ConflictingConnectionParameterSource, nil, "credentials")
		}
		drvConfig.Credentials.User = creds[0]
		drvConfig.Credentials.User = creds[1]
		effectiveDsn = parts[1]
	}

	// new extract parameters after the dns if any
	indexOfQMark := strings.LastIndex(effectiveDsn, "?")
	if indexOfQMark != -1 {
		// check that we did not locate a '?' within the '(...)'
		indexOfClosingP := strings.LastIndex(effectiveDsn, ")")
		if indexOfClosingP == -1 || indexOfClosingP < indexOfQMark {
			// that is a valid ? marker, deal with it.
			connectionParts := strings.SplitN(effectiveDsn, "?", 2)
			effectiveDsn = connectionParts[0]
			var parameterMap map[string]string
			var err error
			if len(connectionParts) == 2 && len(connectionParts[1]) > 0 {
				parameterMap, err = QueryStringToMap(connectionParts[1])
				if err != nil {
					return common.NewOracleError(oracleErrors.NamingDSNInvalid, err, dsn)
				}
				if err := drvConfig.AssignFromMap(parameterMap); err != nil {
					return common.NewOracleError(oracleErrors.NamingDSNInvalid, err, dsn)
				}
			}
		}
	}

	// new handle the connection string.
	if _isLongFormat(effectiveDsn) {
		return handleLongFormatDsn(effectiveDsn, drvConfig)
	} else {
		// everything is already parsed at that stage.
		drvConfig.ConnectDescriptor = effectiveDsn
		return nil
	}
}

func handleLongFormatDsn(dsn string, drvConfig *OracleDriverConfig) error {
	return nil
}
