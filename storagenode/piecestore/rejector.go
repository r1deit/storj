// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package piecestore

import "sync/atomic"

const maxRejectorCount int32 = (1 << 31) - 1

// Rejector counts how many future requests should be rejected
// after each failed request.
type Rejector struct {
	increment int32
	remaining atomic.Int32
}

// NewRejector creates a rejector.
// increment == 0 disables it.
func NewRejector(increment int32) *Rejector {
	return &Rejector{increment: increment}
}

// RecordFailure adds increment to the number of pending rejections.
// The counter saturates at MaxInt32 to avoid overflow.
func (r *Rejector) RecordFailure() {
	if r == nil || r.increment <= 0 {
		return
	}

	for {
		current := r.remaining.Load()
		next := maxRejectorCount
		if current <= maxRejectorCount-r.increment {
			next = current + r.increment
		}
		if r.remaining.CompareAndSwap(current, next) {
			return
		}
	}
}

// TryReject consumes one pending rejection atomically.
// before is the value before decrement; after is the value after decrement.
func (r *Rejector) TryReject() (rejected bool, before int32, after int32) {
	if r == nil {
		return false, 0, 0
	}

	for {
		current := r.remaining.Load()
		if current <= 0 {
			return false, 0, 0
		}
		next := current - 1
		if r.remaining.CompareAndSwap(current, next) {
			return true, current, next
		}
	}
}
