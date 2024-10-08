// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: file/file.proto

package file

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

// Validate checks the field values on DeleteRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeleteRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeleteRequestMultiError, or
// nil if none found.
func (m *DeleteRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetId() < 1 {
		err := DeleteRequestValidationError{
			field:  "Id",
			reason: "value must be greater than or equal to 1",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return DeleteRequestMultiError(errors)
	}

	return nil
}

// DeleteRequestMultiError is an error wrapping multiple validation errors
// returned by DeleteRequest.ValidateAll() if the designated constraints
// aren't met.
type DeleteRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteRequestMultiError) AllErrors() []error { return m }

// DeleteRequestValidationError is the validation error returned by
// DeleteRequest.Validate if the designated constraints aren't met.
type DeleteRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteRequestValidationError) ErrorName() string { return "DeleteRequestValidationError" }

// Error satisfies the builtin error interface
func (e DeleteRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteRequestValidationError{}

// Validate checks the field values on DeleteResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeleteResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeleteResponseMultiError,
// or nil if none found.
func (m *DeleteResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return DeleteResponseMultiError(errors)
	}

	return nil
}

// DeleteResponseMultiError is an error wrapping multiple validation errors
// returned by DeleteResponse.ValidateAll() if the designated constraints
// aren't met.
type DeleteResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteResponseMultiError) AllErrors() []error { return m }

// DeleteResponseValidationError is the validation error returned by
// DeleteResponse.Validate if the designated constraints aren't met.
type DeleteResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteResponseValidationError) ErrorName() string { return "DeleteResponseValidationError" }

// Error satisfies the builtin error interface
func (e DeleteResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteResponseValidationError{}

// Validate checks the field values on DiskInfoRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *DiskInfoRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DiskInfoRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DiskInfoRequestMultiError, or nil if none found.
func (m *DiskInfoRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DiskInfoRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return DiskInfoRequestMultiError(errors)
	}

	return nil
}

// DiskInfoRequestMultiError is an error wrapping multiple validation errors
// returned by DiskInfoRequest.ValidateAll() if the designated constraints
// aren't met.
type DiskInfoRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DiskInfoRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DiskInfoRequestMultiError) AllErrors() []error { return m }

// DiskInfoRequestValidationError is the validation error returned by
// DiskInfoRequest.Validate if the designated constraints aren't met.
type DiskInfoRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DiskInfoRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DiskInfoRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DiskInfoRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DiskInfoRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DiskInfoRequestValidationError) ErrorName() string { return "DiskInfoRequestValidationError" }

// Error satisfies the builtin error interface
func (e DiskInfoRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDiskInfoRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DiskInfoRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DiskInfoRequestValidationError{}

// Validate checks the field values on DiskInfoResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *DiskInfoResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DiskInfoResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DiskInfoResponseMultiError, or nil if none found.
func (m *DiskInfoResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *DiskInfoResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Usage

	// no validation rules for HumanizeUsage

	if len(errors) > 0 {
		return DiskInfoResponseMultiError(errors)
	}

	return nil
}

// DiskInfoResponseMultiError is an error wrapping multiple validation errors
// returned by DiskInfoResponse.ValidateAll() if the designated constraints
// aren't met.
type DiskInfoResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DiskInfoResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DiskInfoResponseMultiError) AllErrors() []error { return m }

// DiskInfoResponseValidationError is the validation error returned by
// DiskInfoResponse.Validate if the designated constraints aren't met.
type DiskInfoResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DiskInfoResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DiskInfoResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DiskInfoResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DiskInfoResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DiskInfoResponseValidationError) ErrorName() string { return "DiskInfoResponseValidationError" }

// Error satisfies the builtin error interface
func (e DiskInfoResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDiskInfoResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DiskInfoResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DiskInfoResponseValidationError{}

// Validate checks the field values on ListRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListRequestMultiError, or
// nil if none found.
func (m *ListRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ListRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for WithoutDeleted

	if m.Page != nil {
		// no validation rules for Page
	}

	if m.PageSize != nil {
		// no validation rules for PageSize
	}

	if len(errors) > 0 {
		return ListRequestMultiError(errors)
	}

	return nil
}

// ListRequestMultiError is an error wrapping multiple validation errors
// returned by ListRequest.ValidateAll() if the designated constraints aren't met.
type ListRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListRequestMultiError) AllErrors() []error { return m }

// ListRequestValidationError is the validation error returned by
// ListRequest.Validate if the designated constraints aren't met.
type ListRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListRequestValidationError) ErrorName() string { return "ListRequestValidationError" }

// Error satisfies the builtin error interface
func (e ListRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListRequestValidationError{}

// Validate checks the field values on ListResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListResponseMultiError, or
// nil if none found.
func (m *ListResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ListResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Page

	// no validation rules for PageSize

	for idx, item := range m.GetItems() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ListResponseValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ListResponseValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ListResponseValidationError{
					field:  fmt.Sprintf("Items[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	// no validation rules for Count

	if len(errors) > 0 {
		return ListResponseMultiError(errors)
	}

	return nil
}

// ListResponseMultiError is an error wrapping multiple validation errors
// returned by ListResponse.ValidateAll() if the designated constraints aren't met.
type ListResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListResponseMultiError) AllErrors() []error { return m }

// ListResponseValidationError is the validation error returned by
// ListResponse.Validate if the designated constraints aren't met.
type ListResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListResponseValidationError) ErrorName() string { return "ListResponseValidationError" }

// Error satisfies the builtin error interface
func (e ListResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListResponseValidationError{}

// Validate checks the field values on MaxUploadSizeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *MaxUploadSizeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on MaxUploadSizeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// MaxUploadSizeRequestMultiError, or nil if none found.
func (m *MaxUploadSizeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *MaxUploadSizeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return MaxUploadSizeRequestMultiError(errors)
	}

	return nil
}

// MaxUploadSizeRequestMultiError is an error wrapping multiple validation
// errors returned by MaxUploadSizeRequest.ValidateAll() if the designated
// constraints aren't met.
type MaxUploadSizeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m MaxUploadSizeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m MaxUploadSizeRequestMultiError) AllErrors() []error { return m }

// MaxUploadSizeRequestValidationError is the validation error returned by
// MaxUploadSizeRequest.Validate if the designated constraints aren't met.
type MaxUploadSizeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MaxUploadSizeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MaxUploadSizeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MaxUploadSizeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MaxUploadSizeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MaxUploadSizeRequestValidationError) ErrorName() string {
	return "MaxUploadSizeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e MaxUploadSizeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMaxUploadSizeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MaxUploadSizeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MaxUploadSizeRequestValidationError{}

// Validate checks the field values on MaxUploadSizeResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *MaxUploadSizeResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on MaxUploadSizeResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// MaxUploadSizeResponseMultiError, or nil if none found.
func (m *MaxUploadSizeResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *MaxUploadSizeResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for HumanizeSize

	// no validation rules for Bytes

	if len(errors) > 0 {
		return MaxUploadSizeResponseMultiError(errors)
	}

	return nil
}

// MaxUploadSizeResponseMultiError is an error wrapping multiple validation
// errors returned by MaxUploadSizeResponse.ValidateAll() if the designated
// constraints aren't met.
type MaxUploadSizeResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m MaxUploadSizeResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m MaxUploadSizeResponseMultiError) AllErrors() []error { return m }

// MaxUploadSizeResponseValidationError is the validation error returned by
// MaxUploadSizeResponse.Validate if the designated constraints aren't met.
type MaxUploadSizeResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MaxUploadSizeResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MaxUploadSizeResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MaxUploadSizeResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MaxUploadSizeResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MaxUploadSizeResponseValidationError) ErrorName() string {
	return "MaxUploadSizeResponseValidationError"
}

// Error satisfies the builtin error interface
func (e MaxUploadSizeResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMaxUploadSizeResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MaxUploadSizeResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MaxUploadSizeResponseValidationError{}

// Validate checks the field values on ShowRecordsRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ShowRecordsRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ShowRecordsRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ShowRecordsRequestMultiError, or nil if none found.
func (m *ShowRecordsRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ShowRecordsRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetId() <= 0 {
		err := ShowRecordsRequestValidationError{
			field:  "Id",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return ShowRecordsRequestMultiError(errors)
	}

	return nil
}

// ShowRecordsRequestMultiError is an error wrapping multiple validation errors
// returned by ShowRecordsRequest.ValidateAll() if the designated constraints
// aren't met.
type ShowRecordsRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ShowRecordsRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ShowRecordsRequestMultiError) AllErrors() []error { return m }

// ShowRecordsRequestValidationError is the validation error returned by
// ShowRecordsRequest.Validate if the designated constraints aren't met.
type ShowRecordsRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ShowRecordsRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ShowRecordsRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ShowRecordsRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ShowRecordsRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ShowRecordsRequestValidationError) ErrorName() string {
	return "ShowRecordsRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ShowRecordsRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sShowRecordsRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ShowRecordsRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ShowRecordsRequestValidationError{}

// Validate checks the field values on ShowRecordsResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ShowRecordsResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ShowRecordsResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ShowRecordsResponseMultiError, or nil if none found.
func (m *ShowRecordsResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ShowRecordsResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return ShowRecordsResponseMultiError(errors)
	}

	return nil
}

// ShowRecordsResponseMultiError is an error wrapping multiple validation
// errors returned by ShowRecordsResponse.ValidateAll() if the designated
// constraints aren't met.
type ShowRecordsResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ShowRecordsResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ShowRecordsResponseMultiError) AllErrors() []error { return m }

// ShowRecordsResponseValidationError is the validation error returned by
// ShowRecordsResponse.Validate if the designated constraints aren't met.
type ShowRecordsResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ShowRecordsResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ShowRecordsResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ShowRecordsResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ShowRecordsResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ShowRecordsResponseValidationError) ErrorName() string {
	return "ShowRecordsResponseValidationError"
}

// Error satisfies the builtin error interface
func (e ShowRecordsResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sShowRecordsResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ShowRecordsResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ShowRecordsResponseValidationError{}
