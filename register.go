package tgbot

import "errors"

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
	bot.mu.Lock()
	defer bot.mu.Unlock()

	if _, ok := bot.commands[name]; !ok {
		bot.commands[name] = make(map[int64]CommonHandler, 1)
	}
	bot.commands[name][chatID] = f
	return nil
}

// UnregisterCommand deletes handler for command name in given chat
func (bot *BotFramework) UnregisterCommand(name string, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.commands[name], chatID)
	return nil
}

// RegisterCallbackQueryHandler binds handler for callback data
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterCallbackQueryHandler(f CommonHandler, dataStartsWith string, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	if _, ok := bot.callbackQueryHandlers[dataStartsWith]; !ok {
		bot.callbackQueryHandlers[dataStartsWith] = make(map[int64]CommonHandler)
	}
	bot.callbackQueryHandlers[dataStartsWith][chatID] = f
	return nil
}

// UnregisterCallbackQueryHandler deletes handler for given chat
func (bot *BotFramework) UnregisterCallbackQueryHandler(dataStartsWith string, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.callbackQueryHandlers[dataStartsWith], chatID)
	return nil
}

// RegisterInlineQueryHandler binds handler for query
// If userID = 0, command will work for any user
func (bot *BotFramework) RegisterInlineQueryHandler(f CommonHandler, query string, userID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	if _, ok := bot.inlineQueryHandlers[query]; !ok {
		bot.inlineQueryHandlers[query] = make(map[int64]CommonHandler)
	}
	bot.inlineQueryHandlers[query][userID] = f
	return nil
}

// UnregisterInlineQueryHandler deletes handler for given user
func (bot *BotFramework) UnregisterInlineQueryHandler(query string, userID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.inlineQueryHandlers[query], userID)
	return nil
}

// RegisterPlainTextHandler binds handler for plain text message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterPlainTextHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["plain"][chatID] = f
	return nil
}

// UnregisterPlainTextHandler deletes handler for given chat
func (bot *BotFramework) UnregisterPlainTextHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["plain"], chatID)
	return nil
}

// RegisterContactHandler binds handler for contact message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterContactHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["contact"][chatID] = f
	return nil
}

// UnregisterContactHandler deletes handler for given chat
func (bot *BotFramework) UnregisterContactHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["contact"], chatID)
	return nil
}

// RegisterPhotoHandler binds handler for photo message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterPhotoHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["photo"][chatID] = f
	return nil
}

// UnregisterPhotoHandler deletes handler for given chat
func (bot *BotFramework) UnregisterPhotoHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["photo"], chatID)
	return nil
}

// RegisterFileHandler binds handler for file from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterFileHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["file"][chatID] = f
	return nil
}

// UnregisterFileHandler deletes handler for given chat
func (bot *BotFramework) UnregisterFileHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["file"], chatID)
	return nil
}

// RegisterStickerHandler binds handler for sticker from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterStickerHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["sticker"][chatID] = f
	return nil
}

// UnregisterStickerHandler deletes handler for given chat
func (bot *BotFramework) UnregisterStickerHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["sticker"], chatID)
	return nil
}

// RegisterAudioHandler binds handler for audio message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterAudioHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["audio"][chatID] = f
	return nil
}

// UnregisterAudioHandler deletes handler for given chat
func (bot *BotFramework) UnregisterAudioHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["audio"], chatID)
	return nil
}

// RegisterVideoHandler binds handler for video message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVideoHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["video"][chatID] = f
	return nil
}

// UnregisterVideoHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVideoHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["video"], chatID)
	return nil
}

// RegisterVideoNoteHandler binds handler for video_note message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVideoNoteHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["video_note"][chatID] = f
	return nil
}

// UnregisterVideoNoteHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVideoNoteHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["video_note"], chatID)
	return nil
}

// RegisterVoiceHandler binds handler for voice message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVoiceHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["voice"][chatID] = f
	return nil
}

// UnregisterVoiceHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVoiceHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["voice"], chatID)
	return nil
}

// RegisterVenueHandler binds handler for venue message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterVenueHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["venue"][chatID] = f
	return nil
}

// UnregisterVenueHandler deletes handler for given chat
func (bot *BotFramework) UnregisterVenueHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["venue"], chatID)
	return nil
}

// RegisterLocationHandler binds handler for location message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterLocationHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["location"][chatID] = f
	return nil
}

// UnregisterLocationHandler deletes handler for given chat
func (bot *BotFramework) UnregisterLocationHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["location"], chatID)
	return nil
}

// RegisterUniversalHandler binds handler for any message from given chat
// If chatID = 0, command will work in any chat
func (bot *BotFramework) RegisterUniversalHandler(f CommonHandler, chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	bot.handlers["any"][chatID] = f
	return nil
}

// UnregisterUniversalHandler deletes handler for given chat
func (bot *BotFramework) UnregisterUniversalHandler(chatID int64) error {
	bot.mu.Lock()
	defer bot.mu.Unlock()
	delete(bot.handlers["any"], chatID)
	return nil
}
