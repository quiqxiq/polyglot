package hotspot

import "github.com/quixiq/polyglot/pkg/fault"

// ErrInvalidInput indicates invalid hotspot operation input.
var ErrInvalidInput = fault.New(fault.KindInvalidInput, "hotspot: invalid input")

// ErrNotFound indicates the requested hotspot resource does not exist.
var ErrNotFound = fault.New(fault.KindNotFound, "hotspot: not found")

// ErrExpireMonitorNotInstalled indicates the expire monitor is not configured.
var ErrExpireMonitorNotInstalled = fault.New(fault.KindFailedPrecondition, "hotspot: expire monitor not installed")
