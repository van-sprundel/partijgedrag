package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OllamaClient struct {
	BaseURL string
	Model   string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
	}
}

func (c *OllamaClient) SimplifyCase(title string, bulletPoints []string) (*SimplifiedCase, error) {
	bpJSON, _ := json.Marshal(bulletPoints)

	prompt := fmt.Sprintf(`
Je bent een taalassistent gespecialiseerd in het eenvoudig en duidelijk herschrijven van politieke teksten in het Nederlands.
Gebruik makkelijke taal en korte zinnen, maar behoud de betekenis.
Geef het resultaat terug als JSON met twee velden: "simplified_title" en "simplified_bullet_points".

Titel: %s
Bullet points: %s`, title, string(bpJSON))

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":  c.Model,
		"prompt": prompt,
		"stream": false,
	})

	resp, err := http.Post(fmt.Sprintf("%s/api/generate", c.BaseURL), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ollama response: %w", err)
	}

	// Ollama returns {"response": "..."} where the response is the model’s text output
	var raw struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid Ollama response: %w", err)
	}

	rawOutput := raw.Response

	// Remove ```json ... ``` or ``` ... ```
	rawOutput = strings.TrimSpace(rawOutput)
	if strings.HasPrefix(rawOutput, "```json") {
		rawOutput = strings.TrimPrefix(rawOutput, "```json")
	} else if strings.HasPrefix(rawOutput, "```") {
		rawOutput = strings.TrimPrefix(rawOutput, "```")
	}
	if strings.HasSuffix(rawOutput, "```") {
		rawOutput = strings.TrimSuffix(rawOutput, "```")
	}
	rawOutput = strings.TrimSpace(rawOutput)

	var simplified SimplifiedCase
	if err := json.Unmarshal([]byte(rawOutput), &simplified); err != nil {
		return nil, fmt.Errorf("failed to parse model output: %w\nRaw output: %s", err, raw.Response)
	}

	return &simplified, nil
}
