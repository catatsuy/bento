package gemini_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	. "github.com/catatsuy/bento/internal/gemini"
	"github.com/google/go-cmp/cmp"
)

func TestChat_Success(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Use the Gemini API default model now.
	param := &Payload{
		Model: "gemini-2.0-flash-lite",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Tell me a joke.",
			},
		},
	}

	token := "test-token"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check headers.
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type expected 'application/json', got %s", r.Header.Get("Content-Type"))
		}
		expectedAuth := "Bearer " + token
		if r.Header.Get("Authorization") != expectedAuth {
			t.Fatalf("Authorization expected '%s', got %s", expectedAuth, r.Header.Get("Authorization"))
		}

		// Check the request body.
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		actualPayload := &Payload{}
		if err := json.Unmarshal(bodyBytes, actualPayload); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actualPayload, param) {
			t.Fatalf("expected %+v, got %+v", param, actualPayload)
		}

		// Serve the success response from the testdata file.
		http.ServeFile(w, r, "testdata/gemini_success.json")
	})

	client, err := NewClient(server.URL, token)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Chat(context.Background(), param)
	if err != nil {
		t.Fatal(err)
	}

	expected := &Response{
		ID:      "gemini123",
		Object:  "chat.completion",
		Created: 1677652288,
		Model:   "gemini-2.0-flash-lite",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Here's a joke: Why did the chicken cross the road? To get to the other side!",
				},
				FinishReason: "completed",
			},
		},
	}

	if diff := cmp.Diff(expected, res); diff != "" {
		t.Errorf("mismatch (-expected +actual):\n%s", diff)
	}
}

func TestChat_Fail(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	param := &Payload{
		Model: "gemini-2.0-flash-lite",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Tell me a fortune.",
			},
		},
	}
	token := "test-token"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		// Serve a failure response from the testdata file.
		http.ServeFile(w, r, "testdata/gemini_fail.json")
	})

	client, err := NewClient(server.URL, token)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Chat(context.Background(), param)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status code: 404") {
		t.Fatalf("expected error to contain 'status code: 404', got %s", err.Error())
	}
}
