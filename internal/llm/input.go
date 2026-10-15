package llm

type SendMessageInput struct {
	Text   string `json:"text"`
	UserID string `json:"user_id"`
}

type SendMessageOutput struct {
	Text string `json:"text"`
}
