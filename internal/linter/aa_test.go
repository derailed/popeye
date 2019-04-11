package linter

import (
	"fmt"
	"testing"

	m "github.com/petergtz/pegomock"
)

func TestSetup(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}
