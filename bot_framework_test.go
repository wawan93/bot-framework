package tgbot

import (
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"
)

type rewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, `{"ok":true}`)
}

func getBot() BotFramework {
	server := httptest.NewServer(http.HandlerFunc(okHandler))
	sURL, err := url.Parse(server.URL)
	client := server.Client()
	client.Transport = rewriteTransport{URL: sURL}
	api, err := tgbotapi.NewBotAPIWithClient(
		"token",
		client,
	)
	if err != nil {
		log.Panic(err)
	}
	return *NewBotFramework(api)
}

func TestBotFramework_GetChatID(t *testing.T) {
	t.Parallel()
	bot := getBot()
	cases := []struct {
		data     *tgbotapi.Update
		expected int64
	}{
		{
			data: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 123},
				},
			},
			expected: 123,
		},
		{
			data: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 345},
					},
				},
			},
			expected: 345,
		},
		{
			data:     &tgbotapi.Update{},
			expected: 0,
		},
	}

	for _, test := range cases {
		if bot.GetChatID(test.data) != test.expected {
			t.Error("chat ID doesn't match")
		}
	}
}

func TestBotFramework_handleUpdate(t *testing.T) {
	t.Parallel()
	bot := getBot()

	u := &tgbotapi.Update{}
	err := bot.HandleUpdate(u)
	if err == nil || err.Error() != "no message" {
		t.Error("empty update must not be handled")
	}
}

func TestBotFramework_CallbackQueryHandlers(t *testing.T) {
	t.Parallel()
	bot := getBot()

	err := bot.RegisterCallbackQueryHandler(
		func(bot *BotFramework, update *tgbotapi.Update) error {
			return errors.New("test")
		},
		"asdf_",
		0,
	)
	if err != nil {
		t.Error(err)
	}

	err = bot.RegisterCallbackQueryHandler(
		func(bot *BotFramework, update *tgbotapi.Update) error {
			return errors.New("test 2")
		},
		"asdf_",
		123,
	)
	if err != nil {
		t.Error(err)
	}

	cases := []struct {
		data     *tgbotapi.Update
		expected string
	}{
		{
			data: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "asdf_123",
				},
			},
			expected: "test",
		},
		{
			data: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "asdfqwerty_123",
				},
			},
			expected: "unknown handler",
		},
		{
			data: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "a",
				},
			},
			expected: "unknown handler",
		},
		{
			data: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "asdf_123",
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: 123},
					},
				},
			},
			expected: "test 2",
		},
		{
			data: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: "asdf_123",
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{
							ID: 12345,
						},
					},
				},
			},
			expected: "test",
		},
	}

	for _, testCase := range cases {
		if err = bot.HandleUpdate(testCase.data); err == nil {
			t.Error("handler must return given error")
		} else if err.Error() != testCase.expected {
			t.Error(err)
		}
	}
}

func TestBotFramework_UnregisterCallbackQueryHandler(t *testing.T) {
	t.Parallel()
	bot := getBot()
	bot.RegisterCallbackQueryHandler(
		func(bot *BotFramework, update *tgbotapi.Update) error {
			return errors.New("test passed")
		},
		"asdf_",
		0,
	)

	u := &tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Data: "asdf_123",
		},
	}
	err := bot.HandleUpdate(u)
	if err == nil {
		t.Error("handler is not set")
	} else if err.Error() != "test passed" {
		t.Error(err)
	}

	bot.UnregisterCallbackQueryHandler("asdf_", 0)
	err = bot.HandleUpdate(u)
	if err == nil {
		t.Error("handler must not be set")
	} else if err.Error() != "unknown handler" {
		t.Error(err)
	}
}

func TestBotFramework_HandleUpdates(t *testing.T) {
	t.Parallel()
	bot := getBot()

	chat := &tgbotapi.Chat{ID: 123}
	bot.RegisterCommand("test 1", func(bot *BotFramework, update *tgbotapi.Update) error {
		panic("test passed")
	}, chat.ID)

	bot.RegisterCommand("test 2", func(bot *BotFramework, update *tgbotapi.Update) error {
		return errors.New("test passed")
	}, chat.ID)

	bot.RegisterCommand("test 3", func(bot *BotFramework, update *tgbotapi.Update) error {
		return nil
	}, chat.ID)

	uc := make(chan tgbotapi.Update, 3)
	go bot.HandleUpdates(uc)

	uc <- tgbotapi.Update{Message: &tgbotapi.Message{
		Chat: chat, Text: "test 2",
	}}
	uc <- tgbotapi.Update{Message: &tgbotapi.Message{
		Chat: chat, Text: "test 3",
	}}
	uc <- tgbotapi.Update{Message: &tgbotapi.Message{
		Chat: chat, Text: "test 1",
	}}
}

func TestBotFramework_PlainTextHandler(t *testing.T) {
	t.Parallel()

	bot := getBot()
	chat := &tgbotapi.Chat{ID: 123}

	var reallySent bool

	u := &tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "hello, world!",
			Chat: chat,
		},
	}
	bot.HandleUpdate(u)
	if reallySent == true {
		t.Error("plain text handler not registered, but no error retuned")
	}

	bot.RegisterPlainTextHandler(
		func(bot *BotFramework, update *tgbotapi.Update) error {
			reallySent = true
			return nil
		},
		chat.ID,
	)
	bot.HandleUpdate(u)
	if reallySent != true {
		t.Error("message must be sent")
	}
	reallySent = false

	u.Message.Chat.ID = 999

	bot.HandleUpdate(u)
	if reallySent != false {
		t.Error("message must not be sent to wrong chat")
	}

	bot.UnregisterPlainTextHandler(chat.ID)
	bot.HandleUpdate(u)
	if reallySent != false {
		t.Error("message must be sent")
	}
}

func TestBotFramework_PhotoHandler(t *testing.T) {
	t.Parallel()

	bot := getBot()
	chat := &tgbotapi.Chat{ID: 123}

	var reallySent bool

	u := &tgbotapi.Update{
		Message: &tgbotapi.Message{
			Photo: &[]tgbotapi.PhotoSize{
				{},
			},
			Chat: chat,
		},
	}
	bot.HandleUpdate(u)
	if reallySent == true {
		t.Error("photo handler not registered, but no error retuned")
	}

	bot.RegisterPhotoHandler(func(bot *BotFramework, update *tgbotapi.Update) error {
		reallySent = true
		return nil
	}, chat.ID)

	bot.HandleUpdate(u)
	if reallySent != true {
		t.Error("message must be sent")
	}
	reallySent = false

	u.Message.Chat.ID = 999

	bot.HandleUpdate(u)
	if reallySent != false {
		t.Error("message must not be sent to wrong chat")
	}

	bot.UnregisterPhotoHandler(chat.ID)
	bot.HandleUpdate(u)
	if reallySent != false {
		t.Error("message must be sent")
	}
}

func TestHandleCommand(t *testing.T) {
	bot := getBot()
	bot.RegisterCommand("/start", func(bot *BotFramework, update *tgbotapi.Update) error {
		if !update.Message.IsCommand() {
			t.Errorf("not command %s", update.Message.Text)
		}

		if update.Message.Command() != "start" {
			t.Errorf("expected command \"start\", got \"%s\"", update.Message.Command())
		}
		return nil
	}, 0)

	u := &tgbotapi.Update{Message: &tgbotapi.Message{
		Chat:     &tgbotapi.Chat{ID: 123},
		Text:     "/start@wawan_pro_bot helloworld",
		Entities: &[]tgbotapi.MessageEntity{{Offset: 0, Length: 6, Type: "bot_command"}},
	}}
	bot.HandleUpdate(u)
}
