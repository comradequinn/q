package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/comradequinn/q/llm/internal/schema"
)

type (
	Config struct {
		APIKey        string
		APIURL        string
		SystemPrompt  string
		ResponseStyle string
		Model         Model
		MaxTokens     int
		Temperature   float64
		TopP          float64
		User          User
	}
	User struct {
		Name        string
		Location    string
		Description string
	}
	Prompt struct {
		History   []Message
		Text      string
		Schema    string
		Grounding bool
	}
	Response struct {
		Tokens int
		Text   string
	}
	Role    string
	Message struct {
		Role Role
		Text string
	}
	Model string
)

const (
	RoleUser         = "user"
	RoleModel        = "model"
	ModelGeminiPro   = "gemini-2.5-pro-preview-03-25"
	ModelGeminiFlash = "gemini-2.5-flash-preview-04-17"
)

var (
	LogPrintf = func(format string, v ...any) {}
)

// Generate queries the configured LLM with the specified prompt and returns the result
func Generate(cfg Config, prompt Prompt) (Response, error) {
	if cfg.Model == "" || cfg.MaxTokens == 0 || cfg.Temperature == 0 {
		return Response{}, fmt.Errorf("invalid prompt. model, maxtokens and temperature must be specified")
	}

	if prompt.Grounding && prompt.Schema != "" {
		return Response{}, fmt.Errorf("invalid prompt. cannot use grounding with a response schema")
	}

	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(cfg.SystemPrompt + ". ")
	systemPrompt.WriteString(`Your responses are printed to a linux terminal. 
		You will ensure those responses are concise and easily rendered in a linux terminal.
		You will not use markdown syntax in your responses as this is not rendered well in terminal output. 
		However you may use clear, plain text formatting that can be read easily and immediately by a human, such as using dashes for list delimiters. 
		All answers should be factually correct and you should take caution regarding hallucinations. 
		You should only answer the specific question given; do not proactively include additional information that is not directly relevant to the question. 
		`)

	systemPrompt.WriteString(fmt.Sprintf("Your responses must not exceed %v words in length. ", float64(cfg.MaxTokens)*0.75)) // rough mapping of tokens to words

	defineAttribute := func(key string, val any, unset any) string {
		if val == unset {
			return ""
		}
		return fmt.Sprintf("Consider in your responses, where it may be relevant, that the user has provided this information regarding their %v: %q", key, val) + ". "
	}

	systemPrompt.WriteString(defineAttribute("location", cfg.User.Location, ""))
	systemPrompt.WriteString(defineAttribute("name", cfg.User.Name, ""))
	systemPrompt.WriteString(defineAttribute("description", cfg.User.Description, ""))
	systemPrompt.WriteString(defineAttribute("preferred response style; note that this only refines your output and does not override any previous instruction where there is a contradiction", cfg.ResponseStyle, ""))

	content := make([]schema.Content, 0, len(prompt.History)+1)

	for _, message := range prompt.History {
		content = append(content, schema.Content{
			Role: string(message.Role),
			Parts: []schema.Part{
				{Text: message.Text},
			},
		})
	}

	content = append(content, schema.Content{
		Role: RoleUser,
		Parts: []schema.Part{
			{Text: prompt.Text},
		},
	})

	tools := []schema.Tool{}

	if prompt.Grounding {
		tools = []schema.Tool{
			{GoogleSearch: &schema.GoogleSearch{}},
		}
	}

	generationConfig := schema.GenerationConfig{
		Temperature:     cfg.Temperature,
		TopP:            cfg.TopP,
		MaxOutputTokens: cfg.MaxTokens,
	}

	if prompt.Schema != "" {
		generationConfig.ResponseMimeType = "application/json"
		generationConfig.ResponseSchema = json.RawMessage(prompt.Schema)
	}

	request := bytes.Buffer{}
	if err := json.NewEncoder(&request).Encode(schema.Request{
		SystemInstruction: schema.SystemInstruction{
			Parts: []schema.Part{{Text: systemPrompt.String()}},
		},
		Contents:         content,
		Tools:            tools,
		GenerationConfig: generationConfig,
	}); err != nil {
		return Response{}, fmt.Errorf("unable to encode llm request as json. %v", err)
	}

	LogPrintf("request=%q", request.Bytes())

	rs, err := http.Post(fmt.Sprintf(cfg.APIURL, cfg.Model, cfg.APIKey), "application/json", &request)

	if err != nil {
		return Response{}, fmt.Errorf("unable to send request to llm api. %v", err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return Response{}, fmt.Errorf("unable to read response body. %v", err)
	}

	LogPrintf("response=%q", string(body))

	if rs.StatusCode != 200 {
		return Response{}, fmt.Errorf("non-200 status code returned from llm api. %s", body)
	}

	response := schema.Response{}

	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("unable to parse response body. %v", err)
	}

	if len(response.Candidates) == 0 || response.Candidates[0].FinishReason != schema.FinishReasonStop {
		return Response{}, fmt.Errorf("no valid response candidates returned. response: %s", body)
	}

	sb := strings.Builder{}

	for _, part := range response.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	return Response{
		Tokens: response.UsageMetadata.TotalTokenCount,
		Text:   sb.String(),
	}, nil
}
