/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

package models

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/golang/protobuf/proto"
)

// Default fmt registry implementation suggests that it's thread safe, but
// we need to monitor it
var sharedFormatsRegistry = strfmt.NewFormats()

func init() {
	// Echo encodes/decodes base64 encoded byte arrays, no verification needed
	b64 := strfmt.Base64([]byte(nil))
	sharedFormatsRegistry.Add(
		"byte", &b64, func(_ string) bool { return true })
}

// Type to distinguish between Validation and Invalid message type errors
type ValidateError struct {
	s string
}

func (e *ValidateError) Error() string {
	if e == nil {
		return "Invalid ValidateError Pointer"
	}
	return e.s
}
func NewValidateError(str string) *ValidateError {
	return &ValidateError{s: str}
}
func ValidateErrorf(format string, a ...interface{}) *ValidateError {
	return NewValidateError(fmt.Sprintf(format, a...))
}

// mconfig_converters provides model receiver convertors to/from mconfig structs
type MconfigConverter interface {
	FromMconfig(msg proto.Message) error
	ToMconfig(msg proto.Message) error
	Verify() error
}
