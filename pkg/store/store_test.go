package store_test

import (
	"testing"

	"github.com/akarasz/yahtzee/pkg/game"
	"github.com/akarasz/yahtzee/pkg/store"
)

func TestStore_Put(t *testing.T) {
	t.Run("should add to store", func(t *testing.T) {
		spy := map[string]*game.Game{}
		s := store.New(spy)
		want := game.New()

		err := s.Put("id", want)
		if err != nil {
			t.Fatalf("returned error %q", err)
		}
		if got := spy["id"]; got != want {
			t.Errorf("wrong item in store. got %v, want %v", got, want)
		}
	})

	t.Run("should fail when trying to add with same id", func(t *testing.T) {
		spy := map[string]*game.Game{}
		s := store.New(spy)
		s.Put("id", game.New())

		got := s.Put("id", game.New())
		if want := store.ErrAlreadyExists; got != want {
			t.Fatalf("wrong error. got %q, want %q", got, want)
		}
	})
}

func TestStore_Get(t *testing.T) {
	t.Run("should return from store", func(t *testing.T) {
		spy := map[string]*game.Game{}
		s := store.New(spy)
		want := game.New()
		spy["id"] = want

		got, err := s.Get("id")
		if err != nil {
			t.Fatalf("returned error %q", err)
		}
		if got != want {
			t.Errorf("wrong item from store. got %v, want %v", got, want)
		}
	})

	t.Run("should fail when trying to add with same id", func(t *testing.T) {
		spy := map[string]*game.Game{}
		s := store.New(spy)

		_, got := s.Get("id")
		if want := store.ErrNotExists; got != want {
			t.Fatalf("wrong error. got %q, want %q", got, want)
		}
	})
}
