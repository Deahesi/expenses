package handlers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/llm"
	"github.com/Deahesi/expenses-backend/internal/user"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type fakeUserService struct {
	record domain.User
	input  user.UpsertTelegramUserInput
	err    error
}

func (s *fakeUserService) UpsertTelegramUser(_ context.Context, input user.UpsertTelegramUserInput) (domain.User, error) {
	s.input = input
	return s.record, s.err
}

type fakeLLMService struct {
	input    *llm.SendMessageInput
	response string
	err      error
	calls    int
}

func (s *fakeLLMService) SendMessage(_ context.Context, input *llm.SendMessageInput) (string, error) {
	s.input = input
	s.calls++
	return s.response, s.err
}

type fakeMessagesService struct {
	history    []domain.Message
	saved      []domain.Message
	telegramID int64
	getErr     error
}

func (s *fakeMessagesService) GetMessages(_ context.Context, telegramID int64) ([]domain.Message, error) {
	s.telegramID = telegramID
	return s.history, s.getErr
}

func (s *fakeMessagesService) AddMessage(_ context.Context, message *domain.Message) error {
	s.saved = append(s.saved, *message)
	return nil
}

func textUpdate(text string) *models.Update {
	return &models.Update{
		Message: &models.Message{
			From: &models.User{ID: 42, FirstName: "Иван", LastName: "Петров"},
			Chat: models.Chat{ID: 99},
			Text: text,
		},
	}
}

func TestHandleMessagePassesSenderAndHistoryToLLM(t *testing.T) {
	owner := uuid.New()
	users := &fakeUserService{record: domain.User{ID: owner, TelegramID: 42}}
	model := &fakeLLMService{response: "Добавлено: кофе — <b>350 ₽</b>"}
	history := &fakeMessagesService{
		history: []domain.Message{
			{TelegramID: 42, Role: domain.MessageUser, Message: "кофе"},
			{TelegramID: 42, Role: domain.MessageApp, Message: "Сколько стоил кофе?"},
		},
	}
	client := &fakeMessageClient{}
	handler := NewMessageHandler(users, nil, model, history)

	handler.handleMessage(context.Background(), client, textUpdate("350"))

	if users.input.TelegramID != 42 || users.input.FirstName != "Иван" {
		t.Fatalf("unexpected sender: %+v", users.input)
	}
	if model.input == nil || model.input.UserID != owner.String() || model.input.Text != "350" {
		t.Fatalf("unexpected LLM input: %+v", model.input)
	}
	if history.telegramID != 42 || !reflect.DeepEqual(model.input.Messages, history.history) {
		t.Error("conversation history was loaded for the wrong user or lost")
	}
	if len(client.edits) != 1 || client.edits[0].ChatID != int64(99) {
		t.Fatal("response was not delivered to the original chat")
	}
	if len(history.saved) != 2 {
		t.Fatalf("saved %d messages, want the request and delivered response", len(history.saved))
	}
	if history.saved[0].Role != domain.MessageUser || history.saved[0].Message != "350" {
		t.Errorf("unexpected saved request: %+v", history.saved[0])
	}
	if history.saved[1].Role != domain.MessageApp || history.saved[1].Message != model.response {
		t.Errorf("unexpected saved response: %+v", history.saved[1])
	}
}

func TestHandleMessageStopsWhenUserCannotBeRegistered(t *testing.T) {
	users := &fakeUserService{err: errors.New("database unavailable")}
	model := &fakeLLMService{}
	history := &fakeMessagesService{}
	client := &fakeMessageClient{}
	handler := NewMessageHandler(users, nil, model, history)

	handler.handleMessage(context.Background(), client, textUpdate("кофе 350"))

	if model.calls != 0 || len(history.saved) != 0 {
		t.Error("processing continued after registration failed")
	}
	if len(client.edits) != 1 || client.edits[0].Text != "Не удалось обработать сообщение" {
		t.Error("registration error was not explained to the user")
	}
}

func TestHandleMessageExplainsLLMFailure(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{
			name: "LLM error",
			err:  errors.New("provider unavailable"),
		},
		{
			name: "empty response",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUserService{record: domain.User{ID: uuid.New()}}
			model := &fakeLLMService{err: tt.err}
			history := &fakeMessagesService{}
			client := &fakeMessageClient{}
			handler := NewMessageHandler(users, nil, model, history)

			handler.handleMessage(context.Background(), client, textUpdate("кофе 350"))

			if len(client.edits) != 1 || !strings.Contains(client.edits[0].Text, "проверьте список трат") {
				t.Fatal("an uncertain write must not be confirmed as successful")
			}
			if len(history.saved) != 2 || history.saved[1].Message != client.edits[0].Text {
				t.Error("history must contain the error actually shown to the user")
			}
		})
	}
}

func TestHandleMessageDoesNotSaveUndeliveredResponse(t *testing.T) {
	deliveryErr := errors.New("Telegram unavailable")
	users := &fakeUserService{record: domain.User{ID: uuid.New()}}
	model := &fakeLLMService{response: "Добавлено: кофе — 350 ₽"}
	history := &fakeMessagesService{}
	client := &fakeMessageClient{
		editErrors: []error{deliveryErr, deliveryErr},
		sendErrors: []error{nil, deliveryErr},
	}
	handler := NewMessageHandler(users, nil, model, history)

	handler.handleMessage(context.Background(), client, textUpdate("кофе 350"))

	if model.calls != 1 || len(client.sends) != 2 {
		t.Fatal("expected one LLM call and a fallback delivery attempt")
	}
	if len(history.saved) != 0 {
		t.Error("an undelivered answer was saved as part of the dialogue")
	}
}

func TestHandleMessageContinuesWithoutCachedHistory(t *testing.T) {
	users := &fakeUserService{record: domain.User{ID: uuid.New()}}
	model := &fakeLLMService{response: "Трат пока нет"}
	history := &fakeMessagesService{getErr: errors.New("Redis unavailable")}
	handler := NewMessageHandler(users, nil, model, history)

	handler.handleMessage(context.Background(), &fakeMessageClient{}, textUpdate("покажи траты"))

	if model.calls != 1 || len(model.input.Messages) != 0 {
		t.Error("cache failure should allow a fresh database-backed request")
	}
}

func TestHandleMessageIgnoresNonTextUpdates(t *testing.T) {
	handler := &MessageHandler{}
	client := &fakeMessageClient{}
	for _, update := range []*models.Update{
		nil,
		{},
		{Message: &models.Message{Text: "кофе 350"}},
		{Message: &models.Message{From: &models.User{ID: 42}}},
	} {
		handler.handleMessage(context.Background(), client, update)
	}
	if len(client.sends) != 0 || len(client.edits) != 0 {
		t.Error("unsupported updates must not start expense processing")
	}
}
