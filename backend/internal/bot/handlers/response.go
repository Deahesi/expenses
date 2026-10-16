package handlers

import (
	"context"
	"fmt"
	"html"
	"strings"
	"unicode/utf8"

	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const maxMessageLength = 4096

type messageClient interface {
	SendMessage(context.Context, *telegram.SendMessageParams) (*models.Message, error)
	EditMessageText(context.Context, *telegram.EditMessageTextParams) (*models.Message, error)
}

// sendResponse returns the text actually delivered, for conversation history.
func sendResponse(
	ctx context.Context,
	client messageClient,
	chatID int64,
	messageID int,
	text string,
) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty bot response")
	}

	if utf8.RuneCountInString(text) <= maxMessageLength {
		_, err := client.EditMessageText(ctx, &telegram.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		})
		if err == nil {
			return text, nil
		}
	}

	plainText := html.UnescapeString(strings.NewReplacer(
		"<b>", "", "</b>", "",
		"<i>", "", "</i>", "",
		"<code>", "", "</code>", "",
	).Replace(text))
	if strings.TrimSpace(plainText) == "" {
		return "", fmt.Errorf("empty bot response")
	}

	runes := []rune(plainText)
	for start := 0; start < len(runes); start += maxMessageLength {
		chunk := string(runes[start:min(start+maxMessageLength, len(runes))])
		if start == 0 {
			_, err := client.EditMessageText(ctx, &telegram.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: messageID,
				Text:      chunk,
			})
			if err == nil {
				continue
			}
		}

		if _, err := client.SendMessage(ctx, &telegram.SendMessageParams{
			ChatID: chatID,
			Text:   chunk,
		}); err != nil {
			return "", fmt.Errorf("send response: %w", err)
		}
	}

	return plainText, nil
}
