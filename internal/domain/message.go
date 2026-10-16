package domain

type MessageRole string

const (
	MessageUser MessageRole = "user"
	MessageApp  MessageRole = "app"
)

type Message struct {
	TelegramID int64       `json:"telegram_id"`
	Message    string      `json:"message"`
	Role       MessageRole `json:"role"`
}
