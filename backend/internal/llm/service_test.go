package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/expenses-backend/internal/config"
	"github.com/Deahesi/expenses-backend/internal/domain"
	"github.com/Deahesi/expenses-backend/internal/expense"
	"github.com/Deahesi/expenses-backend/internal/llm/tools"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/google/uuid"
)

type fakeExpenseRepository struct {
	expenses    map[uuid.UUID]domain.Expense
	createCalls int
	listCalls   int
	err         error
}

func (r *fakeExpenseRepository) Create(_ context.Context, record domain.Expense) (domain.Expense, error) {
	r.createCalls++
	if r.err != nil {
		return domain.Expense{}, r.err
	}

	record.CreatedAt = time.Now()
	record.UpdatedAt = record.CreatedAt
	r.expenses[record.ID] = record
	return record, nil
}

func (r *fakeExpenseRepository) Get(_ context.Context, id, userID uuid.UUID) (domain.Expense, error) {
	record, ok := r.expenses[id]
	if !ok || record.UserID != userID {
		return domain.Expense{}, domain.ErrExpenseNotFound
	}
	return record, nil
}

func (r *fakeExpenseRepository) GetAll(ctx context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	return r.List(ctx, filter)
}

func (r *fakeExpenseRepository) GetWithUser(context.Context, uuid.UUID, uuid.UUID) (domain.ExpenseWithUser, error) {
	return domain.ExpenseWithUser{}, errors.New("unexpected GetWithUser call")
}

func (r *fakeExpenseRepository) GetAllWithUser(context.Context, expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	return nil, errors.New("unexpected GetAllWithUser call")
}

func (r *fakeExpenseRepository) ListWithUser(context.Context, expense.ListFilter) ([]domain.ExpenseWithUser, error) {
	return nil, errors.New("unexpected ListWithUser call")
}

func (r *fakeExpenseRepository) List(_ context.Context, filter expense.ListFilter) ([]domain.Expense, error) {
	r.listCalls++
	if r.err != nil {
		return nil, r.err
	}

	result := make([]domain.Expense, 0)
	for _, record := range r.expenses {
		if filter.UserID != uuid.Nil && record.UserID != filter.UserID {
			continue
		}
		if filter.Category != "" && string(record.Category) != filter.Category {
			continue
		}
		if !filter.DateFrom.IsZero() && record.SpentAt.Before(filter.DateFrom) {
			continue
		}
		if !filter.DateTo.IsZero() && record.SpentAt.After(filter.DateTo) {
			continue
		}
		result = append(result, record)
	}
	return result, nil
}

func (r *fakeExpenseRepository) Update(ctx context.Context, record domain.Expense) (domain.Expense, error) {
	existing, err := r.Get(ctx, record.ID, record.UserID)
	if err != nil {
		return domain.Expense{}, err
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now()
	r.expenses[record.ID] = record
	return record, nil
}

func (r *fakeExpenseRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := r.Get(ctx, id, userID); err != nil {
		return err
	}
	delete(r.expenses, id)
	return nil
}

func newTestLLM(t *testing.T, repo *fakeExpenseRepository, model ai.ModelFunc) *LLMService {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	g := genkit.Init(ctx, genkit.WithDefaultModel("test/expenses"))
	genkit.DefineModel(g, "test/expenses", &ai.ModelOptions{
		Supports: &ai.ModelSupports{
			Multiturn:  true,
			SystemRole: true,
			Tools:      true,
			ToolChoice: true,
		},
	}, model)

	service := &LLMService{
		genkit:         g,
		expenseService: expense.NewExpenseService(repo),
		systemPrompt:   "Use tools to read or change expenses. Today is {{CURRENT_DATE}} in {{TIMEZONE}}.",
		timezone:       time.UTC,
	}
	service.flow = genkit.DefineFlow(g, "testExpenseRequest", service.generate)
	return service
}

func toolCall(name string, input any) *ai.ModelResponse {
	return &ai.ModelResponse{
		Message: &ai.Message{
			Role: ai.RoleModel,
			Content: []*ai.Part{
				ai.NewToolRequestPart(&ai.ToolRequest{
					Name:  name,
					Input: input,
					Ref:   uuid.NewString(),
				}),
			},
		},
		FinishReason: ai.FinishReasonStop,
	}
}

func modelText(text string) *ai.ModelResponse {
	return &ai.ModelResponse{
		Message:      ai.NewModelTextMessage(text),
		FinishReason: ai.FinishReasonStop,
	}
}

func readToolOutput(t *testing.T, request *ai.ModelRequest, output any) {
	t.Helper()

	message := request.Messages[len(request.Messages)-1]
	if message.Role != ai.RoleTool || len(message.Content) != 1 || message.Content[0].ToolResponse == nil {
		t.Fatal("expected one fresh tool result")
	}
	data, err := json.Marshal(message.Content[0].ToolResponse.Output)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, output); err != nil {
		t.Fatal(err)
	}
}

func TestBuildSystemPromptAcrossMidnight(t *testing.T) {
	loc := time.FixedZone("Asia/Novosibirsk", 7*60*60)
	now := time.Date(2026, 10, 16, 16, 59, 59, 0, time.UTC)
	template := "{{CURRENT_DATE}} {{TIMEZONE}}"

	before := buildSystemPrompt(template, now.In(loc))
	after := buildSystemPrompt(template, now.Add(time.Second).In(loc))
	if before != "2026-10-16 Asia/Novosibirsk" {
		t.Errorf("before midnight = %q", before)
	}
	if after != "2026-10-17 Asia/Novosibirsk" {
		t.Errorf("after midnight = %q", after)
	}
}

func TestNewLLMServiceRejectsMissingConfig(t *testing.T) {
	for _, tt := range []struct {
		name    string
		cfg     config.LLMConfig
		wantErr string
	}{
		{name: "missing key", wantErr: "OPENROUTER_API_KEY"},
		{name: "missing model", cfg: config.LLMConfig{OpenRouterAPIKey: "test-key"}, wantErr: "LLM_MODEL"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLLMService(context.Background(), nil, tt.cfg)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("NewLLMService() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestDomainToGenkitMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []domain.Message
		want     []struct {
			role ai.Role
			text string
		}
	}{
		{
			name:     "nil history",
			messages: nil,
			want:     nil,
		},
		{
			name:     "empty history",
			messages: []domain.Message{},
			want:     nil,
		},
		{
			name: "user and app messages keep order",
			messages: []domain.Message{
				{TelegramID: 1, Message: "кофе 350", Role: domain.MessageUser},
				{TelegramID: 1, Message: "✅ Добавлено: кофе — 350 ₽", Role: domain.MessageApp},
			},
			want: []struct {
				role ai.Role
				text string
			}{
				{role: ai.RoleUser, text: "кофе 350"},
				{role: ai.RoleModel, text: "✅ Добавлено: кофе — 350 ₽"},
			},
		},
		{
			name: "blank messages are skipped",
			messages: []domain.Message{
				{TelegramID: 1, Message: "", Role: domain.MessageUser},
				{TelegramID: 1, Message: "   \n\t ", Role: domain.MessageApp},
				{TelegramID: 1, Message: "кофе 350", Role: domain.MessageUser},
			},
			want: []struct {
				role ai.Role
				text string
			}{
				{role: ai.RoleUser, text: "кофе 350"},
			},
		},
		{
			name: "unknown roles are skipped",
			messages: []domain.Message{
				{TelegramID: 1, Message: "system leak", Role: domain.MessageRole("system")},
				{TelegramID: 1, Message: "", Role: domain.MessageUser},
				{TelegramID: 1, Message: "привет", Role: domain.MessageUser},
			},
			want: []struct {
				role ai.Role
				text string
			}{
				{role: ai.RoleUser, text: "привет"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domainToGenkitMessages(tt.messages)

			if len(got) != len(tt.want) {
				t.Fatalf("domainToGenkitMessages() = %d messages, want %d", len(got), len(tt.want))
			}

			for i, wantMsg := range tt.want {
				if got[i].Role != wantMsg.role {
					t.Errorf("message[%d].Role = %q, want %q", i, got[i].Role, wantMsg.role)
				}

				if len(got[i].Content) != 1 {
					t.Fatalf("message[%d] has %d parts, want 1", i, len(got[i].Content))
				}

				if got[i].Content[0].Text != wantMsg.text {
					t.Errorf("message[%d].Text = %q, want %q", i, got[i].Content[0].Text, wantMsg.text)
				}
			}
		})
	}
}

// The result must always be safe to spread into ai.WithMessages: an empty
// history yields an empty (non-nil is not required, but empty) slice.
func TestDomainToGenkitMessages_EmptyResult(t *testing.T) {
	got := domainToGenkitMessages([]domain.Message{
		{TelegramID: 1, Message: "  ", Role: domain.MessageUser},
	})

	if len(got) != 0 {
		t.Errorf("domainToGenkitMessages() = %d messages, want 0", len(got))
	}

	if got == nil {
		t.Error("domainToGenkitMessages() = nil, want an empty slice")
	}

	if !reflect.DeepEqual(got, []*ai.Message{}) {
		t.Errorf("domainToGenkitMessages() = %#v, want empty []*ai.Message", got)
	}
}

func TestSendMessageCanReplyWithoutTools(t *testing.T) {
	for _, tt := range []struct {
		name  string
		text  string
		reply string
	}{
		{
			name:  "greeting",
			text:  "привет",
			reply: "Привет! Помогу учесть расходы.",
		},
		{
			name:  "thanks",
			text:  "спасибо",
			reply: "Пожалуйста!",
		},
		{
			name:  "help",
			text:  "что ты умеешь?",
			reply: "Могу добавлять, показывать, изменять и удалять ваши траты.",
		},
		{
			name:  "agreement with advice",
			text:  "Да, я так и сделаю",
			reply: "Хорошо!",
		},
		{
			name:  "general advice",
			text:  "как экономить на продуктах?",
			reply: "Составляйте список покупок и сравнивайте цены.",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeExpenseRepository{expenses: make(map[uuid.UUID]domain.Expense)}
			calls := 0
			service := newTestLLM(t, repo, func(
				_ context.Context,
				request *ai.ModelRequest,
				_ ai.ModelStreamCallback,
			) (*ai.ModelResponse, error) {
				calls++
				if request.ToolChoice != ai.ToolChoiceAuto {
					t.Error("ordinary dialogue must allow a reply without tools")
				}
				return modelText(tt.reply), nil
			})

			response, err := service.SendMessage(context.Background(), &SendMessageInput{
				Text:   tt.text,
				UserID: uuid.NewString(),
				Messages: []domain.Message{
					{Role: domain.MessageUser, Message: "кофе 350"},
					{Role: domain.MessageApp, Message: "Кофе 350 ₽ добавлен"},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if response != tt.reply || calls != 1 {
				t.Fatalf("response = %q, model calls = %d", response, calls)
			}
			if repo.createCalls != 0 || repo.listCalls != 0 || len(repo.expenses) != 0 {
				t.Fatal("ordinary dialogue accessed expenses or replayed an old request")
			}
		})
	}
}

func TestSendMessageCreatesExpenseThroughTool(t *testing.T) {
	for _, tt := range []struct {
		name     string
		messages []domain.Message
		existing bool
	}{
		{
			name: "nil history",
		},
		{
			name:     "empty history",
			messages: []domain.Message{},
		},
		{
			name:     "same expense already exists",
			existing: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			owner := uuid.New()
			repo := &fakeExpenseRepository{expenses: make(map[uuid.UUID]domain.Expense)}
			if tt.existing {
				id := uuid.New()
				repo.expenses[id] = domain.Expense{
					ID: id, UserID: owner, Amount: 350, Category: domain.CategoryFood,
					SpentAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), Comment: "кофе",
				}
			}
			wantCount := len(repo.expenses) + 1
			calls := 0
			service := newTestLLM(t, repo, func(
				_ context.Context,
				request *ai.ModelRequest,
				_ ai.ModelStreamCallback,
			) (*ai.ModelResponse, error) {
				calls++
				if calls == 1 {
					return toolCall("create_expense", map[string]any{
						"amount":   350,
						"category": "еда",
						"spent_at": "2026-10-16",
						"comment":  "  кофе  ",
					}), nil
				}

				var output tools.CreateExpenseOutput
				readToolOutput(t, request, &output)
				if !output.Success || output.Expense.UserID != owner.String() || output.Expense.Comment != "кофе" {
					t.Fatalf("unexpected saved expense: %+v", output)
				}
				if _, ok := repo.expenses[uuid.MustParse(output.Expense.ID)]; !ok {
					t.Fatal("tool reported success without saving the expense")
				}
				return modelText("Добавлено: кофе — 350 ₽"), nil
			})

			response, err := service.SendMessage(context.Background(), &SendMessageInput{
				Text:     "кофе 350",
				UserID:   owner.String(),
				Messages: tt.messages,
			})
			if err != nil {
				t.Fatal(err)
			}
			if response != "Добавлено: кофе — 350 ₽" || repo.createCalls != 1 || len(repo.expenses) != wantCount {
				t.Fatalf("unexpected response or create count: %q, %d", response, repo.createCalls)
			}
		})
	}
}

func TestSendMessageReadsDatabaseInsteadOfHistory(t *testing.T) {
	repo := &fakeExpenseRepository{expenses: make(map[uuid.UUID]domain.Expense)}
	calls := 0
	service := newTestLLM(t, repo, func(
		_ context.Context,
		request *ai.ModelRequest,
		_ ai.ModelStreamCallback,
	) (*ai.ModelResponse, error) {
		calls++
		if calls == 1 {
			return toolCall("list_expenses", map[string]any{
				"category":  "",
				"date_from": "",
				"date_to":   "",
			}), nil
		}

		if request.ToolChoice != ai.ToolChoiceAuto {
			t.Error("the model must be able to answer after receiving tool results")
		}
		var output tools.ListExpensesOutput
		readToolOutput(t, request, &output)
		if !output.Success || len(output.Expenses) != 0 {
			t.Fatalf("expected an empty database result: %+v", output)
		}
		return modelText("Трат пока нет"), nil
	})

	response, err := service.SendMessage(context.Background(), &SendMessageInput{
		Text:   "покажи мои траты",
		UserID: uuid.NewString(),
		Messages: []domain.Message{
			{Role: domain.MessageUser, Message: "несуществующая трата 350"},
			{Role: domain.MessageApp, Message: "Добавлена несуществующая трата 350 ₽"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response != "Трат пока нет" || repo.listCalls != 1 {
		t.Fatalf("unexpected response or read count: %q, %d", response, repo.listCalls)
	}
}

func TestSendMessageReportsFailedWrite(t *testing.T) {
	repo := &fakeExpenseRepository{
		expenses: make(map[uuid.UUID]domain.Expense),
		err:      errors.New("database unavailable"),
	}
	calls := 0
	service := newTestLLM(t, repo, func(
		_ context.Context,
		request *ai.ModelRequest,
		_ ai.ModelStreamCallback,
	) (*ai.ModelResponse, error) {
		calls++
		if calls == 1 {
			return toolCall("create_expense", map[string]any{
				"amount":   350,
				"category": "еда",
				"spent_at": "2026-10-16",
				"comment":  "кофе",
			}), nil
		}

		var output tools.CreateExpenseOutput
		readToolOutput(t, request, &output)
		if output.Success {
			t.Fatal("failed write was reported as successful")
		}
		return modelText("Не удалось сохранить трату."), nil
	})

	response, err := service.SendMessage(context.Background(), &SendMessageInput{
		Text:   "кофе 350",
		UserID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response != "Не удалось сохранить трату." || calls != 2 || repo.createCalls != 1 || len(repo.expenses) != 0 {
		t.Error("failed mutation was repeated or persisted")
	}
}

func TestSendMessageEditsExpenseAfterReadingCurrentRecords(t *testing.T) {
	owner := uuid.New()
	id := uuid.New()
	repo := &fakeExpenseRepository{
		expenses: map[uuid.UUID]domain.Expense{
			id: {
				ID:        id,
				UserID:    owner,
				Amount:    350,
				Category:  domain.CategoryFood,
				SpentAt:   time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
				Comment:   "Кофе",
				CreatedAt: time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	calls := 0
	service := newTestLLM(t, repo, func(
		_ context.Context,
		request *ai.ModelRequest,
		_ ai.ModelStreamCallback,
	) (*ai.ModelResponse, error) {
		calls++
		switch calls {
		case 1:
			return toolCall("list_expenses", map[string]any{
				"category":  "",
				"date_from": "",
				"date_to":   "",
			}), nil
		case 2:
			var output tools.ListExpensesOutput
			readToolOutput(t, request, &output)
			if !output.Success || len(output.Expenses) != 1 {
				t.Fatalf("unexpected lookup result: %+v", output)
			}
			found := output.Expenses[0]
			return toolCall("edit_expense", map[string]any{
				"expense_id": found.ID,
				"amount":     450,
				"category":   found.Category,
				"spent_at":   found.SpentAt.Format("2006-01-02"),
				"comment":    found.Comment,
			}), nil
		default:
			var output tools.EditExpenseOutput
			readToolOutput(t, request, &output)
			if !output.Success || output.Expense.Amount != 450 {
				t.Fatalf("unexpected edit result: %+v", output)
			}
			return modelText("Изменено: 350 ₽ → 450 ₽"), nil
		}
	})

	_, err := service.SendMessage(context.Background(), &SendMessageInput{
		Text:   "последнее было не 350, а 450",
		UserID: owner.String(),
		Messages: []domain.Message{
			{Role: domain.MessageApp, Message: "устаревшая запись 1000 ₽"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || repo.expenses[id].Amount != 450 || repo.expenses[id].Comment != "Кофе" {
		t.Fatal("edit did not use the current expense or changed unrelated fields")
	}
}

func TestSendMessageCanAskForMissingAmountWithoutCreatingExpense(t *testing.T) {
	repo := &fakeExpenseRepository{expenses: make(map[uuid.UUID]domain.Expense)}
	calls := 0
	service := newTestLLM(t, repo, func(
		_ context.Context,
		request *ai.ModelRequest,
		_ ai.ModelStreamCallback,
	) (*ai.ModelResponse, error) {
		calls++
		if request.ToolChoice != ai.ToolChoiceAuto {
			t.Error("a clarification must not require a database call")
		}
		return modelText("Сколько стоил кофе?"), nil
	})

	response, err := service.SendMessage(context.Background(), &SendMessageInput{
		Text:   "кофе",
		UserID: uuid.NewString(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response != "Сколько стоил кофе?" || calls != 1 || repo.createCalls != 0 || repo.listCalls != 0 {
		t.Fatal("missing information triggered a database call instead of a clarification")
	}
}

func TestGenerateRejectsInvalidIdentity(t *testing.T) {
	service := &LLMService{}
	for _, input := range []*SendMessageInput{
		nil,
		{Text: "кофе 350", UserID: "invalid"},
		{Text: "кофе 350", UserID: uuid.Nil.String()},
		{Text: "  ", UserID: uuid.NewString()},
	} {
		if _, err := service.generate(context.Background(), input); !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("generate(%+v) error = %v, want ErrInvalidInput", input, err)
		}
	}
}

func TestSendMessageReturnsToolTotal(t *testing.T) {
	owner, id := uuid.New(), uuid.New()
	foreignID := uuid.New()
	repo := &fakeExpenseRepository{expenses: map[uuid.UUID]domain.Expense{
		id:        {ID: id, UserID: owner, Amount: 350, Category: domain.CategoryFood},
		foreignID: {ID: foreignID, UserID: uuid.New(), Amount: 900, Category: domain.CategoryFood},
	}}
	calls := 0
	service := newTestLLM(t, repo, func(
		_ context.Context,
		request *ai.ModelRequest,
		_ ai.ModelStreamCallback,
	) (*ai.ModelResponse, error) {
		calls++
		if calls == 1 {
			return toolCall("total_expenses", map[string]any{
				"category": "", "date_from": "", "date_to": "",
			}), nil
		}

		var output tools.TotalExpensesOutput
		readToolOutput(t, request, &output)
		if !output.Success || output.Total != 350 {
			t.Fatalf("unexpected total: %+v", output)
		}
		return modelText(fmt.Sprintf("Всего потрачено: %d ₽", output.Total)), nil
	})
	response, err := service.SendMessage(context.Background(), &SendMessageInput{
		Text: "Сколько всего потратил?", UserID: owner.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response != "Всего потрачено: 350 ₽" || calls != 2 || repo.listCalls != 1 {
		t.Fatalf("response = %q, model calls = %d, list calls = %d", response, calls, repo.listCalls)
	}
}
