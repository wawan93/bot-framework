package storage

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Your commands must implement Command interface
// to be registered in bot
type Command interface {
	Exec(update *tgbotapi.Update) error
}

type Serializable interface {
	CommandName() string
	Serialize() (string, error)
	Deserialize(data string) Command
}
