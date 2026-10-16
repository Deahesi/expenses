package redis_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/storage/redis"
	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *goredis.Client) {
	t.Helper()

	mr := miniredis.RunT(t)

	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	return mr, client
}

func TestMessagesRepository_AddAndGet(t *testing.T) {
	mr, client := newTestRedis(t)
	repo := redis.NewMessagesRepository(client)
	ctx := context.Background()

	userMsg := &domain.Message{TelegramID: 42, Message: "кофе 350", Role: domain.MessageUser}
	appMsg := &domain.Message{TelegramID: 42, Message: "✅ Добавлено: кофе — 350 ₽", Role: domain.MessageApp}

	if err := repo.AddMessage(ctx, userMsg); err != nil {
		t.Fatalf("AddMessage(user) error = %v", err)
	}

	if err := repo.AddMessage(ctx, appMsg); err != nil {
		t.Fatalf("AddMessage(app) error = %v", err)
	}

	messages, err := repo.GetMessages(ctx, 42)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("GetMessages() = %d messages, want 2", len(messages))
	}

	if messages[0] != *userMsg {
		t.Errorf("GetMessages()[0] = %+v, want %+v", messages[0], *userMsg)
	}

	if messages[1] != *appMsg {
		t.Errorf("GetMessages()[1] = %+v, want %+v", messages[1], *appMsg)
	}

	// The history key must be scoped to the telegram user.
	if !mr.Exists("tg:user:42:messages") {
		t.Error("history key tg:user:42:messages was not created")
	}
}

func TestMessagesRepository_IsolatedPerTelegramID(t *testing.T) {
	_, client := newTestRedis(t)
	repo := redis.NewMessagesRepository(client)
	ctx := context.Background()

	if err := repo.AddMessage(ctx, &domain.Message{TelegramID: 1, Message: "первый", Role: domain.MessageUser}); err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	if err := repo.AddMessage(ctx, &domain.Message{TelegramID: 2, Message: "второй", Role: domain.MessageUser}); err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	first, err := repo.GetMessages(ctx, 1)
	if err != nil {
		t.Fatalf("GetMessages(1) error = %v", err)
	}

	if len(first) != 1 || first[0].Message != "первый" {
		t.Errorf("GetMessages(1) = %+v, want only the first user's message", first)
	}

	second, err := repo.GetMessages(ctx, 2)
	if err != nil {
		t.Fatalf("GetMessages(2) error = %v", err)
	}

	if len(second) != 1 || second[0].Message != "второй" {
		t.Errorf("GetMessages(2) = %+v, want only the second user's message", second)
	}
}

func TestMessagesRepository_GetUnknownUser(t *testing.T) {
	_, client := newTestRedis(t)
	repo := redis.NewMessagesRepository(client)

	messages, err := repo.GetMessages(context.Background(), 999)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("GetMessages() = %d messages, want 0", len(messages))
	}
}

// AddMessage keeps only the last 20 messages of the history.
func TestMessagesRepository_TrimsHistory(t *testing.T) {
	_, client := newTestRedis(t)
	repo := redis.NewMessagesRepository(client)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		msg := &domain.Message{
			TelegramID: 7,
			Message:    "сообщение " + string(rune('a'+i)),
			Role:       domain.MessageUser,
		}
		if err := repo.AddMessage(ctx, msg); err != nil {
			t.Fatalf("AddMessage(%d) error = %v", i, err)
		}
	}

	messages, err := repo.GetMessages(ctx, 7)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(messages) != 20 {
		t.Fatalf("GetMessages() = %d messages, want the last 20", len(messages))
	}

	if messages[len(messages)-1].Message != "сообщение y" {
		t.Errorf("last message = %q, want the most recently added one", messages[len(messages)-1].Message)
	}
}

// The history key expires after two hours.
func TestMessagesRepository_ExpireAfterTTL(t *testing.T) {
	mr, client := newTestRedis(t)
	repo := redis.NewMessagesRepository(client)
	ctx := context.Background()

	if err := repo.AddMessage(ctx, &domain.Message{TelegramID: 5, Message: "кофе 350", Role: domain.MessageUser}); err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	mr.FastForward(2*time.Hour + time.Minute)

	messages, err := repo.GetMessages(ctx, 5)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("GetMessages() after TTL = %d messages, want 0", len(messages))
	}
}

func TestMessagesRepository_AddMessageError(t *testing.T) {
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	mr.Close()

	repo := redis.NewMessagesRepository(client)

	err := repo.AddMessage(context.Background(), &domain.Message{TelegramID: 1, Message: "x", Role: domain.MessageUser})
	if err == nil {
		t.Fatal("AddMessage() expected an error against an unavailable Redis, got nil")
	}
}

func TestNewRedisStorage(t *testing.T) {
	t.Run("connects and wires repositories", func(t *testing.T) {
		mr := miniredis.RunT(t)

		storage, err := redis.NewRedisStorage(context.Background(), &config.RedisConfig{RedisAddr: mr.Addr()})
		if err != nil {
			t.Fatalf("NewRedisStorage() error = %v", err)
		}
		defer storage.Close()

		if storage.Repositories.Message == nil {
			t.Error("NewRedisStorage() did not initialize the message repository")
		}
	})

	t.Run("fails when redis is unreachable", func(t *testing.T) {
		mr := miniredis.RunT(t)
		addr := mr.Addr()
		mr.Close()

		storage, err := redis.NewRedisStorage(context.Background(), &config.RedisConfig{RedisAddr: addr})
		if err == nil {
			storage.Close()
			t.Fatal("NewRedisStorage() expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "redis initialization") {
			t.Errorf("NewRedisStorage() error = %q, want it to contain %q", err.Error(), "redis initialization")
		}
	})
}
