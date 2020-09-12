package store

import (
	"reflect"
	"testing"

	"github.com/akarasz/yahtzee/models"
)

func TestInMemory_Put(t *testing.T) {
	t.Run("should add to store", func(t *testing.T) {
		s := New()
		want := *models.NewGame()

		err := s.Save("id", want)
		if err != nil {
			t.Fatalf("returned error %q", err)
		}
		if got := s.repo["id"]; !reflect.DeepEqual(got, want) {
			t.Errorf("wrong item in store. got %v, want %v", got, want)
		}
	})
}

func TestInMemory_Get(t *testing.T) {
	t.Run("should return from store", func(t *testing.T) {
		s := New()
		want := *models.NewGame()
		s.repo["id"] = want

		got, err := s.Load("id")
		if err != nil {
			t.Fatalf("returned error %q", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("wrong item from store. got %v, want %v", got, want)
		}
	})

	t.Run("should fail when trying to add with same id", func(t *testing.T) {
		s := New()

		_, got := s.Load("id")
		if want := ErrNotExists; got != want {
			t.Fatalf("wrong error. got %q, want %q", got, want)
		}
	})
}
