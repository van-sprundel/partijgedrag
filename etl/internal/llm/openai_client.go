package llm

import (
	"context"
	"encoding/json"
	"fmt"
	openai "github.com/sashabaranov/go-openai" // lightweight community client
	"strings"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *OpenAIClient) SimplifyCase(title string, bulletPoints []string) (*SimplifiedCase, error) {
	bpJSON, _ := json.Marshal(bulletPoints)
	prompt := fmt.Sprintf(`
Je bent een taalassistent gespecialiseerd in het eenvoudig en duidelijk herschrijven van politieke teksten in het Nederlands.
Gebruik makkelijke taal en korte zinnen, maar behoud de betekenis.
Geef het resultaat terug als JSON met twee velden: "simplified_title" en "simplified_bullet_points".

Titel: %s
Bullet points: %s`, title, string(bpJSON))

	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{{
			Role:    "user",
			Content: prompt,
		}},
	})
	if err != nil {
		return nil, err
	}
	rawOutput := resp.Choices[0].Message.Content
	//fmt.Printf("\n🔹 Model response:\n%s\n", resp.Choices[0].Message.Content)
	//if (resp.Usage != openai.Usage{}) {
	//	fmt.Printf("🔸 Tokens used — Prompt: %d, Completion: %d, Total: %d\n",
	//		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	//} else {
	//	fmt.Println("⚠️ No token usage info returned by API")
	//}

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
		return nil, fmt.Errorf("failed to parse model output: %w\nRaw output: %s", err, rawOutput)
	}

	// Add "vz: " prefix to simplified bullet points where original bullet starts with "verzoekt"
	for i, bp := range bulletPoints {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(bp)), "verzoekt") {
			if i < len(simplified.SimplifiedBulletPoints) && !strings.HasPrefix(simplified.SimplifiedBulletPoints[i], "vz: ") {
				simplified.SimplifiedBulletPoints[i] = "vz: " + simplified.SimplifiedBulletPoints[i]
			}
		}
	}

	return &simplified, nil
}
