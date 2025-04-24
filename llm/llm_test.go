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
		if r.URL.Query().Get("model") != llm.Models.Flash {
			t.Fatalf("expected model to be %v. got %v", llm.Models.Flash, r.URL.Query()["model"][0])
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
		Model:         llm.Models.Flash,
		MaxTokens:     1000,
		Temperature:   1.0,
		TopP:          1.0,
		User: llm.User{
			Name:        "test-name",
			Location:    "test-location",
			Description: "test-description",
		},
		DebugPrintf: func(string, ...any) {},
	}

	prompt := llm.Prompt{
		Text:      "test prompt",
		Grounding: true,
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

	assert := func(t *testing.T, condition bool, format string, v ...any) {
		if !condition {
			t.Fatalf(format, v...)
		}
	}

	assertResponse := func(t *testing.T, rs llm.Response, err error) {
		assert(t, err == nil, "expected no error generating response. got %v", err)
		assert(t, actualRq.GenerationConfig.MaxOutputTokens == cfg.MaxTokens, "expected max output tokens to be %v. got %v", cfg.MaxTokens, actualRq.GenerationConfig.MaxOutputTokens)
		assert(t, actualRq.GenerationConfig.Temperature == cfg.Temperature, "expected temperature to be %v. got %v", cfg.Temperature, actualRq.GenerationConfig.Temperature)
		assert(t, actualRq.GenerationConfig.TopP == cfg.TopP, "expected top-p to be %v. got %v", cfg.TopP, actualRq.GenerationConfig.TopP)

		if prompt.Grounding {
			assert(t, len(actualRq.Tools) == 1 && actualRq.Tools[0].GoogleSearch != nil, "expected 1 tool of type google-search to be specified when grounding enabled. got %v", len(actualRq.Tools))
		} else {
			assert(t, len(actualRq.Tools) == 0, "expected 0 tools to be specified when grounding disabled. got %v", len(actualRq.Tools))
		}

		if prompt.Schema != "" {
			assert(t, actualRq.GenerationConfig.ResponseMimeType == "application/json", "expected response mime type to be application/json when a response schema is specified. got %v")
			data, _ := actualRq.GenerationConfig.ResponseSchema.MarshalJSON()
			assert(t, string(data) == prompt.Schema, "expected response schema to be %v. got %v", prompt.Schema, string(data))
		} else {
			assert(t, actualRq.GenerationConfig.ResponseMimeType == "", "expected response mime type to be empty string when no response schema specified. got %v")
		}

		systemPrompt := actualRq.SystemInstruction.Parts[0].Text

		assert(t, strings.Contains(systemPrompt, cfg.SystemPrompt), "expected system prompt %q to contain %q", systemPrompt, cfg.SystemPrompt)
		assert(t, strings.Contains(systemPrompt, cfg.User.Name), "expected system prompt to contain %v", cfg.User.Name)
		assert(t, strings.Contains(systemPrompt, cfg.User.Location), "expected system prompt to contain %v", cfg.User.Location)
		assert(t, strings.Contains(systemPrompt, cfg.User.Description), "expected system prompt to contain %v", cfg.User.Description)
		assert(t, strings.Contains(systemPrompt, cfg.ResponseStyle), "expected system prompt to contain %v", cfg.ResponseStyle)
		assert(t, len(actualRq.Contents) == 3, "expected 3 content entries. got %v", len(actualRq.Contents))

		for i := range len(prompt.History) {
			assert(t, actualRq.Contents[i].Role == string(prompt.History[i].Role), "expected role to be %v. got %v", llm.RoleUser, actualRq.Contents[0].Role)
			assert(t, actualRq.Contents[i].Parts[0].Text == string(prompt.History[i].Text), "expected text to be %v. got %v", prompt.Text, actualRq.Contents[0].Parts[0].Text)
		}

		assert(t, actualRq.Contents[2].Role == string(llm.RoleUser), "expected role to be %v. got %v", llm.RoleUser, actualRq.Contents[0].Role)
		assert(t, actualRq.Contents[2].Parts[0].Text == prompt.Text, "expected text to be %v. got %v", prompt.Text, actualRq.Contents[0].Parts[0].Text)
		assert(t, rs.Text == expectedResponse.Candidates[0].Content.Parts[0].Text+expectedResponse.Candidates[0].Content.Parts[1].Text, "expected response text to be %v. got %v", expectedResponse.Candidates[0].Content.Parts[0].Text+expectedResponse.Candidates[0].Content.Parts[1].Text, rs.Text)
		assert(t, rs.Tokens == expectedResponse.UsageMetadata.TotalTokenCount, "expected response token count to be %v. got %v", expectedResponse.UsageMetadata.TotalTokenCount, rs.Tokens)
	}

	rs, err := llm.Generate(cfg, prompt)

	assertResponse(t, rs, err)

	prompt.Grounding = false
	prompt.Schema = `{"type":"object","properties":{"response":{"type":"string"}}}`

	rs, err = llm.Generate(cfg, prompt)

	assertResponse(t, rs, err)
}
