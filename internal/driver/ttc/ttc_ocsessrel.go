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

package ttc

import (
	"context"

	driverCommon "github.com/oracle/go-oracledb/v26/internal/driver/common"
)

// ttiSPFOCSessrel server's "session-release values" message for a pooled-session release.
// This message is received from the server and therefore only implements
// UnMarshalFrom and GetMsgCode; it does not support MarshalTo.
type ttiSPFOCSessrel struct {
	sessrlstag  string
	sessrlsmode driverCommon.UB4
}

// newttiSPFOCSessrel allocates a new receiver for TTISPF/OCSSESSREL payloads.
// The returned value implements common.Message and is intended to be populated
// via UnMarshalFrom by the MessageStreamer.
func newttiSPFOCSessrel() driverCommon.Message[driverCommon.MessageType] {

	return &ttiSPFOCSessrel{sessrlsmode: 0}
}

// GetMsgCode implements common.Message and identifies this message as TTISPF
// (Server-side piggyback).
func (spf *ttiSPFOCSessrel) GetMsgCode() driverCommon.MessageType {
	return TTIONEWAYFN
}

// getFuncCode returns the piggyback function code associated with this message.
func (spf *ttiSPFOCSessrel) GetFuncCode() driverCommon.FunctionType {
	return driverCommon.FunctionType(ocsessrls)
}

// UnMarshalFrom reads a TTISPF/OCSSESSREL payload from the wire.
// Expected layout (as observed from network traces):
func (spf *ttiSPFOCSessrel) MarshalTo(ctx context.Context, engine driverCommon.Marshaller) error {
	engine.MarshalSB4(ctx, 0)
	engine.MarshalNullPTR(ctx)
	engine.MarshalUB4(ctx, spf.sessrlsmode)

	return nil
}
