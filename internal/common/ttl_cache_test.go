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

package common

import (
	"testing"
	"time"
)

func TestNewTTLCache(t *testing.T) {
	cache := NewTTLCache[string](2, 50*time.Millisecond)
	if cache == nil {
		t.Fatal("expected cache instance")
	}
	if cache.maxSize != 2 {
		t.Fatalf("expected maxSize 2, got %d", cache.maxSize)
	}
	if cache.ttl != 50*time.Millisecond {
		t.Fatalf("expected ttl 50ms, got %s", cache.ttl)
	}
	if cache.entries == nil {
		t.Fatal("expected entries map to be initialized")
	}
	if len(cache.entries) != 0 {
		t.Fatalf("expected empty cache, got %d entries", len(cache.entries))
	}
}

func TestTTLCacheStoresStringPointerValue(t *testing.T) {
	cache := NewTTLCache[*string](2, time.Minute)
	ip := "192.168.1.10"

	cache.Put(ip, &ip)

	got, found := cache.Get(ip)
	if !found {
		t.Fatal("expected pointer value to be found")
	}
	if got == nil {
		t.Fatal("expected non-nil pointer value")
	}
	if *got != ip {
		t.Fatalf("expected IP %q, got %q", ip, *got)
	}
}
