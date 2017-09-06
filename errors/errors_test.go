package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHere(t *testing.T) {
	funcName, fileName, fileLine, callerErr := Here(1)
	assert.Equal(t, "TestHere", funcName)
	assert.Equal(t, "errors/errors_test.go", fileName)
	assert.Condition(t, func() (success bool) {
		if fileLine > 0 {
			success = true
		}
		return
	})
	assert.Nil(t, callerErr)
}

func TestErrorfValueEquality(t *testing.T) {
	str := "ERR"
	e1 := Errorf(str)
	e2 := Errorf(str)
	assert.Equal(t, e1, e2)
}

func TestErrorCallerf(t *testing.T) {
	err := ErrorCallerf("ERR").Error()
	assert.Condition(t, func() (success bool) {
		success = strings.Contains(err, "errors/errors_test.go:")
		return
	})
	assert.Condition(t, func() (success bool) {
		success = strings.Contains(err, "TestErrorCallerf(): ERR")
		return
	})
}

var _ error = ErrorCollection(nil)

func TestErrorCollection(t *testing.T) {
	var x ErrorCollection
	x = append(x, errors.New("ERR 1"))
	x = append(x, errors.New("ERR 2"))

	var err error = x
	assert.Equal(t, "[ERR 1] [ERR 2]", err.Error())
}
