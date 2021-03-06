// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: metrics/metrics.proto

package metrics

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

// Validate checks the field values on TopPodRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *TopPodRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TopPodRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in TopPodRequestMultiError, or
// nil if none found.
func (m *TopPodRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *TopPodRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetNamespace()) < 1 {
		err := TopPodRequestValidationError{
			field:  "Namespace",
			reason: "value length must be at least 1 bytes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(m.GetPod()) < 1 {
		err := TopPodRequestValidationError{
			field:  "Pod",
			reason: "value length must be at least 1 bytes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return TopPodRequestMultiError(errors)
	}

	return nil
}

// TopPodRequestMultiError is an error wrapping multiple validation errors
// returned by TopPodRequest.ValidateAll() if the designated constraints
// aren't met.
type TopPodRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TopPodRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TopPodRequestMultiError) AllErrors() []error { return m }

// TopPodRequestValidationError is the validation error returned by
// TopPodRequest.Validate if the designated constraints aren't met.
type TopPodRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TopPodRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TopPodRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TopPodRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TopPodRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TopPodRequestValidationError) ErrorName() string { return "TopPodRequestValidationError" }

// Error satisfies the builtin error interface
func (e TopPodRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTopPodRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TopPodRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TopPodRequestValidationError{}

// Validate checks the field values on TopPodResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *TopPodResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TopPodResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in TopPodResponseMultiError,
// or nil if none found.
func (m *TopPodResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *TopPodResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Cpu

	// no validation rules for Memory

	// no validation rules for HumanizeCpu

	// no validation rules for HumanizeMemory

	// no validation rules for Time

	// no validation rules for Length

	if len(errors) > 0 {
		return TopPodResponseMultiError(errors)
	}

	return nil
}

// TopPodResponseMultiError is an error wrapping multiple validation errors
// returned by TopPodResponse.ValidateAll() if the designated constraints
// aren't met.
type TopPodResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TopPodResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TopPodResponseMultiError) AllErrors() []error { return m }

// TopPodResponseValidationError is the validation error returned by
// TopPodResponse.Validate if the designated constraints aren't met.
type TopPodResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TopPodResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TopPodResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TopPodResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TopPodResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TopPodResponseValidationError) ErrorName() string { return "TopPodResponseValidationError" }

// Error satisfies the builtin error interface
func (e TopPodResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTopPodResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TopPodResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TopPodResponseValidationError{}

// Validate checks the field values on CpuMemoryInNamespaceRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CpuMemoryInNamespaceRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CpuMemoryInNamespaceRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CpuMemoryInNamespaceRequestMultiError, or nil if none found.
func (m *CpuMemoryInNamespaceRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *CpuMemoryInNamespaceRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetNamespaceId() <= 0 {
		err := CpuMemoryInNamespaceRequestValidationError{
			field:  "NamespaceId",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return CpuMemoryInNamespaceRequestMultiError(errors)
	}

	return nil
}

// CpuMemoryInNamespaceRequestMultiError is an error wrapping multiple
// validation errors returned by CpuMemoryInNamespaceRequest.ValidateAll() if
// the designated constraints aren't met.
type CpuMemoryInNamespaceRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CpuMemoryInNamespaceRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CpuMemoryInNamespaceRequestMultiError) AllErrors() []error { return m }

// CpuMemoryInNamespaceRequestValidationError is the validation error returned
// by CpuMemoryInNamespaceRequest.Validate if the designated constraints
// aren't met.
type CpuMemoryInNamespaceRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CpuMemoryInNamespaceRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CpuMemoryInNamespaceRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CpuMemoryInNamespaceRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CpuMemoryInNamespaceRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CpuMemoryInNamespaceRequestValidationError) ErrorName() string {
	return "CpuMemoryInNamespaceRequestValidationError"
}

// Error satisfies the builtin error interface
func (e CpuMemoryInNamespaceRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCpuMemoryInNamespaceRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CpuMemoryInNamespaceRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CpuMemoryInNamespaceRequestValidationError{}

// Validate checks the field values on CpuMemoryInNamespaceResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CpuMemoryInNamespaceResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CpuMemoryInNamespaceResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CpuMemoryInNamespaceResponseMultiError, or nil if none found.
func (m *CpuMemoryInNamespaceResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *CpuMemoryInNamespaceResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Cpu

	// no validation rules for Memory

	if len(errors) > 0 {
		return CpuMemoryInNamespaceResponseMultiError(errors)
	}

	return nil
}

// CpuMemoryInNamespaceResponseMultiError is an error wrapping multiple
// validation errors returned by CpuMemoryInNamespaceResponse.ValidateAll() if
// the designated constraints aren't met.
type CpuMemoryInNamespaceResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CpuMemoryInNamespaceResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CpuMemoryInNamespaceResponseMultiError) AllErrors() []error { return m }

// CpuMemoryInNamespaceResponseValidationError is the validation error returned
// by CpuMemoryInNamespaceResponse.Validate if the designated constraints
// aren't met.
type CpuMemoryInNamespaceResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CpuMemoryInNamespaceResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CpuMemoryInNamespaceResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CpuMemoryInNamespaceResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CpuMemoryInNamespaceResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CpuMemoryInNamespaceResponseValidationError) ErrorName() string {
	return "CpuMemoryInNamespaceResponseValidationError"
}

// Error satisfies the builtin error interface
func (e CpuMemoryInNamespaceResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCpuMemoryInNamespaceResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CpuMemoryInNamespaceResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CpuMemoryInNamespaceResponseValidationError{}

// Validate checks the field values on CpuMemoryInProjectRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CpuMemoryInProjectRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CpuMemoryInProjectRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CpuMemoryInProjectRequestMultiError, or nil if none found.
func (m *CpuMemoryInProjectRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *CpuMemoryInProjectRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetProjectId() <= 0 {
		err := CpuMemoryInProjectRequestValidationError{
			field:  "ProjectId",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return CpuMemoryInProjectRequestMultiError(errors)
	}

	return nil
}

// CpuMemoryInProjectRequestMultiError is an error wrapping multiple validation
// errors returned by CpuMemoryInProjectRequest.ValidateAll() if the
// designated constraints aren't met.
type CpuMemoryInProjectRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CpuMemoryInProjectRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CpuMemoryInProjectRequestMultiError) AllErrors() []error { return m }

// CpuMemoryInProjectRequestValidationError is the validation error returned by
// CpuMemoryInProjectRequest.Validate if the designated constraints aren't met.
type CpuMemoryInProjectRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CpuMemoryInProjectRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CpuMemoryInProjectRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CpuMemoryInProjectRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CpuMemoryInProjectRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CpuMemoryInProjectRequestValidationError) ErrorName() string {
	return "CpuMemoryInProjectRequestValidationError"
}

// Error satisfies the builtin error interface
func (e CpuMemoryInProjectRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCpuMemoryInProjectRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CpuMemoryInProjectRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CpuMemoryInProjectRequestValidationError{}

// Validate checks the field values on CpuMemoryInProjectResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *CpuMemoryInProjectResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CpuMemoryInProjectResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CpuMemoryInProjectResponseMultiError, or nil if none found.
func (m *CpuMemoryInProjectResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *CpuMemoryInProjectResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Cpu

	// no validation rules for Memory

	if len(errors) > 0 {
		return CpuMemoryInProjectResponseMultiError(errors)
	}

	return nil
}

// CpuMemoryInProjectResponseMultiError is an error wrapping multiple
// validation errors returned by CpuMemoryInProjectResponse.ValidateAll() if
// the designated constraints aren't met.
type CpuMemoryInProjectResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CpuMemoryInProjectResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CpuMemoryInProjectResponseMultiError) AllErrors() []error { return m }

// CpuMemoryInProjectResponseValidationError is the validation error returned
// by CpuMemoryInProjectResponse.Validate if the designated constraints aren't met.
type CpuMemoryInProjectResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CpuMemoryInProjectResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CpuMemoryInProjectResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CpuMemoryInProjectResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CpuMemoryInProjectResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CpuMemoryInProjectResponseValidationError) ErrorName() string {
	return "CpuMemoryInProjectResponseValidationError"
}

// Error satisfies the builtin error interface
func (e CpuMemoryInProjectResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCpuMemoryInProjectResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CpuMemoryInProjectResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CpuMemoryInProjectResponseValidationError{}
