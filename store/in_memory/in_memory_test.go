package in_memory_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/yahtzee/store/in_memory"
	storetest "github.com/akarasz/yahtzee/store/internal/test"
)

func TestStore(t *testing.T) {
	s := in_memory.New()
	suite.Run(t, storetest.New(s))
}
