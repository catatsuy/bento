package openai_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	. "github.com/catatsuy/bento/internal/openai"
	"github.com/google/go-cmp/cmp"
)

func TestPostText_Success(t *testing.T) {
	muxAPI := http.NewServeMux()
	testAPIServer := httptest.NewServer(muxAPI)
	defer testAPIServer.Close()

	param := &Payload{
		Model: "gpt-3.5-turbo",
		Input: "Hello. I am a student.",
	}

	token := "token"

	muxAPI.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		expectedType := "application/json"
		if contentType != expectedType {
			t.Fatalf("Content-Type expected %s, but %s", expectedType, contentType)
		}

		authorization := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + token
		if authorization != expectedAuth {
			t.Fatalf("Authorization expected %s, but %s", expectedAuth, authorization)
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		actualBody := &Payload{}
		err = json.Unmarshal(bodyBytes, actualBody)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(actualBody, param) {
			t.Fatalf("expected %q to equal %q", actualBody, param)
		}

		http.ServeFile(w, r, "testdata/chat_ok.json")
	})

	c, err := NewClient(testAPIServer.URL, token)
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Chat(t.Context(), param)
	if err != nil {
		t.Fatal(err)
	}

	expected := &Response{
		ID: "resp_67b73f697ba4819183a15cc17d011509",
		Output: []OutputMessage{
			{
				ID:   "msg_67b73f697ba4819183a15cc17d011509",
				Type: "message",
				Role: "assistant",
				Content: []Content{
					{
						Type:        "text",
						Text:        "\n\nHello there, how may I assist you today?",
						Annotations: []string{},
					},
				},
			},
		},
		Usage: Usage{
			PromptTokens:     9,
			CompletionTokens: 12,
			TotalTokens:      21,
		},
	}

	if diff := cmp.Diff(res, expected); diff != "" {
		t.Errorf("file list mismatch (-actual +expected):\n%s", diff)
	}
}

func TestPostText_Fail(t *testing.T) {
	muxAPI := http.NewServeMux()
	testAPIServer := httptest.NewServer(muxAPI)
	defer testAPIServer.Close()

	param := &Payload{
		Model: "gpt-3",
		Input: "Hello. I am a student.",
	}

	muxAPI.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, "testdata/chat_fail.json")
	})

	c, err := NewClient(testAPIServer.URL, "token")
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Chat(t.Context(), param)

	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "status code: 404"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q to contain %q", err.Error(), expected)
	}
}
