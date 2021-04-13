package tgbot

import (
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/wawan93/bot-framework/storage"
)

type Storage interface {
	Set(kind string, name string, chatID int64, handler storage.Command)
	Get(kind string, name string, chatID int64) (storage.Command, error)
	Unset(kind string, name string, chatID int64)
}

type Kind string

const (
	CommandKind       Kind = "command"
	PlainTextKind     Kind = "plain"
	ContactKind       Kind = "contact"
	PhotoKind         Kind = "photo"
	FileKind          Kind = "file"
	StickerKind       Kind = "sticker"
	AudioKind         Kind = "audio"
	VideoKind         Kind = "video"
	VideoNoteKind     Kind = "video_note"
	VoiceKind         Kind = "voice"
	VenueKind         Kind = "venue"
	LocationKind      Kind = "location"
	InlineQueryKind   Kind = "inline_query"
	CallbackQueryKind Kind = "callback_query"
	AnyKind           Kind = "any"
)

// BotFramework main object to work with. Instantiate using NewBotFramework
type BotFramework struct {
	tgbotapi.BotAPI
	storage      Storage
	ErrorHandler func(u tgbotapi.Update, err error)
}

// NewBotFramework creates new bot instance
func NewBotFramework(api *tgbotapi.BotAPI, storage Storage) *BotFramework {
	bot := BotFramework{
		BotAPI:  *api,
		storage: storage,
	}
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
			if err == nil {
				return
			}
			bot.ErrorHandler(u, err)
		}()
	}
}

// HandleUpdate handles single update from channel
func (bot *BotFramework) HandleUpdate(update *tgbotapi.Update) error {
	anyErr := bot.handle(update, AnyKind)
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
		return bot.handle(update, PhotoKind)
	}
	if update.Message.Document != nil {
		return bot.handle(update, FileKind)
	}
	if update.Message.Contact != nil {
		return bot.handle(update, ContactKind)
	}
	if update.Message.Sticker != nil {
		return bot.handle(update, StickerKind)
	}
	if update.Message.Audio != nil {
		return bot.handle(update, AudioKind)
	}
	if update.Message.Video != nil {
		return bot.handle(update, VideoKind)
	}
	if update.Message.VideoNote != nil {
		return bot.handle(update, VideoNoteKind)
	}
	if update.Message.Voice != nil {
		return bot.handle(update, VoiceKind)
	}
	if update.Message.Location != nil {
		return bot.handle(update, LocationKind)
	}
	if update.Message.Venue != nil {
		return bot.handle(update, VenueKind)
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
func (bot *BotFramework) RegisterCommand(name string, f storage.Command, chatID int64) error {
	if f == nil {
		return errors.New("handler must not be nil")
	}

	bot.storage.Set(string(CommandKind), name, chatID, f)
	return nil
}

// UnregisterCommand deletes handler for command name in given chat
func (bot *BotFramework) UnregisterCommand(name string, chatID int64) error {
	bot.storage.Unset(string(CommandKind), name, chatID)
	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)

	if update.Message.IsCommand() {
		cmd, err := bot.storage.Get(string(CommandKind), "/"+update.Message.Command(), chatID)
		if err == nil {
			return cmd.Exec(update)
		}
	}

	cmd, err := bot.storage.Get(string(CommandKind), update.Message.Text, chatID)
	if err == nil {
		return cmd.Exec(update)
	}

	return bot.handle(update, PlainTextKind)
}

// RegisterCallbackQueryHandler binds handler for callback data
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterCallbackQueryHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(CallbackQueryKind), "", chatID, f)
	return nil
}

// UnregisterCallbackQueryHandler deletes handler for given chat
func (bot *BotFramework) UnregisterCallbackQueryHandler(chatID int64) error {
	bot.storage.Unset(string(CallbackQueryKind), "", chatID)
	return nil
}

func (bot *BotFramework) handleCallbackQuery(update *tgbotapi.Update) error {
	return bot.handle(update, CallbackQueryKind)
}

// RegisterInlineQueryHandler binds handler for query
// If userID = 0, command will work for any user
func (bot *BotFramework) RegisterInlineQueryHandler(f storage.Command, userID int64) error {
	bot.storage.Set(string(InlineQueryKind), "", userID, f)
	return nil
}

// UnregisterInlineQueryHandler deletes handler for given user
func (bot *BotFramework) UnregisterInlineQueryHandler(userID int64) error {
	bot.storage.Unset(string(InlineQueryKind), "", userID)
	return nil
}

func (bot *BotFramework) handleInlineQuery(update *tgbotapi.Update) error {
	return bot.handle(update, InlineQueryKind)
}

func (bot *BotFramework) handle(update *tgbotapi.Update, event Kind) error {
	chatID := bot.GetChatID(update)
	if cmd, err := bot.storage.Get(string(event), "", chatID); err == nil {
		return cmd.Exec(update)
	} else if cmd, err = bot.storage.Get(string(event), "", 0); err == nil {
		return cmd.Exec(update)
	}
	return errors.New("no handlers")
}

// RegisterPlainTextHandler binds handler for plain text message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterPlainTextHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(PlainTextKind), "", chatID, f)
	return nil
}

// UnregisterPlainTextHandler deletes handler for given chat
func (bot *BotFramework) UnregisterPlainTextHandler(chatID int64) error {
	bot.storage.Unset(string(PlainTextKind), "", chatID)
	return nil
}

// RegisterContactHandler binds handler for contact message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterContactHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(ContactKind), "", chatID, f)
	return nil
}

// UnregisterContactHandler deletes handler for given chat
func (bot *BotFramework) UnregisterContactHandler(chatID int64) error {
	bot.storage.Unset(string(ContactKind), "", chatID)
	return nil
}

// RegisterPhotoHandler binds handler for photo message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterPhotoHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(PhotoKind), "", chatID, f)
	return nil
}

// UnregisterPhotoHandler deletes handler for given chat
func (bot *BotFramework) UnregisterPhotoHandler(chatID int64) error {
	bot.storage.Unset(string(PhotoKind), "", chatID)
	return nil
}

// RegisterFileHandler binds handler for file from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterFileHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(FileKind), "", chatID, f)
	return nil
}

// UnregisterFileHandler deletes handler for given chat
func (bot *BotFramework) UnregisterFileHandler(chatID int64) error {
	bot.storage.Unset(string(FileKind), "", chatID)
	return nil
}

// RegisterStickerHandler binds handler for sticker from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterStickerHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(StickerKind), "", chatID, f)
	return nil
}

// UnregisterStickerHandler deletes handler for given chat
func (bot *BotFramework) UnregisterStickerHandler(chatID int64) error {
	bot.storage.Unset(string(StickerKind), "", chatID)
	return nil
}

// RegisterAudioHandler binds handler for audio message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterAudioHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(AudioKind), "", chatID, f)
	return nil
}

// UnregisterAudioHandler deletes handler for given chat
func (bot *BotFramework) UnregisterAudioHandler(chatID int64) error {
	bot.storage.Unset(string(AudioKind), "", chatID)
	return nil
}

// RegisterVideoHandler binds handler for video message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVideoHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(VideoKind), "", chatID, f)
	return nil
}

// UnregisterVideoHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVideoHandler(chatID int64) error {
	bot.storage.Unset(string(VideoKind), "", chatID)
	return nil
}

// RegisterVideoNoteHandler binds handler for video_note message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVideoNoteHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(VideoNoteKind), "", chatID, f)
	return nil
}

// UnregisterVideoNoteHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVideoNoteHandler(chatID int64) error {
	bot.storage.Unset(string(VideoNoteKind), "", chatID)
	return nil
}

// RegisterVoiceHandler binds handler for voice message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVoiceHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(VoiceKind), "", chatID, f)
	return nil
}

// UnregisterVoiceHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVoiceHandler(chatID int64) error {
	bot.storage.Unset(string(VoiceKind), "", chatID)
	return nil
}

// RegisterVenueHandler binds handler for venue message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVenueHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(VenueKind), "", chatID, f)
	return nil
}

// UnregisterVenueHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVenueHandler(chatID int64) error {
	bot.storage.Unset(string(VenueKind), "", chatID)
	return nil
}

// RegisterLocationHandler binds handler for location message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterLocationHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(LocationKind), "", chatID, f)
	return nil
}

// UnregisterLocationHandler deletes handler for given chat
func (bot *BotFramework) UnregisterLocationHandler(chatID int64) error {
	bot.storage.Unset(string(LocationKind), "", chatID)
	return nil
}

// RegisterUniversalHandler binds handler for any message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterUniversalHandler(f storage.Command, chatID int64) error {
	bot.storage.Set(string(AnyKind), "", chatID, f)
	return nil
}

// UnregisterUniversalHandler deletes handler for given chat
func (bot *BotFramework) UnregisterUniversalHandler(chatID int64) error {
	bot.storage.Unset(string(AnyKind), "", chatID)
	return nil
}
