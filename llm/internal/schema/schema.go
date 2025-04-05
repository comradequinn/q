package schema

const (
	FinishReasonStop = "STOP"
)

type (
	Request struct {
		SystemInstruction SystemInstruction `json:"system_instruction"`
		Contents          []Content         `json:"contents"`
		Tools             []Tool            `json:"tools"`
		GenerationConfig  GenerationConfig  `json:"generationConfig"`
	}
	SystemInstruction struct {
		Parts []Part `json:"parts"`
	}
	GoogleSearch struct{}
	Tool         struct {
		GoogleSearch *GoogleSearch `json:"googleSearch,omitempty"`
	}
	GenerationConfig struct {
		Temperature     float64 `json:"temperature"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	}
)

type (
	Response struct {
		Candidates    []Candidate   `json:"candidates"`
		UsageMetadata UsageMetadata `json:"usageMetadata"`
	}
	UsageMetadata struct {
		TotalTokenCount int `json:"totalTokenCount"`
	}
)

type (
	Part struct {
		Text string `json:"text"`
	}
	Content struct {
		Role  string `json:"role"`
		Parts []Part `json:"parts"`
	}
	Candidate struct {
		Content      Content `json:"content"`
		FinishReason string  `json:"finishReason"`
	}
)
