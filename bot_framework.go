package tgbot

import (
	"errors"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var NoHandlersError = errors.New("no handlers for update")

// CommonHandler is a short type alias for handler function
type CommonHandler func(bot *BotFramework, update *tgbotapi.Update) error

// BotFramework main object to work with. Instantiate using NewBotFramework
type BotFramework struct {
	tgbotapi.BotAPI
	commands              map[string]map[int64]CommonHandler
	handlers              map[string]map[int64]CommonHandler
	callbackQueryHandlers map[string]map[int64]CommonHandler
	inlineQueryHandlers   map[string]map[int64]CommonHandler
	mu                    sync.RWMutex
	ErrorHandler          func(u tgbotapi.Update, err error)
}

// NewBotFramework creates new bot instance
func NewBotFramework(api *tgbotapi.BotAPI) *BotFramework {
	bot := BotFramework{
		BotAPI:                *api,
		commands:              make(map[string]map[int64]CommonHandler),
		handlers:              make(map[string]map[int64]CommonHandler),
		callbackQueryHandlers: make(map[string]map[int64]CommonHandler),
		inlineQueryHandlers:   make(map[string]map[int64]CommonHandler),
	}
	bot.handlers["plain"] = make(map[int64]CommonHandler)
	bot.handlers["photo"] = make(map[int64]CommonHandler)
	bot.handlers["file"] = make(map[int64]CommonHandler)
	bot.handlers["contact"] = make(map[int64]CommonHandler)
	bot.handlers["sticker"] = make(map[int64]CommonHandler)
	bot.handlers["audio"] = make(map[int64]CommonHandler)
	bot.handlers["video"] = make(map[int64]CommonHandler)
	bot.handlers["video_note"] = make(map[int64]CommonHandler)
	bot.handlers["voice"] = make(map[int64]CommonHandler)
	bot.handlers["location"] = make(map[int64]CommonHandler)
	bot.handlers["venue"] = make(map[int64]CommonHandler)
	bot.handlers["any"] = make(map[int64]CommonHandler)
	bot.ErrorHandler = func(u tgbotapi.Update, err error) {
		if bot.GetChatID(&u) > 0 {
			_, _ = bot.Send(tgbotapi.NewMessage(
				bot.GetChatID(&u),
				err.Error(),
			))
		}
	}
	return &bot
}

// GetChatID returns chat ID from update message and callback query
func (bot *BotFramework) GetChatID(update *tgbotapi.Update) int64 {
	if update.Message != nil {
		if update.Message.Chat != nil {
			return update.Message.Chat.ID
		}
	}

	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message != nil {
			return update.CallbackQuery.Message.Chat.ID
		}
	}

	return 0
}

// HandleUpdates handles all updates from channel.
// save for panics
func (bot *BotFramework) HandleUpdates(ch tgbotapi.UpdatesChannel) {
	for update := range ch {
		u := update
		go func() {
			err := bot.HandleUpdate(&u)
			if err != nil {
				bot.ErrorHandler(u, err)
			}
		}()
	}
}

// HandleUpdate handles single update from channel
func (bot *BotFramework) HandleUpdate(update *tgbotapi.Update) error {
	anyErr := bot.handle(update, "any")
	if anyErr == nil || !errors.Is(anyErr, NoHandlersError) {
		return anyErr
	}

	if update.CallbackQuery != nil {
		return bot.handleCallbackQuery(update)
	}
	if update.InlineQuery != nil {
		return bot.handleInlineQuery(update)
	}
	if update.Message == nil {
		return errors.New("no message")
	}

	switch {
	case update.Message.Photo != nil:
		return bot.handle(update, "photo")
	case update.Message.Document != nil:
		return bot.handle(update, "file")
	case update.Message.Contact != nil:
		return bot.handle(update, "contact")
	case update.Message.Sticker != nil:
		return bot.handle(update, "sticker")
	case update.Message.Audio != nil:
		return bot.handle(update, "audio")
	case update.Message.Video != nil:
		return bot.handle(update, "video")
	case update.Message.VideoNote != nil:
		return bot.handle(update, "video_note")
	case update.Message.Voice != nil:
		return bot.handle(update, "voice")
	case update.Message.Location != nil:
		return bot.handle(update, "location")
	case update.Message.Venue != nil:
		return bot.handle(update, "venue")
	case update.Message.Text != "":
		return bot.handleCommand(update)
	}

	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)

	key := update.Message.Text
	if update.Message.IsCommand() {
		key = "/" + update.Message.Command()
	}

	bot.mu.RLock()

	if commands, ok := bot.commands[key]; ok {
		if command, ok := commands[chatID]; ok {
			bot.mu.RUnlock()
			return command(bot, update)
		} else if command, ok = commands[0]; ok {
			bot.mu.RUnlock()
			return command(bot, update)
		}
	}

	bot.mu.RUnlock()
	return bot.handle(update, "plain")
}

func (bot *BotFramework) handleCallbackQuery(update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)
	data := update.CallbackQuery.Data

	bot.mu.RLock()

	for key := range bot.callbackQueryHandlers {
		if len(key) > len(data) {
			continue
		}
		if data[:len(key)] == key {
			if command, ok := bot.callbackQueryHandlers[key][chatID]; ok {
				bot.mu.RUnlock()
				return command(bot, update)
			} else if command, ok = bot.callbackQueryHandlers[key][0]; ok {
				bot.mu.RUnlock()
				return command(bot, update)
			}
		}
	}

	bot.mu.RUnlock()
	return fmt.Errorf("%w: callback, chatID=%d, data=%s", NoHandlersError, chatID, data)
}

func (bot *BotFramework) handleInlineQuery(update *tgbotapi.Update) error {
	userID := int64(update.InlineQuery.From.ID)
	query := update.InlineQuery.Query

	bot.mu.RLock()

	for key := range bot.inlineQueryHandlers {
		if len(query) > len(key) {
			continue
		}
		if key[:len(query)] == query {
			if command, ok := bot.inlineQueryHandlers[key][userID]; ok {
				bot.mu.RUnlock()
				return command(bot, update)
			} else if command, ok = bot.inlineQueryHandlers[key][0]; ok {
				bot.mu.RUnlock()
				return command(bot, update)
			}
		}
	}

	bot.mu.RUnlock()
	return fmt.Errorf("%w: inline, userID=%d, query=%s", NoHandlersError, userID, query)
}

func (bot *BotFramework) handle(update *tgbotapi.Update, event string) error {
	chatID := bot.GetChatID(update)

	bot.mu.RLock()

	if command, ok := bot.handlers[event][chatID]; ok {
		bot.mu.RUnlock()
		return command(bot, update)
	} else if command, ok = bot.handlers[event][0]; ok {
		bot.mu.RUnlock()
		return command(bot, update)
	}

	bot.mu.RUnlock()
	return fmt.Errorf("%w: chatID=%d, event=%s", NoHandlersError, chatID, event)
}
