package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type fakeMessageClient struct {
	edits      []*telegram.EditMessageTextParams
	sends      []*telegram.SendMessageParams
	editErrors []error
	sendErrors []error
}

func (c *fakeMessageClient) EditMessageText(_ context.Context, params *telegram.EditMessageTextParams) (*models.Message, error) {
	c.edits = append(c.edits, params)
	index := len(c.edits) - 1
	if index < len(c.editErrors) && c.editErrors[index] != nil {
		return nil, c.editErrors[index]
	}
	return &models.Message{ID: params.MessageID}, nil
}

func (c *fakeMessageClient) SendMessage(_ context.Context, params *telegram.SendMessageParams) (*models.Message, error) {
	c.sends = append(c.sends, params)
	index := len(c.sends) - 1
	if index < len(c.sendErrors) && c.sendErrors[index] != nil {
		return nil, c.sendErrors[index]
	}
	return &models.Message{ID: 10}, nil
}

func TestSendResponse(t *testing.T) {
	editErr := errors.New("cannot edit message")
	for _, tt := range []struct {
		name       string
		text       string
		editErrors []error
		sendErr    error
		want       string
		wantEdits  int
		wantSends  int
		wantErr    bool
	}{
		{
			name:      "HTML response",
			text:      "<b>350 ₽</b>",
			want:      "<b>350 ₽</b>",
			wantEdits: 1,
		},
		{
			name:       "plain fallback",
			text:       "<b>350 ₽</b> &amp; кофе",
			editErrors: []error{editErr},
			want:       "350 ₽ & кофе",
			wantEdits:  2,
		},
		{
			name:       "new message fallback",
			text:       "<b>350 ₽</b>",
			editErrors: []error{editErr, editErr},
			want:       "350 ₽",
			wantEdits:  2,
			wantSends:  1,
		},
		{
			name:       "delivery failure",
			text:       "350 ₽",
			editErrors: []error{editErr, editErr},
			sendErr:    errors.New("send failed"),
			wantEdits:  2,
			wantSends:  1,
			wantErr:    true,
		},
		{
			name:    "empty response",
			text:    "  ",
			wantErr: true,
		},
		{
			name:      "long Unicode response",
			text:      strings.Repeat("я", maxMessageLength+1),
			want:      strings.Repeat("я", maxMessageLength+1),
			wantEdits: 1,
			wantSends: 1,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeMessageClient{editErrors: tt.editErrors, sendErrors: []error{tt.sendErr}}
			got, err := sendResponse(context.Background(), client, 42, 7, tt.text)
			if (err != nil) != tt.wantErr {
				t.Fatalf("sendResponse() error = %v, want error = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("delivered response differs from expected text")
			}
			if len(client.edits) != tt.wantEdits || len(client.sends) != tt.wantSends {
				t.Fatalf("edits/sends = %d/%d, want %d/%d", len(client.edits), len(client.sends), tt.wantEdits, tt.wantSends)
			}
			for i, edit := range client.edits {
				if edit.ChatID != int64(42) || edit.MessageID != 7 {
					t.Errorf("edit targets an unexpected chat or message")
				}
				if i > 0 && edit.ParseMode != "" {
					t.Error("fallback must not use HTML parsing")
				}
			}
			for _, send := range client.sends {
				if send.ChatID != int64(42) || send.ParseMode != "" {
					t.Error("fallback must send plain text to the original chat")
				}
			}
			if len(client.sends) > 0 && tt.sendErr == nil {
				var delivered strings.Builder
				if len(client.edits) > len(tt.editErrors) {
					delivered.WriteString(client.edits[len(client.edits)-1].Text)
				}
				for _, send := range client.sends {
					delivered.WriteString(send.Text)
					if utf8.RuneCountInString(send.Text) > maxMessageLength {
						t.Error("fallback chunk exceeds Telegram's message length")
					}
				}
				if delivered.String() != tt.want {
					t.Error("fallback lost or duplicated response text")
				}
			}
		})
	}
}

type failingHTTPClient struct{}

func (failingHTTPClient) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("Telegram unavailable")
}

func TestHandleMessageStopsWhenInitialSendFails(t *testing.T) {
	client, err := telegram.New("test-token", telegram.WithSkipGetMe(), telegram.WithHTTPClient(time.Second, failingHTTPClient{}))
	if err != nil {
		t.Fatalf("telegram.New() error = %v", err)
	}

	// Nil services would panic if processing continued after the failed send.
	handler := &MessageHandler{}
	handler.HandleMessage(context.Background(), client, &models.Update{
		Message: &models.Message{
			From: &models.User{ID: 42},
			Chat: models.Chat{ID: 42},
			Text: "кофе 350",
		},
	})
}
