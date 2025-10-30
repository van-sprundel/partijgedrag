package llm

type LLMClient interface {
	SimplifyCase(title string, bulletPoints []string) (*SimplifiedCase, error)
}

type SimplifiedCase struct {
	SimplifiedTitel        string   `json:"simplified_title"`
	SimplifiedBulletPoints []string `json:"simplified_bullet_points"`
}
