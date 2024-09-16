package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mikhaylov/beebroker/internal/config"
	"mikhaylov/beebroker/internal/store"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestGetTopics(t *testing.T) {
	config := config.Config{
		Host:        "localhost",
		Port:        8000,
		MaxMessages: 10,
		MaxQueues:   10,
		Timeout:     1,
	}

	memoryStore := store.NewMemoryStore(config)
	t.Run("put queue topic", func(t *testing.T) {
		request := newPushRequest("test", []byte(`{"message": "HELLO"}`))
		response := httptest.NewRecorder()

		srv := BrokerServer{store: memoryStore, config: config}
		srv.ServeHTTP(response, request)

		assertStatus(t, response.Code, 200)
	})
	t.Run("get queue topic", func(t *testing.T) {
		request := newPushRequest("test", []byte(`{"message": "HELLO"}`))
		response := httptest.NewRecorder()

		srv := BrokerServer{store: memoryStore, config: config}
		srv.ServeHTTP(response, request)

		request = newGetTopicRequest("test")
		response = httptest.NewRecorder()

		srv.ServeHTTP(response, request)
		dec := json.NewDecoder(response.Body)

		p := store.Payload{}
		err := dec.Decode(&p)
		if err != nil {
			t.Errorf("cand decode response")
		}

		assertResponseBody(t, p.Message, "HELLO")
		assertStatus(t, response.Code, 200)
	})

	t.Run("return queue topic name", func(t *testing.T) {
		request := newGetTopicRequest("test")
		topic := strings.Split(request.URL.Path, "/")

		got := topic[2]
		want := "test"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("return request timeout", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/queue/%s?timeout=5", "test"), nil)

		timeout := request.URL.Query().Get("timeout")
		got, _ := strconv.Atoi(timeout)
		want := 5

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func newGetTopicRequest(name string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/queue/%s", name), nil)
	return req
}

func newPushRequest(name string, data []byte) *http.Request {
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/queue/%s", name), bytes.NewReader(data))
	return req
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
