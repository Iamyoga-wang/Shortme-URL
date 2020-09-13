package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

const (
	expTime       = 60
	longURL       = "https://www.example.com"
	shortLink     = "IFHzaO"
	shortLinkInfo = `{"url": "https://www.example.com", "created_at": "2020-09-12 13:55:32.354023 +0100 IST m=+1279.653669722", "expiration_in_minutes":1}`
)

type storageMock struct {
	mock.Mock
}

var app main.App
var mockR *storageMock

func (s *storageMock) Shorten(url string, exp int64) (string, error) {
	args := s.Called(url, exp)
	return args.String(0), args.Error(1)
}

func (s *storageMock) Unshorten(eid string) (string, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func (s *storageMock) ShortLinkInfo(eid string) (interface{}, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func init() {
	app = main.App{}
	mockR = new(storageMock)
	app.Initialize(&main.Env{S: mockR})
}

func TestCreateShortlink(t *testing.T) {
	var jsonStr = []byte(`{"url": "https://www.example.com", "expiration_in_minutes": 60}`)
	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonStr))

	if err != nil {
		t.Fatal("Should be able to create a request.", err)
	}
	req.Header.Set("Content-Type", "application/json")

	mockR.On("Shorten", longURL, int64(expTime)).Return(shortLink, nil).Once()
	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatal("Expected receive %d. Got %d", http.StatusCreated, rw.Code)
	}

	resp := struct {
		Shortlink string `json:"shortlink"`
	}{}
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("Should decode the response")
	}

	if resp.Shortlink != shortLink {
		t.Fatalf("Expected receive %s. Got %s", shortLink, resp.Shortlink)
	}
}

func TestRedirect(t *testing.T) {
	r := fmt.Sprintf("/%s", shortLink)
	req, err := http.NewRequest("GET", r, nil)
	if err != nil {
		t.Fatal("Should be able to create a request.", err)
	}

	mockR.On("Unshorten", shortLink).Return(longURL, nil).Once()
	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected receive %d. Got %d", http.StatusTemporaryRedirect, rw.Code)
	}
}
