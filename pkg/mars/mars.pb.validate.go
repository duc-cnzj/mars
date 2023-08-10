// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: mars/mars.proto

package mars

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

// Validate checks the field values on Config with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Config) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Config with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ConfigMultiError, or nil if none found.
func (m *Config) ValidateAll() error {
	return m.validate(true)
}

func (m *Config) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for ConfigFile

	// no validation rules for ConfigFileValues

	// no validation rules for ConfigField

	// no validation rules for IsSimpleEnv

	// no validation rules for ConfigFileType

	// no validation rules for LocalChartPath

	// no validation rules for ValuesYaml

	for idx, item := range m.GetElements() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ConfigValidationError{
						field:  fmt.Sprintf("Elements[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ConfigValidationError{
						field:  fmt.Sprintf("Elements[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ConfigValidationError{
					field:  fmt.Sprintf("Elements[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if m.GetDisplayName() != "" {

		if len(m.GetDisplayName()) > 64 {
			err := ConfigValidationError{
				field:  "DisplayName",
				reason: "value length must be at most 64 bytes",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if !_Config_DisplayName_Pattern.MatchString(m.GetDisplayName()) {
			err := ConfigValidationError{
				field:  "DisplayName",
				reason: "value does not match regex pattern \"^[A-Za-z]([A-Z-_a-z]*[^_-])*$\"",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

	}

	if len(errors) > 0 {
		return ConfigMultiError(errors)
	}

	return nil
}

// ConfigMultiError is an error wrapping multiple validation errors returned by
// Config.ValidateAll() if the designated constraints aren't met.
type ConfigMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ConfigMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ConfigMultiError) AllErrors() []error { return m }

// ConfigValidationError is the validation error returned by Config.Validate if
// the designated constraints aren't met.
type ConfigValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ConfigValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ConfigValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ConfigValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ConfigValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ConfigValidationError) ErrorName() string { return "ConfigValidationError" }

// Error satisfies the builtin error interface
func (e ConfigValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sConfig.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ConfigValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ConfigValidationError{}

var _Config_DisplayName_Pattern = regexp.MustCompile("^[A-Za-z]([A-Z-_a-z]*[^_-])*$")

// Validate checks the field values on Element with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Element) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Element with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ElementMultiError, or nil if none found.
func (m *Element) ValidateAll() error {
	return m.validate(true)
}

func (m *Element) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(m.GetPath()) < 1 {
		err := ElementValidationError{
			field:  "Path",
			reason: "value length must be at least 1 bytes",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if _, ok := ElementType_name[int32(m.GetType())]; !ok {
		err := ElementValidationError{
			field:  "Type",
			reason: "value must be one of the defined enum values",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	// no validation rules for Default

	// no validation rules for Description

	// no validation rules for Order

	if len(errors) > 0 {
		return ElementMultiError(errors)
	}

	return nil
}

// ElementMultiError is an error wrapping multiple validation errors returned
// by Element.ValidateAll() if the designated constraints aren't met.
type ElementMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ElementMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ElementMultiError) AllErrors() []error { return m }

// ElementValidationError is the validation error returned by Element.Validate
// if the designated constraints aren't met.
type ElementValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ElementValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ElementValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ElementValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ElementValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ElementValidationError) ErrorName() string { return "ElementValidationError" }

// Error satisfies the builtin error interface
func (e ElementValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sElement.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ElementValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ElementValidationError{}
