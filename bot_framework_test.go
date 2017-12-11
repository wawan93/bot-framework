package bot_framework

import (
	"testing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"errors"
)

type testSender struct {
	Sender
	messages chan bool
}

func (s *testSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.messages <- true
	return tgbotapi.Message{}, nil
}

func newMock() *testSender {
	mock := new(testSender)
	mock.messages = make(chan bool, 1)
	return mock
}

func TestTestSendable(t *testing.T) {
	mock := newMock()
	mock.Send(
		&tgbotapi.MessageConfig{},
	)

	if ! <-mock.messages {
		t.Error("Message not sent")
	}
}

func TestNewBotFramework(t *testing.T) {
	t.Parallel()

	bot := NewBotFramework(newMock())

	if bot == nil {
		t.Error("Not created bot")
	}
}

func TestCommands(t *testing.T) {
	t.Parallel()

	bot := NewBotFramework(newMock())
	bot.RegisterCommand(&Command{
		Name: "/test",
		Handler: func(bot Sender, update *tgbotapi.Update) error {
			return nil
		},
	})

	t.Run("register", func(t *testing.T) {
		t.Parallel()

		err := bot.RegisterCommand(&Command{
			Name: "asdf",
		})
		if err == nil || err.Error() != "command must start with slash" {
			t.Error("command without slash must return error")
		}

		if len(bot.commands) != 1 {
			t.Error("Command not registered")
		}
	})

	t.Run("handle command", func(t *testing.T) {
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

	t.Run("not handle command", func(t *testing.T) {
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

func TestBotFramework_HandleUpdates(t *testing.T) {
	t.Parallel()

	mock := newMock()
	bot := NewBotFramework(mock)
	bot.RegisterCommand(&Command{
		Name: "/test",
		Handler: func(bot Sender, update *tgbotapi.Update) error {
			bot.Send(&tgbotapi.MessageConfig{})
			return nil
		},
	})

	channel := make(chan tgbotapi.Update)

	command := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Entities: &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}},
			Text:     "/test",
		},
	}

	t.Run("handle nil message", func(t *testing.T) {
		t.Parallel()
		err := bot.handleUpdate(&tgbotapi.Update{})
		if err == nil {
			t.Error("empty update must return error")
		}
	})

	t.Run("not handled message", func(t *testing.T) {
		t.Parallel()
		err := bot.handleUpdate(&tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "asdf",
			},
		})
		if err == nil {
			t.Error("empty update must return error")
		}
	})

	t.Run("handle commands", func(t *testing.T) {
		t.Parallel()

		err := bot.handleUpdate(&command)
		if err != nil {
			t.Error(err)
		}

		if ! <-mock.messages {
			t.Error("Message not sent")
		}
	})

	t.Run("handle updates channel", func(t *testing.T) {
		t.Parallel()

		go bot.HandleUpdates(channel)
		channel <- command

		if ! <-mock.messages {
			t.Error("Message not sent")
		}

		channel <- tgbotapi.Update{
			Message: &tgbotapi.Message{
				Entities: &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}},
				Text:     "/test2",
				Chat:     &tgbotapi.Chat{ID: 1},
			},
		}

		if ! <-mock.messages {
			t.Error("message not sent")
		}
	})
}

func TestBotFramework_RegisterKeyboardCommand(t *testing.T) {
	t.Parallel()

	mock := newMock()
	bot := NewBotFramework(mock)

	bot.RegisterKeyboardCommand(&Command{
		Name: "👍 test",
		Handler: func(bot Sender, update *tgbotapi.Update) error {
			bot.Send(&tgbotapi.MessageConfig{})
			return nil
		},
	})

	ch := make(chan tgbotapi.Update)

	t.Run("register", func(t *testing.T) {
		t.Parallel()
		err := bot.RegisterKeyboardCommand(&Command{
			Name: "/asdf",
		})
		if err == nil || err.Error() != "keyboard command must not start with slash" {
			t.Error("keyboard command with slash must return error")
		}

		if len(bot.commands) != 1 {
			t.Error("Command not registered")
		}
	})

	t.Run("handle commands", func(t *testing.T) {
		t.Parallel()
		err := bot.handleKeyboardCommand(&tgbotapi.Update{})
		if err == nil || err.Error() != "no message" {
			t.Error("handle must return error")
		}
	})

	t.Run("handle keyboard command updates", func(t *testing.T) {
		t.Parallel()

		go bot.HandleUpdates(ch)

		ch <- tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "👍 test",
			},
		}

		if ! <-mock.messages {
			t.Error("Message not sent")
		}

		bot.RegisterKeyboardCommand(&Command{
			Name: "👎 test",
			Handler: func(bot Sender, update *tgbotapi.Update) error {
				return errors.New("test some error")
			},
		})

		ch <- tgbotapi.Update{
			Message: &tgbotapi.Message{
				Text: "👎 test",
				Chat: &tgbotapi.Chat{ID: 1},
			},
		}

		if ! <-mock.messages {
			t.Error("Message not sent")
		}
	})
}
