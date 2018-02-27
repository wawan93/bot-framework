package bot_framework

import (
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type commonHandler func(bot *BotFramework, update *tgbotapi.Update) error

type BotFramework struct {
	tgbotapi.BotAPI
	commands              map[string]map[int64]commonHandler
	handlers              map[string]map[int64]commonHandler
	callbackQueryHandlers map[string]map[int64]commonHandler
}

func NewBotFramework(api *tgbotapi.BotAPI) *BotFramework {
	bot := BotFramework{
		*api,
		make(map[string]map[int64]commonHandler),
		make(map[string]map[int64]commonHandler),
		make(map[string]map[int64]commonHandler),
	}
	bot.handlers["plain"] = make(map[int64]commonHandler)
	bot.handlers["photo"] = make(map[int64]commonHandler)
	bot.handlers["file"] = make(map[int64]commonHandler)
	bot.handlers["contact"] = make(map[int64]commonHandler)
	return &bot
}

func (bot *BotFramework) HandleUpdates(ch tgbotapi.UpdatesChannel) {
	for update := range ch {
		u := update
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println(r)
				}
			}()
			err := bot.HandleUpdate(&u)
			if err == nil {
				return
			}
			if bot.GetChatID(&u) != 0 {
				bot.Send(tgbotapi.NewMessage(
					bot.GetChatID(&u),
					err.Error(),
				))
			}
		}()
	}
}

func (bot *BotFramework) HandleUpdate(update *tgbotapi.Update) error {
	if update.CallbackQuery != nil {
		return bot.handleCallbackQuery(update)
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
	if update.Message.Text != "" {
		err := bot.handleCommand(update)
		if err == nil {
			return nil
		}
		if err.Error() != "command not found" {
			return err
		}
		err = bot.handle(update, "plain")
		if err == nil {
			return nil
		}
	}
	return nil
}

func (bot *BotFramework) RegisterCommand(name string, f commonHandler, chatID int64) error {
	if f == nil {
		return errors.New("handler must not be nil")
	}
	if _, ok := bot.commands[name]; !ok {
		bot.commands[name] = make(map[int64]commonHandler, 1)
	}
	bot.commands[name][chatID] = f
	return nil
}

func (bot *BotFramework) UnregisterCommand(name string, chatID int64) error {
	delete(bot.commands[name], chatID)
	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)

	if update.Message.IsCommand() {
		if commands, ok := bot.commands[update.Message.Command()]; ok {
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
	return errors.New("command not found")
}

func (bot *BotFramework) RegisterPlainTextHandler(f commonHandler, chatID int64) error {
	bot.handlers["plain"][chatID] = f
	return nil
}

func (bot *BotFramework) UnregisterPlainTextHandler(chatID int64) error {
	delete(bot.handlers["plain"], chatID)
	return nil
}

func (bot *BotFramework) RegisterContactHandler(f commonHandler, chatID int64) error {
	bot.handlers["contact"][chatID] = f
	return nil
}

func (bot *BotFramework) UnregisterContactHandler(chatID int64) error {
	delete(bot.handlers["contact"], chatID)
	return nil
}

func (bot *BotFramework) RegisterPhotoHandler(f commonHandler, chatID int64) error {
	bot.handlers["photo"][chatID] = f
	return nil
}

func (bot *BotFramework) UnregisterPhotoHandler(chatID int64) error {
	delete(bot.handlers["photo"], chatID)
	return nil
}

func (bot *BotFramework) RegisterFileHandler(f commonHandler, chatID int64) error {
	bot.handlers["file"][chatID] = f
	return nil
}

func (bot *BotFramework) UnregisterFileHandler(chatID int64) error {
	delete(bot.handlers["file"], chatID)
	return nil
}

func (bot *BotFramework) RegisterCallbackQueryHandler(f commonHandler, dataStartsWith string, chatID int64) error {
	if _, ok := bot.callbackQueryHandlers[dataStartsWith]; !ok {
		bot.callbackQueryHandlers[dataStartsWith] = make(map[int64]commonHandler)
	}
	bot.callbackQueryHandlers[dataStartsWith][chatID] = f
	return nil
}

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

func (bot *BotFramework) handle(update *tgbotapi.Update, event string) error {
	chatID := bot.GetChatID(update)
	if command, ok := bot.handlers[event][chatID]; ok {
		return command(bot, update)
	} else if command, ok = bot.handlers[event][0]; ok {
		return command(bot, update)
	}
	return errors.New("unknown handler")
}

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
