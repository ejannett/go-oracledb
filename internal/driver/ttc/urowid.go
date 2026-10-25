/*
** Copyright (c) 2026 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively, the "Software"), free of charge and under any and all copyright
** rights and under any and all patent rights owned or freely licensable by each
** licensor hereunder covering either (i) the unmodified Software as contributed
** to or provided by such licensor, or (ii) the Larger Works (as defined below),
** to deal in both the Software and any Larger Work, without restriction.
**
** This license is subject to the condition that the above copyright notice and
** this permission notice be included in all copies or substantial portions of
** the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 */

package ttc

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

const (
	urowidPhysicalType = 1
	urowidTypeCount    = 5
)

var (
	urowidBase64Alphabet = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	urowidTypePrefix     = [...]byte{'A', '*', '-', '(', ')'}
)

// DecodeUrowidColumn converts Oracle's internal universal rowid bytes to the
// printable rowid representation returned by Oracle clients.
func DecodeUrowidColumn(columnContext columnContext, value any) (driver.Value, error) {
	data, err := rawColumnData(value)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	rowid, err := urowidToString(data)
	if err != nil {
		return nil, rowDecodeError(columnContext, err, "UROWID")
	}
	return rowid, nil
}

func urowidToString(data []byte) (string, error) {
	rowidType := data[0]
	if rowidType == urowidPhysicalType {
		// Physical universal rowids contain a type byte followed by the
		// object, file, block, and slot components in big-endian order.
		if len(data) != 13 {
			return "", fmt.Errorf("physical UROWID has %d bytes; want 13", len(data))
		}
		object := binary.BigEndian.Uint32(data[1:5])
		file := binary.BigEndian.Uint16(data[5:7])
		block := binary.BigEndian.Uint32(data[7:11])
		slot := binary.BigEndian.Uint16(data[11:13])

		out := make([]byte, 0, 18)
		out = append(out, urowidBase64Uint(uint64(object), 6)...)
		out = append(out, urowidBase64Uint(uint64(file), 3)...)
		out = append(out, urowidBase64Uint(uint64(block), 6)...)
		out = append(out, urowidBase64Uint(uint64(slot), 3)...)
		return string(out), nil
	}
	if rowidType < 1 || rowidType > urowidTypeCount {
		return "", fmt.Errorf("unknown UROWID type indicator %d", rowidType)
	}

	body := base64.RawStdEncoding.EncodeToString(data[1:])
	return string(urowidTypePrefix[rowidType-1]) + body, nil
}

func urowidBase64Uint(value uint64, digits int) []byte {
	encoded := make([]byte, digits)
	for i := digits - 1; i >= 0; i-- {
		encoded[i] = urowidBase64Alphabet[value&0x3f]
		value >>= 6
	}
	return encoded
}
