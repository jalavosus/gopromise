package testutil

import (
	"github.com/stretchr/testify/assert"
)

func NilAssertion(wantNil bool) assert.ValueAssertionFunc {
	if wantNil {
		return assert.Nil
	}

	return assert.NotNil
}

func BoolAssertion(want bool) assert.BoolAssertionFunc {
	if want {
		return assert.True
	}

	return assert.False
}

func ErrorAssertion(want bool) assert.ErrorAssertionFunc {
	if want {
		return assert.Error
	}

	return assert.NoError
}
