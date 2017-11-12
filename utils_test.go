package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomName(t *testing.T) {
	assert.Equal(t, randomAlliterateCombo(10), randomAlliterateCombo(10))
	assert.NotEqual(t, randomAlliterateCombo(11), randomAlliterateCombo(10))
}
