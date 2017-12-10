package bot_framework

import (
	"testing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type testSendable struct {
	Sendable
}

func (s testSendable) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return tgbotapi.Message{}, nil
}

func TestNewBotFramework(t *testing.T) {
	bot := NewBotFramework(new(testSendable))

	if bot == nil {
		t.Error("Not created bot")
	}
}



func TestBotFramework_RegisterCommand(t *testing.T) {
	bot := NewBotFramework(new(testSendable))
	bot.RegisterCommand(&Command{
		Name: "test",
		Handler: func(bot Sendable, update *tgbotapi.Update) error {
			return nil
		},
	})

	if len(bot.commands) != 1 {
		t.Error("Command not registered")
	}
}

func TestBotFramework_HandleCommand(t *testing.T) {
	bot := NewBotFramework(new(testSendable))
	bot.RegisterCommand(&Command{
		Name: "test",
		Handler: func(bot Sendable, update *tgbotapi.Update) error {
			return nil
		},
	})

	if len(bot.commands) != 1 {
		t.Error("Command not registered")
	}

	err := bot.HandleCommand(&tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}},
			Text:     "/test",
		},
	})

	if err != nil {
		t.Error("command not handled")
	}

	err = bot.HandleCommand(&tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "test",
		},
	})

	if err == nil {
		t.Error("command handled, but must not")
	}
}
