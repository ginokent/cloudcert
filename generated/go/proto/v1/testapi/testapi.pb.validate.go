// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: proto/v1/testapi/testapi.proto

package testapiv1

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

// Validate checks the field values on TestAPIServiceEchoRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *TestAPIServiceEchoRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestAPIServiceEchoRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// TestAPIServiceEchoRequestMultiError, or nil if none found.
func (m *TestAPIServiceEchoRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *TestAPIServiceEchoRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if utf8.RuneCountInString(m.GetMessage()) < 1 {
		err := TestAPIServiceEchoRequestValidationError{
			field:  "Message",
			reason: "value length must be at least 1 runes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return TestAPIServiceEchoRequestMultiError(errors)
	}

	return nil
}

// TestAPIServiceEchoRequestMultiError is an error wrapping multiple validation
// errors returned by TestAPIServiceEchoRequest.ValidateAll() if the
// designated constraints aren't met.
type TestAPIServiceEchoRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestAPIServiceEchoRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestAPIServiceEchoRequestMultiError) AllErrors() []error { return m }

// TestAPIServiceEchoRequestValidationError is the validation error returned by
// TestAPIServiceEchoRequest.Validate if the designated constraints aren't met.
type TestAPIServiceEchoRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestAPIServiceEchoRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestAPIServiceEchoRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestAPIServiceEchoRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestAPIServiceEchoRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestAPIServiceEchoRequestValidationError) ErrorName() string {
	return "TestAPIServiceEchoRequestValidationError"
}

// Error satisfies the builtin error interface
func (e TestAPIServiceEchoRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestAPIServiceEchoRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestAPIServiceEchoRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestAPIServiceEchoRequestValidationError{}

// Validate checks the field values on TestAPIServiceEchoResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *TestAPIServiceEchoResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestAPIServiceEchoResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// TestAPIServiceEchoResponseMultiError, or nil if none found.
func (m *TestAPIServiceEchoResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *TestAPIServiceEchoResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	if len(errors) > 0 {
		return TestAPIServiceEchoResponseMultiError(errors)
	}

	return nil
}

// TestAPIServiceEchoResponseMultiError is an error wrapping multiple
// validation errors returned by TestAPIServiceEchoResponse.ValidateAll() if
// the designated constraints aren't met.
type TestAPIServiceEchoResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestAPIServiceEchoResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestAPIServiceEchoResponseMultiError) AllErrors() []error { return m }

// TestAPIServiceEchoResponseValidationError is the validation error returned
// by TestAPIServiceEchoResponse.Validate if the designated constraints aren't met.
type TestAPIServiceEchoResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestAPIServiceEchoResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestAPIServiceEchoResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestAPIServiceEchoResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestAPIServiceEchoResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestAPIServiceEchoResponseValidationError) ErrorName() string {
	return "TestAPIServiceEchoResponseValidationError"
}

// Error satisfies the builtin error interface
func (e TestAPIServiceEchoResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestAPIServiceEchoResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestAPIServiceEchoResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestAPIServiceEchoResponseValidationError{}

// Validate checks the field values on TestAPIServiceRaiseErrorRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *TestAPIServiceRaiseErrorRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestAPIServiceRaiseErrorRequest with
// the rules defined in the proto definition for this message. If any rules
// are violated, the result is a list of violation errors wrapped in
// TestAPIServiceRaiseErrorRequestMultiError, or nil if none found.
func (m *TestAPIServiceRaiseErrorRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *TestAPIServiceRaiseErrorRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	if len(errors) > 0 {
		return TestAPIServiceRaiseErrorRequestMultiError(errors)
	}

	return nil
}

// TestAPIServiceRaiseErrorRequestMultiError is an error wrapping multiple
// validation errors returned by TestAPIServiceRaiseErrorRequest.ValidateAll()
// if the designated constraints aren't met.
type TestAPIServiceRaiseErrorRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestAPIServiceRaiseErrorRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestAPIServiceRaiseErrorRequestMultiError) AllErrors() []error { return m }

// TestAPIServiceRaiseErrorRequestValidationError is the validation error
// returned by TestAPIServiceRaiseErrorRequest.Validate if the designated
// constraints aren't met.
type TestAPIServiceRaiseErrorRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestAPIServiceRaiseErrorRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestAPIServiceRaiseErrorRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestAPIServiceRaiseErrorRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestAPIServiceRaiseErrorRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestAPIServiceRaiseErrorRequestValidationError) ErrorName() string {
	return "TestAPIServiceRaiseErrorRequestValidationError"
}

// Error satisfies the builtin error interface
func (e TestAPIServiceRaiseErrorRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestAPIServiceRaiseErrorRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestAPIServiceRaiseErrorRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestAPIServiceRaiseErrorRequestValidationError{}

// Validate checks the field values on TestAPIServiceRaiseErrorResponse with
// the rules defined in the proto definition for this message. If any rules
// are violated, the first error encountered is returned, or nil if there are
// no violations.
func (m *TestAPIServiceRaiseErrorResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestAPIServiceRaiseErrorResponse with
// the rules defined in the proto definition for this message. If any rules
// are violated, the result is a list of violation errors wrapped in
// TestAPIServiceRaiseErrorResponseMultiError, or nil if none found.
func (m *TestAPIServiceRaiseErrorResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *TestAPIServiceRaiseErrorResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	if len(errors) > 0 {
		return TestAPIServiceRaiseErrorResponseMultiError(errors)
	}

	return nil
}

// TestAPIServiceRaiseErrorResponseMultiError is an error wrapping multiple
// validation errors returned by
// TestAPIServiceRaiseErrorResponse.ValidateAll() if the designated
// constraints aren't met.
type TestAPIServiceRaiseErrorResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestAPIServiceRaiseErrorResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestAPIServiceRaiseErrorResponseMultiError) AllErrors() []error { return m }

// TestAPIServiceRaiseErrorResponseValidationError is the validation error
// returned by TestAPIServiceRaiseErrorResponse.Validate if the designated
// constraints aren't met.
type TestAPIServiceRaiseErrorResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestAPIServiceRaiseErrorResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestAPIServiceRaiseErrorResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestAPIServiceRaiseErrorResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestAPIServiceRaiseErrorResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestAPIServiceRaiseErrorResponseValidationError) ErrorName() string {
	return "TestAPIServiceRaiseErrorResponseValidationError"
}

// Error satisfies the builtin error interface
func (e TestAPIServiceRaiseErrorResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestAPIServiceRaiseErrorResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestAPIServiceRaiseErrorResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestAPIServiceRaiseErrorResponseValidationError{}