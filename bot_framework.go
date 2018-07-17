package tgbot

import (
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"runtime/debug"
)

// CommonHandler is a short type alias for handler function
type CommonHandler func(bot *BotFramework, update *tgbotapi.Update) error

// BotFramework main object to work with. Instantiate using NewBotFramework
type BotFramework struct {
	tgbotapi.BotAPI
	commands              map[string]map[int64]CommonHandler
	handlers              map[string]map[int64]CommonHandler
	callbackQueryHandlers map[string]map[int64]CommonHandler
	inlineQueryHandlers   map[string]map[int64]CommonHandler
}

// NewBotFramework creates new bot instance
func NewBotFramework(api *tgbotapi.BotAPI) *BotFramework {
	bot := BotFramework{
		*api,
		make(map[string]map[int64]CommonHandler),
		make(map[string]map[int64]CommonHandler),
		make(map[string]map[int64]CommonHandler),
		make(map[string]map[int64]CommonHandler),
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
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
					debug.PrintStack()
				}
			}()
			err := bot.HandleUpdate(&u)
			if err == nil {
				return
			}
			if bot.GetChatID(&u) > 0 {
				bot.Send(tgbotapi.NewMessage(
					bot.GetChatID(&u),
					err.Error(),
				))
			}
		}()
	}
}

// HandleUpdate handles single update from channel
func (bot *BotFramework) HandleUpdate(update *tgbotapi.Update) error {
	anyErr := bot.handle(update, "any")
	if anyErr == nil || anyErr.Error() != "no handlers" {
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
	if update.Message.Photo != nil {
		return bot.handle(update, "photo")
	}
	if update.Message.Document != nil {
		return bot.handle(update, "file")
	}
	if update.Message.Contact != nil {
		return bot.handle(update, "contact")
	}
	if update.Message.Sticker != nil {
		return bot.handle(update, "sticker")
	}
	if update.Message.Audio != nil {
		return bot.handle(update, "audio")
	}
	if update.Message.Video != nil {
		return bot.handle(update, "video")
	}
	if update.Message.VideoNote != nil {
		return bot.handle(update, "video_note")
	}
	if update.Message.Voice != nil {
		return bot.handle(update, "voice")
	}
	if update.Message.Location != nil {
		return bot.handle(update, "location")
	}
	if update.Message.Venue != nil {
		return bot.handle(update, "venue")
	}
	if update.Message.Text != "" {
		err := bot.handleCommand(update)
		return err
	}
	return nil
}

// RegisterCommand binds handler for commands
// If chatID=0, command will work in any chat
// Commands name can be any string
// For example:
// 	bot.RegisterCommand("/start", SomeStartHandler, -1001234567)
// binds same handler for "/start", "/start@your_bot", "/start@your_bot someReferralCode"
// 	bot.RegisterCommand("🔔 Subscribe", SomeSubscribeHandler, 0)
// binds handler for message text "🔔 Subscribe"
func (bot *BotFramework) RegisterCommand(name string, f CommonHandler, chatID int64) error {
	if f == nil {
		return errors.New("handler must not be nil")
	}
	if _, ok := bot.commands[name]; !ok {
		bot.commands[name] = make(map[int64]CommonHandler, 1)
	}
	bot.commands[name][chatID] = f
	return nil
}

// UnregisterCommand deletes handler for command name in given chat
func (bot *BotFramework) UnregisterCommand(name string, chatID int64) error {
	delete(bot.commands[name], chatID)
	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)

	if update.Message.IsCommand() {
		if commands, ok := bot.commands["/"+update.Message.Command()]; ok {
			if command, ok := commands[chatID]; ok {
				return command(bot, update)
			} else if command, ok = commands[0]; ok {
				return command(bot, update)
			}
		}
	}
	if commands, ok := bot.commands[update.Message.Text]; ok {
		if command, ok := commands[chatID]; ok {
			return command(bot, update)
		} else if command, ok = commands[0]; ok {
			return command(bot, update)
		}
	}
	return bot.handle(update, "plain")
}

// RegisterCallbackQueryHandler binds handler for callback data
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterCallbackQueryHandler(f CommonHandler, dataStartsWith string, chatID int64) error {
	if _, ok := bot.callbackQueryHandlers[dataStartsWith]; !ok {
		bot.callbackQueryHandlers[dataStartsWith] = make(map[int64]CommonHandler)
	}
	bot.callbackQueryHandlers[dataStartsWith][chatID] = f
	return nil
}

// UnregisterCallbackQueryHandler deletes handler for given chat
func (bot *BotFramework) UnregisterCallbackQueryHandler(dataStartsWith string, chatID int64) error {
	delete(bot.callbackQueryHandlers[dataStartsWith], chatID)
	return nil
}

func (bot *BotFramework) handleCallbackQuery(update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)
	data := update.CallbackQuery.Data

	for key := range bot.callbackQueryHandlers {
		if len(key) > len(data) {
			continue
		}
		if data[:len(key)] == key {
			if command, ok := bot.callbackQueryHandlers[key][chatID]; ok {
				return command(bot, update)
			} else if command, ok = bot.callbackQueryHandlers[key][0]; ok {
				return command(bot, update)
			}
		}
	}

	return errors.New("unknown handler")
}

// RegisterInlineQueryHandler binds handler for query
// If userID = 0, command will work for any user
func (bot *BotFramework) RegisterInlineQueryHandler(f CommonHandler, query string, userID int64) error {
	if _, ok := bot.inlineQueryHandlers[query]; !ok {
		bot.inlineQueryHandlers[query] = make(map[int64]CommonHandler)
	}
	bot.inlineQueryHandlers[query][userID] = f
	return nil
}

// UnregisterInlineQueryHandler deletes handler for given user
func (bot *BotFramework) UnregisterInlineQueryHandler(query string, userID int64) error {
	delete(bot.inlineQueryHandlers[query], userID)
	return nil
}

func (bot *BotFramework) handleInlineQuery(update *tgbotapi.Update) error {
	userID := int64(update.InlineQuery.From.ID)
	query := update.InlineQuery.Query

	for key := range bot.inlineQueryHandlers {
		if len(query) > len(key) {
			continue
		}
		if key[:len(query)] == query {
			if command, ok := bot.inlineQueryHandlers[key][userID]; ok {
				return command(bot, update)
			} else if command, ok = bot.inlineQueryHandlers[key][0]; ok {
				return command(bot, update)
			}
		}
	}

	return errors.New("unknown handler")
}

func (bot *BotFramework) handle(update *tgbotapi.Update, event string) error {
	chatID := bot.GetChatID(update)
	if command, ok := bot.handlers[event][chatID]; ok {
		return command(bot, update)
	} else if command, ok = bot.handlers[event][0]; ok {
		return command(bot, update)
	}
	return errors.New("no handlers")
}

// RegisterPlainTextHandler binds handler for plain text message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterPlainTextHandler(f CommonHandler, chatID int64) error {
	bot.handlers["plain"][chatID] = f
	return nil
}

// UnregisterPlainTextHandler deletes handler for given chat
func (bot *BotFramework) UnregisterPlainTextHandler(chatID int64) error {
	delete(bot.handlers["plain"], chatID)
	return nil
}

// RegisterContactHandler binds handler for contact message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterContactHandler(f CommonHandler, chatID int64) error {
	bot.handlers["contact"][chatID] = f
	return nil
}

// UnregisterContactHandler deletes handler for given chat
func (bot *BotFramework) UnregisterContactHandler(chatID int64) error {
	delete(bot.handlers["contact"], chatID)
	return nil
}

// RegisterPhotoHandler binds handler for photo message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterPhotoHandler(f CommonHandler, chatID int64) error {
	bot.handlers["photo"][chatID] = f
	return nil
}

// UnregisterPhotoHandler deletes handler for given chat
func (bot *BotFramework) UnregisterPhotoHandler(chatID int64) error {
	delete(bot.handlers["photo"], chatID)
	return nil
}

// RegisterFileHandler binds handler for file from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterFileHandler(f CommonHandler, chatID int64) error {
	bot.handlers["file"][chatID] = f
	return nil
}

// UnregisterFileHandler deletes handler for given chat
func (bot *BotFramework) UnregisterFileHandler(chatID int64) error {
	delete(bot.handlers["file"], chatID)
	return nil
}

// RegisterStickerHandler binds handler for sticker from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterStickerHandler(f CommonHandler, chatID int64) error {
	bot.handlers["sticker"][chatID] = f
	return nil
}

// UnregisterStickerHandler deletes handler for given chat
func (bot *BotFramework) UnregisterStickerHandler(chatID int64) error {
	delete(bot.handlers["sticker"], chatID)
	return nil
}

// RegisterAudioHandler binds handler for audio message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterAudioHandler(f CommonHandler, chatID int64) error {
	bot.handlers["audio"][chatID] = f
	return nil
}

// UnregisterAudioHandler deletes handler for given chat
func (bot *BotFramework) UnregisterAudioHandler(chatID int64) error {
	delete(bot.handlers["audio"], chatID)
	return nil
}

// RegisterVideoHandler binds handler for video message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVideoHandler(f CommonHandler, chatID int64) error {
	bot.handlers["video"][chatID] = f
	return nil
}

// UnregisterVideoHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVideoHandler(chatID int64) error {
	delete(bot.handlers["video"], chatID)
	return nil
}

// RegisterVideoNoteHandler binds handler for video_note message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVideoNoteHandler(f CommonHandler, chatID int64) error {
	bot.handlers["video_note"][chatID] = f
	return nil
}

// UnregisterVideoNoteHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVideoNoteHandler(chatID int64) error {
	delete(bot.handlers["video_note"], chatID)
	return nil
}

// RegisterVoiceHandler binds handler for voice message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVoiceHandler(f CommonHandler, chatID int64) error {
	bot.handlers["voice"][chatID] = f
	return nil
}

// UnregisterVoiceHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVoiceHandler(chatID int64) error {
	delete(bot.handlers["voice"], chatID)
	return nil
}

// RegisterVenueHandler binds handler for venue message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVenueHandler(f CommonHandler, chatID int64) error {
	bot.handlers["venue"][chatID] = f
	return nil
}

// UnregisterVenueHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVenueHandler(chatID int64) error {
	delete(bot.handlers["venue"], chatID)
	return nil
}

// RegisterLocationHandler binds handler for location message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterLocationHandler(f CommonHandler, chatID int64) error {
	bot.handlers["location"][chatID] = f
	return nil
}

// UnregisterLocationHandler deletes handler for given chat
func (bot *BotFramework) UnregisterLocationHandler(chatID int64) error {
	delete(bot.handlers["location"], chatID)
	return nil
}

// RegisterUniversalHandler binds handler for any message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterUniversalHandler(f CommonHandler, chatID int64) error {
	bot.handlers["any"][chatID] = f
	return nil
}

// UnregisterUniversalHandler deletes handler for given chat
func (bot *BotFramework) UnregisterUniversalHandler(chatID int64) error {
	delete(bot.handlers["any"], chatID)
	return nil
}
