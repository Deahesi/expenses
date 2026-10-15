package user

// CreateUserInput contains data for registering a new family member.
// TelegramID is optional: 0 means the user is not linked to Telegram yet.
type CreateUserInput struct {
	TelegramID int64
	FirstName  string
	LastName   string
}

// UpdateUserInput contains the editable user fields.
type UpdateUserInput struct {
	FirstName string
	LastName  string
}

type UpsertTelegramUserInput struct {
	TelegramID int64
	FirstName  string
	LastName   string
}
