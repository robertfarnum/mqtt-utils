// Copyright 2021 Comcast. All rights reserved.

package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// WantError interface type
type WantError interface {
	error
	Compare(error) bool
}

// WantRegexpError uses regexp to pattern match for an error
type WantRegexpError struct {
	WantError
	regex regexp.Regexp
}

// NewWantRegexpError creates a new WantRegexpError with the given regex string
func NewWantRegexpError(regex string) error {
	return WantRegexpError{
		regex: *regexp.MustCompile(regex),
	}
}

// Error is a WantRegexpError string
func (wantErr WantRegexpError) Error() string {
	return fmt.Sprintf("%s", wantErr.regex.String())
}

// Compare is a WantRegexpError compare function
func (wantErr WantRegexpError) Compare(gotErr error) bool {
	gotErrStr := fmt.Sprintf("%v", gotErr)
	return wantErr.regex.MatchString(gotErrStr)
}

// WantStringError uses a string form matching an error
type WantStringError struct {
	WantError
	str string
}

// NewWantStringError creates a new WantStringError with the given string
func NewWantStringError(str string) error {
	return WantStringError{
		str: str,
	}
}

// Error is a WantStringError string
func (wantErr WantStringError) Error() string {
	return fmt.Sprintf("%s", wantErr.str)
}

// Compare is a WantStringError compare function
func (wantErr WantStringError) Compare(gotErr error) bool {
	if gotErr == nil {
		return false
	}

	return wantErr.Error() == gotErr.Error()
}

// ErrorPredicate functions handles error determination
type ErrorPredicate func(error) bool

// WantErrorPredicateError use an ErrorPredicate to determin an error
type WantErrorPredicateError struct {
	WantError
	predicate ErrorPredicate
}

// NewWantErrorPredicateError creates a new WantErrorPredicateError with the give ErrorPredicate
func NewWantErrorPredicateError(predicate ErrorPredicate) error {
	return WantErrorPredicateError{
		predicate: predicate,
	}
}

// Error is a WantErrorPredicateError string producer
func (wantErr WantErrorPredicateError) Error() string {
	return fmt.Sprintf("PredicateError")
}

// Compare is a WantErrorPredicateError compare function
func (wantErr WantErrorPredicateError) Compare(gotErr error) bool {
	if wantErr.predicate == nil {
		return false
	}

	return wantErr.predicate(gotErr)
}

// CheckResults interface
type CheckResults interface {
	Check() (bool, error)
	Compare() bool
	Error() error
}

// Results is a test result structure
type Results struct {
	CheckResults
	Got  interface{}
	Want interface{}
}

// ResultPredicate handler results determination
type ResultPredicate func(got interface{}) bool

func (results Results) compareInterfaceAndString(got interface{}, want string) bool {
	switch got := got.(type) {
	case *string:
		return want == *got
	case string:
		return want == got
	case interface{}:
		gotBytes, err := json.Marshal(got)
		if err != nil {
			return false
		}
		gotStr := string(gotBytes[:])
		return want == gotStr
	default:
		return reflect.DeepEqual(results.Got, results.Want)
	}
}

func (results Results) compareInterfaceAndRegexp(got interface{}, want regexp.Regexp) bool {
	switch got := got.(type) {
	case *string:
		return want.MatchString(*got)
	case interface{}:
		gotBytes, err := json.Marshal(got)
		if err != nil {
			return false
		}
		gotStr := string(gotBytes[:])
		return want.MatchString(gotStr)
	default:
		return reflect.DeepEqual(results.Got, results.Want)
	}
}

func (results Results) compareInterfaceAndResultPredicate(got interface{}, want ResultPredicate) bool {
	switch got := got.(type) {
	case *string:
		return want(*got)
	default:
		return want(got)
	}
}

func (results Results) compareInterfaceAndInterface(got interface{}, want interface{}) bool {
	gotBytes, err := json.Marshal(got)
	if err != nil {
		return false
	}
	wantBytes, err := json.Marshal(want)
	if err != nil {
		return false
	}
	gotStr := string(gotBytes[:])
	wantStr := string(wantBytes[:])
	return gotStr == wantStr
}

// Compare the results
func (results Results) Compare() bool {
	switch want := results.Want.(type) {
	case string:
		switch got := results.Got.(type) {
		case interface{}:
			return results.compareInterfaceAndString(got, want)

		default:
			return reflect.DeepEqual(results.Got, results.Want)
		}
	case regexp.Regexp:
		switch got := results.Got.(type) {
		case interface{}:
			return results.compareInterfaceAndRegexp(got, want)
		default:
			return reflect.DeepEqual(results.Got, results.Want)
		}
	case ResultPredicate:
		switch got := results.Got.(type) {
		case interface{}:
			return results.compareInterfaceAndResultPredicate(got, want)
		default:
			return want(got)
		}
	case interface{}:
		switch got := results.Got.(type) {
		case interface{}:
			return results.compareInterfaceAndInterface(got, want)
		default:
			return reflect.DeepEqual(results.Got, results.Want)
		}

	default:
		return reflect.DeepEqual(results.Got, results.Want)
	}
}

// Check the results
func (results Results) Check() (bool, error) {
	if results.Got == nil && results.Want == nil {
		return false, nil
	}

	if results.Got != nil && results.Want == nil {
		return false, results.Error()
	}

	if results.Got == nil && results.Want != nil {
		return false, results.Error()
	}

	if !results.Compare() {
		return false, results.Error()
	}

	return true, nil
}

// Error builds a results error
func (results Results) Error() error {
	got, _ := json.MarshalIndent(results.Got, "", "  ")
	want := ""
	switch resultsWant := results.Want.(type) {
	case regexp.Regexp:
		want = "Regexp"
	case ResultPredicate:
		want = "ResultPredicate"
	case interface{}:
		json, _ := json.MarshalIndent(resultsWant, "", "  ")
		want = string(json)
	case string:
		want = resultsWant
	}
	return fmt.Errorf("Results\n	got: \"%v\"\n	want: \"%v\"", string(got[:]), string(want[:]))
}

// ErrorResults is a test result error structure
type ErrorResults struct {
	CheckResults
	Got  error
	Want interface{}
}

// Compare the results
func (errorResults ErrorResults) Compare() bool {
	switch want := errorResults.Want.(type) {
	case ErrorPredicate:
		return want(errorResults.Got)
	case WantError:
		return want.Compare(errorResults.Got)
	case error:
		return errors.Is(errorResults.Got, want)
	default:
		return reflect.DeepEqual(errorResults.Got, errorResults.Want)
	}
}

// Check the error results
func (errorResults ErrorResults) Check() (bool, error) {
	if errorResults.Got == nil && errorResults.Want == nil {
		return false, nil
	}

	if errorResults.Got != nil && errorResults.Want == nil {
		return false, errorResults.Error()
	}

	if errorResults.Got == nil && errorResults.Want != nil {
		return false, errorResults.Error()
	}

	if !errorResults.Compare() {
		return false, errorResults.Error()
	}

	return true, nil
}

// Error builds a error results error
func (errorResults ErrorResults) Error() error {
	got := errorResults.Got
	want := ""
	switch resultsWant := errorResults.Want.(type) {
	case WantError:
		wantError := resultsWant.(WantError)
		want = wantError.Error()
	case error:
		wantError := resultsWant.(error)
		want = wantError.Error()
	case ErrorPredicate:
		want = "ErrorPredicate"
	case interface{}:
		json, _ := json.MarshalIndent(resultsWant, "", "  ")
		want = string(json)
	case string:
		want = resultsWant
	}
	return fmt.Errorf("Error Results\n	got: \"%v\"\n	want: \"%v\"", got, want)
}

// CheckTestResults takes the expected versus actual results and produces an error if detected
func CheckTestResults(got interface{}, want interface{}, err error, wantErr interface{}) error {
	errorResults := ErrorResults{
		Got:  err,
		Want: wantErr,
	}
	if quit, err := errorResults.Check(); quit || err != nil {
		return err
	}

	results := Results{
		Got:  got,
		Want: want,
	}
	_, err = results.Check()

	return err
}
