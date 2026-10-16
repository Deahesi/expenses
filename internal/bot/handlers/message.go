package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Deahesi/expenses-backend/internal/domain"
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
	h.handleMessage(ctx, client, update)
}

func (h *MessageHandler) handleMessage(
	ctx context.Context,
	client messageClient,
	update *models.Update,
) {
	var response string
	if update == nil || update.Message == nil || update.Message.From == nil {
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
		return
	}

	user, err := h.usersService.UpsertTelegramUser(ctx, user.UpsertTelegramUserInput{
		FirstName:  update.Message.From.FirstName,
		LastName:   update.Message.From.LastName,
		TelegramID: update.Message.From.ID,
	})
	if err != nil {
		slog.Error("failed to process message", "error", err)
		response = "Не удалось обработать сообщение"
		if _, err := sendResponse(ctx, client, msg.Chat.ID, responseMessage.ID, response); err != nil {
			slog.Error("failed to send response", "error", err)
		}
		return
	}

	messages, err := h.messagesService.GetMessages(ctx, update.Message.From.ID)
	if err != nil {
		slog.Error("failed to get messages history", "error", err)
	}

	llmResponse, err := h.llmService.SendMessage(ctx, &llm.SendMessageInput{
		Text:     msg.Text,
		UserID:   user.ID.String(),
		Messages: messages,
	})

	response = "Не удалось получить ответ. Перед повтором проверьте список трат."
	if err != nil {
		slog.Error("failed to process message", "error", err)
	} else if strings.TrimSpace(llmResponse) != "" {
		response = llmResponse
	}

	response, err = sendResponse(ctx, client, msg.Chat.ID, responseMessage.ID, response)
	if err != nil {
		slog.Error("failed to send response", "error", err)
		return
	}

	err = h.messagesService.AddMessage(ctx, &domain.Message{
		TelegramID: update.Message.From.ID,
		Message:    msg.Text,
		Role:       domain.MessageUser,
	})
	if err != nil {
		slog.Error("failed to save user message", "error", err)
	}
	err = h.messagesService.AddMessage(ctx, &domain.Message{
		TelegramID: update.Message.From.ID,
		Message:    response,
		Role:       domain.MessageApp,
	})
	if err != nil {
		slog.Error("failed to save app message", "error", err)
	}
}
