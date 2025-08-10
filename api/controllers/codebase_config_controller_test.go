package controllers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
)

func TestNewCodebaseConfigController(t *testing.T) {
	// Basic test to ensure the controller can be instantiated
	controller := controllers.NewCodebaseConfigController(nil)
	assert.NotNil(t, controller)
}
