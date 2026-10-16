package messages_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/messages"
)

// fakeRepository is an in-memory messages.MessagesRepository implementation.
type fakeRepository struct {
	history map[int64][]domain.Message
	addErr  error
	getErr  error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{history: make(map[int64][]domain.Message)}
}

func (r *fakeRepository) AddMessage(_ context.Context, message *domain.Message) error {
	if r.addErr != nil {
		return r.addErr
	}

	r.history[message.TelegramID] = append(r.history[message.TelegramID], *message)

	return nil
}

func (r *fakeRepository) GetMessages(_ context.Context, telegramID int64) ([]domain.Message, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}

	messages := make([]domain.Message, 0, len(r.history[telegramID]))
	messages = append(messages, r.history[telegramID]...)

	return messages, nil
}

func TestMessagesService_AddAndGet(t *testing.T) {
	repo := newFakeRepository()
	service := messages.NewMessagesService(repo)
	ctx := context.Background()

	userMsg := &domain.Message{TelegramID: 42, Message: "кофе 350", Role: domain.MessageUser}
	appMsg := &domain.Message{TelegramID: 42, Message: "✅ Добавлено: кофе — 350 ₽", Role: domain.MessageApp}

	if err := service.AddMessage(ctx, userMsg); err != nil {
		t.Fatalf("AddMessage(user) error = %v", err)
	}

	if err := service.AddMessage(ctx, appMsg); err != nil {
		t.Fatalf("AddMessage(app) error = %v", err)
	}

	got, err := service.GetMessages(ctx, 42)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	want := []domain.Message{*userMsg, *appMsg}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMessages() = %+v, want %+v", got, want)
	}
}

func TestMessagesService_IsolatedPerTelegramID(t *testing.T) {
	repo := newFakeRepository()
	service := messages.NewMessagesService(repo)
	ctx := context.Background()

	if err := service.AddMessage(ctx, &domain.Message{TelegramID: 1, Message: "первый", Role: domain.MessageUser}); err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	if err := service.AddMessage(ctx, &domain.Message{TelegramID: 2, Message: "второй", Role: domain.MessageUser}); err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	first, err := service.GetMessages(ctx, 1)
	if err != nil {
		t.Fatalf("GetMessages(1) error = %v", err)
	}

	if len(first) != 1 || first[0].Message != "первый" {
		t.Errorf("GetMessages(1) = %+v, want only the first user's message", first)
	}

	second, err := service.GetMessages(ctx, 2)
	if err != nil {
		t.Fatalf("GetMessages(2) error = %v", err)
	}

	if len(second) != 1 || second[0].Message != "второй" {
		t.Errorf("GetMessages(2) = %+v, want only the second user's message", second)
	}
}

func TestMessagesService_ErrorsAreWrapped(t *testing.T) {
	sentinel := errors.New("storage unavailable")
	ctx := context.Background()

	t.Run("add message", func(t *testing.T) {
		service := messages.NewMessagesService(&fakeRepository{addErr: sentinel})

		err := service.AddMessage(ctx, &domain.Message{TelegramID: 1, Message: "x", Role: domain.MessageUser})
		if !errors.Is(err, sentinel) {
			t.Errorf("AddMessage() error = %v, want it to wrap the repository error", err)
		}
	})

	t.Run("get messages", func(t *testing.T) {
		service := messages.NewMessagesService(&fakeRepository{getErr: sentinel})

		_, err := service.GetMessages(ctx, 1)
		if !errors.Is(err, sentinel) {
			t.Errorf("GetMessages() error = %v, want it to wrap the repository error", err)
		}
	})
}
