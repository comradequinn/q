package llm_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/comradequinn/q/llm"
	"github.com/comradequinn/q/llm/internal/schema"
)

func TestLLM(t *testing.T) {
	actualRq := schema.Request{}
	expectedResponse := schema.Response{
		Candidates: []schema.Candidate{
			{
				Content: schema.Content{
					Role: "model",
					Parts: []schema.Part{
						{Text: "test-response-a"},
						{Text: "test-response-b"},
					},
				},
				FinishReason: schema.FinishReasonStop,
			},
			{
				Content: schema.Content{
					Role: "model",
					Parts: []schema.Part{
						{Text: "test-response-ignore-a"},
						{Text: "test-response-ignore-b"},
					},
				},
				FinishReason: schema.FinishReasonStop,
			},
		},
		UsageMetadata: schema.UsageMetadata{
			TotalTokenCount: 1000,
		},
	}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("model") != llm.ModelGeminiFlash {
			t.Fatalf("expected model to be %v. got %v", llm.ModelGeminiFlash, r.URL.Query()["model"][0])
		}

		if r.URL.Query().Get("api-key") != "test=api-key" {
			t.Fatalf("expected api key to be %v. got %v", "test=api-key", r.URL.Query()["api-key"][0])
		}

		if err := json.NewDecoder(r.Body).Decode(&actualRq); err != nil {
			t.Fatalf("unable to decode llm stub request body. %v", err)
		}

		if err := json.NewEncoder(w).Encode(&expectedResponse); err != nil {
			t.Fatalf("unable to encode llm stub response body. %v", err)
		}
	}))
	defer svr.Close()

	cfg := llm.Config{
		APIKey:        "test=api-key",
		APIURL:        svr.URL + "/test-url/?model=%v&api-key=%v",
		ResponseStyle: "test-style",
		SystemPrompt:  "test-system-prompt",
		User: llm.User{
			Name:       "test-name",
			Location:   "test-location",
			Family:     "test-family",
			Occupation: "test-occupation",
			Age:        "test-age",
			Sex:        "test-sex",
		},
	}

	prompt := llm.Prompt{
		Model:       llm.ModelGeminiFlash,
		MaxTokens:   1000,
		Text:        "test prompt",
		Temperature: 1.0,
		History: []llm.Message{
			{
				Role: llm.RoleUser,
				Text: "test-history-1",
			},
			{
				Role: llm.RoleModel,
				Text: "test-history-2",
			},
		},
	}

	rs, err := llm.Generate(cfg, prompt)

	if err != nil {
		t.Fatalf("expected no error generating response. got %v", err)
	}
	if actualRq.GenerationConfig.MaxOutputTokens != prompt.MaxTokens {
		t.Fatalf("expected max output tokens to be %v. got %v", prompt.MaxTokens, actualRq.GenerationConfig.MaxOutputTokens)
	}
	if actualRq.GenerationConfig.Temperature != prompt.Temperature {
		t.Fatalf("expected temperature to be %v. got %v", prompt.Temperature, actualRq.GenerationConfig.Temperature)
	}

	systemPrompt := actualRq.SystemInstruction.Parts[0].Text

	if !strings.Contains(systemPrompt, cfg.SystemPrompt) {
		t.Fatalf("expected system instruction to contain %v", cfg.SystemPrompt)
	}

	if !strings.Contains(systemPrompt, cfg.User.Name) {
		t.Fatalf("expected system instruction to contain %v", cfg.User.Name)
	}
	if !strings.Contains(systemPrompt, cfg.User.Location) {
		t.Fatalf("expected system instruction to contain %v", cfg.User.Location)
	}
	if !strings.Contains(systemPrompt, cfg.User.Family) {
		t.Fatalf("expected system instruction to contain %v", cfg.User.Family)
	}
	if !strings.Contains(systemPrompt, cfg.User.Occupation) {
		t.Fatalf("expected system instruction to contain %v", cfg.User.Occupation)
	}
	if !strings.Contains(systemPrompt, cfg.User.Age) {
		t.Fatalf("expected system instruction to contain %v", cfg.User.Age)
	}
	if !strings.Contains(systemPrompt, cfg.User.Sex) {
		t.Fatalf("expected system instruction to contain %v", cfg.User.Sex)
	}
	if !strings.Contains(systemPrompt, cfg.ResponseStyle) {
		t.Fatalf("expected system instruction to contain %v", cfg.ResponseStyle)
	}

	if len(actualRq.Contents) != 3 {
		t.Fatalf("expected 1 content. got %v", len(actualRq.Contents))
	}

	for i := range len(prompt.History) {
		if actualRq.Contents[i].Role != string(prompt.History[i].Role) {
			t.Fatalf("expected role to be %v. got %v", llm.RoleUser, actualRq.Contents[0].Role)
		}

		if actualRq.Contents[i].Parts[0].Text != string(prompt.History[i].Text) {
			t.Fatalf("expected text to be %v. got %v", prompt.Text, actualRq.Contents[0].Parts[0].Text)
		}
	}

	if actualRq.Contents[2].Role != string(llm.RoleUser) {
		t.Fatalf("expected role to be %v. got %v", llm.RoleUser, actualRq.Contents[0].Role)
	}

	if actualRq.Contents[2].Parts[0].Text != prompt.Text {
		t.Fatalf("expected text to be %v. got %v", prompt.Text, actualRq.Contents[0].Parts[0].Text)
	}

	if rs.Text != expectedResponse.Candidates[0].Content.Parts[0].Text+expectedResponse.Candidates[0].Content.Parts[1].Text {
		t.Fatalf("expected response text to be %v. got %v", expectedResponse.Candidates[0].Content.Parts[0].Text+expectedResponse.Candidates[0].Content.Parts[1].Text, rs.Text)
	}

	if rs.Tokens != expectedResponse.UsageMetadata.TotalTokenCount {
		t.Fatalf("expected response token count to be %v. got %v", expectedResponse.UsageMetadata.TotalTokenCount, rs.Tokens)
	}
}
