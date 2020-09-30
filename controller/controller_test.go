package controller

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/akarasz/yahtzee/events"
	events_mocks "github.com/akarasz/yahtzee/events/mocks"
	"github.com/akarasz/yahtzee/models"
	service_mocks "github.com/akarasz/yahtzee/service/mocks"
	store_mocks "github.com/akarasz/yahtzee/store/mocks"
)

func TestCreate(t *testing.T) {
	t.Run("should return the id of the saved game", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		var savedID string
		mockStore.EXPECT().
			Save(gomock.Any(), gomock.Any()).
			Do(func(id string, g models.Game) {
				savedID = id
			}).
			Return(nil).
			Times(1)

		returnedID, err := c.Create()
		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got, want := returnedID, savedID; got != want {
			t.Errorf("invalid ID returned; got %q want %q", got, want)
		}
	})

	t.Run("should save a game with zero counters and no players", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		var savedGame *models.Game
		mockStore.EXPECT().
			Save(gomock.Any(), gomock.Any()).
			Do(func(id string, g models.Game) {
				savedGame = &g
			}).
			Return(nil).
			Times(1)

		_, err := c.Create()
		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if got, want := len(savedGame.Players), 0; got != want {
			t.Errorf("invalid number of players; got %q want %q", got, want)
		}
		if got, want := savedGame.RollCount, 0; got != want {
			t.Errorf("invalid roll count; got %q want %q", got, want)
		}
		if got, want := savedGame.CurrentPlayer, 0; got != want {
			t.Errorf("invalid current player; got %q want %q", got, want)
		}
		if got, want := savedGame.Round, 0; got != want {
			t.Errorf("invalid round; got %q want %q", got, want)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("should return the loaded game from store", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		want := &models.Game{
			Players:   []*models.Player{},
			Dices:     []*models.Dice{},
			RollCount: 2,
			Round:     8,
		}

		mockStore.EXPECT().
			Load(gomock.Eq("gameID")).
			Return(*want, nil).
			AnyTimes()

		got, err := c.Get("gameID")
		if err != nil {
			t.Fatalf("unexpected error: %T: %v", err, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("not returning the same from the store; got %v want %v", got, want)
		}
	})
}

func TestAddPlayer(t *testing.T) {
	t.Run("should apply add player on loaded game and save again", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockService := service_mocks.NewMockGame(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		u := models.NewUser("alice")
		before := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}
		after := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}

		gomock.InOrder(
			mockStore.EXPECT().Load("gameID").Return(before, nil),
			mockServiceProvider.EXPECT().Create(gomock.Eq(before), *u).Return(mockService),
			mockService.EXPECT().AddPlayer().Return(after, nil),
			mockStore.EXPECT().Save("gameID", gomock.Eq(after)),
			mockEvents.EXPECT().Emit("gameID", events.AddPlayer, gomock.Any()),
		)

		c.AddPlayer(u, "gameID")
	})
}

func TestRoll(t *testing.T) {
	t.Run("should apply roll on loaded game and save again", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockService := service_mocks.NewMockGame(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		u := models.NewUser("alice")
		before := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}
		after := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}

		gomock.InOrder(
			mockStore.EXPECT().Load("gameID").Return(before, nil),
			mockServiceProvider.EXPECT().Create(gomock.Eq(before), *u).Return(mockService),
			mockService.EXPECT().Roll().Return(after, nil),
			mockStore.EXPECT().Save("gameID", gomock.Eq(after)),
			mockEvents.EXPECT().Emit("gameID", events.Roll, gomock.Any()),
		)

		c.Roll(u, "gameID")
	})
}

func TestLock(t *testing.T) {
	t.Run("should apply lock on loaded game and save again", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockService := service_mocks.NewMockGame(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		u := models.NewUser("alice")
		before := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}
		after := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}

		gomock.InOrder(
			mockStore.EXPECT().Load("gameID").Return(before, nil),
			mockServiceProvider.EXPECT().Create(gomock.Eq(before), *u).Return(mockService),
			mockService.EXPECT().Lock(4).Return(after, nil),
			mockStore.EXPECT().Save("gameID", gomock.Eq(after)),
			mockEvents.EXPECT().Emit("gameID", events.Lock, gomock.Any()),
		)

		c.Lock(u, "gameID", "4")
	})
}

func TestScore(t *testing.T) {
	t.Run("should apply score on loaded game and save again", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockStore := store_mocks.NewMockStore(mockCtrl)
		mockServiceProvider := service_mocks.NewMockProvider(mockCtrl)
		mockService := service_mocks.NewMockGame(mockCtrl)
		mockEvents := events_mocks.NewMockEmitter(mockCtrl)

		c := New(mockStore, mockServiceProvider, mockEvents)

		u := models.NewUser("alice")
		before := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}
		after := models.Game{
			Players: []*models.Player{},
			Dices:   []*models.Dice{},
		}

		gomock.InOrder(
			mockStore.EXPECT().Load("gameID").Return(before, nil),
			mockServiceProvider.EXPECT().Create(gomock.Eq(before), *u).Return(mockService),
			mockService.EXPECT().Score(models.Category("test")).Return(after, nil),
			mockStore.EXPECT().Save("gameID", gomock.Eq(after)),
			mockEvents.EXPECT().Emit("gameID", events.Score, gomock.Any()),
		)

		c.Score(u, "gameID", models.Category("test"))
	})
}