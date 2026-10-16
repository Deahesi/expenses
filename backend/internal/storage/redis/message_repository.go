package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

type MessagesRepository struct {
	rdb *redis.Client
}

func NewMessagesRepository(rdb *redis.Client) *MessagesRepository {
	return &MessagesRepository{
		rdb: rdb,
	}
}

func (r *MessagesRepository) AddMessage(
	ctx context.Context,
	message *domain.Message,
) error {
	key := "tg:user:" + strconv.FormatInt(message.TelegramID, 10) + ":messages"

	payload, err := json.Marshal(message)

	if err != nil {
		return err
	}

	pipe := r.rdb.TxPipeline()
	pipe.RPush(ctx, key, payload)

	//TODO: перевести на env конфиг
	pipe.LTrim(ctx, key, -20, -1)
	pipe.Expire(ctx, key, time.Duration(2*time.Hour))

	_, err = pipe.Exec(ctx)

	return err
}

func (r *MessagesRepository) GetMessages(
	ctx context.Context,
	telegramID int64,
) ([]domain.Message, error) {
	key := "tg:user:" + strconv.FormatInt(telegramID, 10) + ":messages"

	records, err := r.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]domain.Message, 0, len(records))

	for _, record := range records {
		var msg domain.Message
		err = json.Unmarshal([]byte(record), &msg)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}
