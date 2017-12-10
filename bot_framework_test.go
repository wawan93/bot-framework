package bot_framework

import (
	"testing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type testSendable struct {
	Sendable
	MessageSent bool
}

func (s testSendable) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.MessageSent = true
	return tgbotapi.Message{}, nil
}

func TestNewBotFramework(t *testing.T) {
	t.Parallel()

	bot := NewBotFramework(new(testSendable))

	if bot == nil {
		t.Error("Not created bot")
	}
}

func TestCommands(t *testing.T) {
	t.Parallel()

	bot := NewBotFramework(new(testSendable))
	bot.RegisterCommand(&Command{
		Name: "test",
		Handler: func(bot Sendable, update *tgbotapi.Update) error {
			return nil
		},
	})

	t.Run("Register", func (t *testing.T) {
		t.Parallel()

		if len(bot.commands) != 1 {
			t.Error("Command not registered")
		}
	})

	t.Run("Handle command", func(t *testing.T) {
		t.Parallel()

		err := bot.handleCommand(&tgbotapi.Update{
			Message: &tgbotapi.Message{
				Entities: &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}},
				Text:     "/test",
			},
		})

		if err != nil {
			t.Error("command not handled")
		}
	})

	t.Run("Not handle command", func(t *testing.T) {
		err := bot.handleCommand(&tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "test",
			},
		})

		if err == nil {
			t.Error("command handled, but must not")
		}
	})
}
