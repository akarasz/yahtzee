package in_app_test

import (
	"testing"

	"github.com/akarasz/yahtzee/event"
	"github.com/akarasz/yahtzee/event/in_app"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	subject := in_app.New()
	suite.Run(t, &event.TestSuite{
		S: subject,
		E: subject,
	})
}
