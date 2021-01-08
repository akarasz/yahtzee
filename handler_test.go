package yahtzee_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akarasz/yahtzee"
	"github.com/akarasz/yahtzee/model"
	store "github.com/akarasz/yahtzee/store/embedded"
)

func TestCreate(t *testing.T) {
	s := store.New()
	h := yahtzee.NewHandler(s, nil, nil)

	rr := newRequest(t, h, "POST", "/")

	assert.Exactly(t, http.StatusCreated, rr.Code)

	if assert.Contains(t, rr.HeaderMap, "Location") && assert.Len(t, rr.HeaderMap["Location"], 1) {
		got, err := s.Load(strings.TrimLeft(rr.HeaderMap["Location"][0], "/"))
		require.NoError(t, err)
		assert.Exactly(t, *model.NewGame(), got)
	}
}

func newRequest(t *testing.T, h http.Handler, method string, url string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	return rr
}
