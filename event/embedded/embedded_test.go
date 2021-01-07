package embedded_test

import (
	"testing"

	"github.com/akarasz/yahtzee/event"
	"github.com/akarasz/yahtzee/event/embedded"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	subject := embedded.New()
	suite.Run(t, &event.TestSuite{
		S: subject,
		E: subject,
	})
}
