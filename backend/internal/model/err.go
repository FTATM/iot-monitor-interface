package model

import "errors"

// Custome Error
var ErrDuplicate = errors.New("Duplicate Data")
var ErrSecurityViolation = errors.New("Security Violation")
var ErrNotActive = errors.New("Not Active")
var ErrInUsed = errors.New("In Used")
