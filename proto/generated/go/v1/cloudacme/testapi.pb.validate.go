// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: v1/cloudacme/testapi.proto

package cloudacme

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on TestAPIEchoRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *TestAPIEchoRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestAPIEchoRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// TestAPIEchoRequestMultiError, or nil if none found.
func (m *TestAPIEchoRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *TestAPIEchoRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	if len(errors) > 0 {
		return TestAPIEchoRequestMultiError(errors)
	}

	return nil
}

// TestAPIEchoRequestMultiError is an error wrapping multiple validation errors
// returned by TestAPIEchoRequest.ValidateAll() if the designated constraints
// aren't met.
type TestAPIEchoRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestAPIEchoRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestAPIEchoRequestMultiError) AllErrors() []error { return m }

// TestAPIEchoRequestValidationError is the validation error returned by
// TestAPIEchoRequest.Validate if the designated constraints aren't met.
type TestAPIEchoRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestAPIEchoRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestAPIEchoRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestAPIEchoRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestAPIEchoRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestAPIEchoRequestValidationError) ErrorName() string {
	return "TestAPIEchoRequestValidationError"
}

// Error satisfies the builtin error interface
func (e TestAPIEchoRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestAPIEchoRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestAPIEchoRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestAPIEchoRequestValidationError{}

// Validate checks the field values on TestAPIEchoResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *TestAPIEchoResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestAPIEchoResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// TestAPIEchoResponseMultiError, or nil if none found.
func (m *TestAPIEchoResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *TestAPIEchoResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	if len(errors) > 0 {
		return TestAPIEchoResponseMultiError(errors)
	}

	return nil
}

// TestAPIEchoResponseMultiError is an error wrapping multiple validation
// errors returned by TestAPIEchoResponse.ValidateAll() if the designated
// constraints aren't met.
type TestAPIEchoResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestAPIEchoResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestAPIEchoResponseMultiError) AllErrors() []error { return m }

// TestAPIEchoResponseValidationError is the validation error returned by
// TestAPIEchoResponse.Validate if the designated constraints aren't met.
type TestAPIEchoResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestAPIEchoResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestAPIEchoResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestAPIEchoResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestAPIEchoResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestAPIEchoResponseValidationError) ErrorName() string {
	return "TestAPIEchoResponseValidationError"
}

// Error satisfies the builtin error interface
func (e TestAPIEchoResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestAPIEchoResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestAPIEchoResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestAPIEchoResponseValidationError{}