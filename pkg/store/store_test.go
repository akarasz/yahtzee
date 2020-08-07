package store

import (
	"testing"

	"github.com/akarasz/yahtzee/pkg/game"
)

func TestInMemory_Put(t *testing.T) {
	t.Run("should add to store", func(t *testing.T) {
		s := NewInMemory()
		want := game.New()

		err := s.Put("id", want)
		if err != nil {
			t.Fatalf("returned error %q", err)
		}
		if got := s.repo["id"]; got != want {
			t.Errorf("wrong item in store. got %v, want %v", got, want)
		}
	})

	t.Run("should fail when trying to add with same id", func(t *testing.T) {
		s := NewInMemory()
		s.Put("id", game.New())

		got := s.Put("id", game.New())
		if want := ErrAlreadyExists; got != want {
			t.Fatalf("wrong error. got %q, want %q", got, want)
		}
	})
}

func TestInMemory_Get(t *testing.T) {
	t.Run("should return from store", func(t *testing.T) {
		s := NewInMemory()
		want := game.New()
		s.repo["id"] = want

		got, err := s.Get("id")
		if err != nil {
			t.Fatalf("returned error %q", err)
		}
		if got != want {
			t.Errorf("wrong item from store. got %v, want %v", got, want)
		}
	})

	t.Run("should fail when trying to add with same id", func(t *testing.T) {
		s := NewInMemory()

		_, got := s.Get("id")
		if want := ErrNotExists; got != want {
			t.Fatalf("wrong error. got %q, want %q", got, want)
		}
	})
}
