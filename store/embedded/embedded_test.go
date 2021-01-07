package embedded_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee/store"
	"github.com/akarasz/yahtzee/store/embedded"
)

func TestSuite(t *testing.T) {
	s := embedded.New()
	suite.Run(t, &store.TestSuite{Subject: s})
}
