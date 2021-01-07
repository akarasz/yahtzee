package in_memory_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee/store"
	"github.com/akarasz/yahtzee/store/in_memory"
)

func TestSuite(t *testing.T) {
	s := in_memory.New()
	suite.Run(t, &store.TestSuite{Subject: s})
}
