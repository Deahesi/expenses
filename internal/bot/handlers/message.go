package handlers

import (
	"context"
	"log/slog"

	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/user"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *MessageHandler) HandleMessage(
	ctx context.Context,
	client *telegram.Bot,
	update *models.Update,
) {
	var response string
	if update.Message == nil || update.Message.From == nil {
		return
	}

	msg := update.Message

	if msg.Text == "" {
		return
	}

	responseMessage, err := client.SendMessage(ctx, &telegram.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   "Сообщение обрабатывается...",
	})

	if err != nil {
		slog.Error("failed to send message", "error", err)
	}

	user, err := h.usersService.UpsertTelegramUser(ctx, user.UpsertTelegramUserInput{
		FirstName:  update.Message.From.FirstName,
		LastName:   update.Message.From.LastName,
		TelegramID: update.Message.From.ID,
	})

	if err != nil {
		slog.Error("failed to process message", "error", err)
		response = "Не удалось обработать сообщение"
	}

	llmResponse, err := h.llmService.SendMessage(ctx, &llm.SendMessageInput{
		Text:   msg.Text,
		UserID: user.ID.String(),
	})

	if err != nil {
		slog.Error("failed to process message", "error", err)
		response = "Не удалось обработать сообщение"
	}
	if err == nil && llmResponse != "" {
		response = llmResponse
	}

	_, err = client.EditMessageText(ctx, &telegram.EditMessageTextParams{
		ChatID:    msg.Chat.ID,
		MessageID: responseMessage.ID,
		Text:      response,
		ParseMode: models.ParseModeHTML,
	})

	if err != nil {
		slog.Error("failed to send message", "error", err)
	}
}
