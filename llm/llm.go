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
		User          User
	}
	User struct {
		Name       string
		Location   string
		Family     string
		Occupation string
		Age        string
		Sex        string
	}
	Prompt struct {
		Model       Model
		MaxTokens   int
		Temperature float64
		History     []Message
		Text        string
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
	if prompt.Model == "" || prompt.MaxTokens == 0 || prompt.Temperature == 0 {
		return Response{}, fmt.Errorf("invalid prompt. model, maxtokens and temperature must be specified")
	}

	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(cfg.SystemPrompt)
	systemPrompt.WriteString(`Your responses are printed to a linux terminal. 
		You will ensure those responses are concise and easily rendered in a linux terminal.
		You will not use markdown syntax in your responses as this is not rendered well in terminal output. 
		However you may use clear, plain text formatting that can be read easily and immediately by a human, such as using dashes for list delimiters. 
		All answers should be factually correct and you should take caution regarding hallucinations. 
		You should only answer the specific question given; do not proactively include additional information that is not directly relevant to the question. 
		`)

	systemPrompt.WriteString(fmt.Sprintf("Your responses must not exceed %v words in length. ", float64(prompt.MaxTokens)*0.75)) // rough mapping of tokens to words

	defineAttribute := func(key string, val any, unset any) string {
		if val == unset {
			return ""
		}
		return fmt.Sprintf("Consider in your responses, where it may be relevant, that the user has provided this information regarding their %v: %v", key, val) + ". "
	}

	systemPrompt.WriteString(defineAttribute("location", cfg.User.Location, ""))
	systemPrompt.WriteString(defineAttribute("name", cfg.User.Name, ""))
	systemPrompt.WriteString(defineAttribute("family", cfg.User.Family, ""))
	systemPrompt.WriteString(defineAttribute("occupation", cfg.User.Occupation, ""))
	systemPrompt.WriteString(defineAttribute("age", cfg.User.Age, 0))
	systemPrompt.WriteString(defineAttribute("sex", cfg.User.Sex, ""))
	systemPrompt.WriteString(defineAttribute("preferred response style; note that this refines your output. It does not override any previous instruction where there is a contradiction", cfg.ResponseStyle, ""))

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

	request := bytes.Buffer{}
	json.NewEncoder(&request).Encode(schema.Request{
		SystemInstruction: schema.SystemInstruction{
			Parts: []schema.Part{{Text: systemPrompt.String()}},
		},
		Contents: content,
		Tools: []schema.Tool{
			{GoogleSearch: &schema.GoogleSearch{}},
		},
		GenerationConfig: schema.GenerationConfig{
			Temperature:     prompt.Temperature,
			MaxOutputTokens: prompt.MaxTokens,
		},
	})

	LogPrintf("request=%q", request.Bytes())

	rs, err := http.Post(fmt.Sprintf(cfg.APIURL, prompt.Model, cfg.APIKey), "application/json", &request)

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
