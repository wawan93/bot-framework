package bot_framework

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"errors"
)

type Command struct {
	Name    string
	Handler func(bot Sendable, update *tgbotapi.Update) error
}

type Sendable interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type BotFramework struct {
	Sendable
	commands map[string]*Command
	messages chan tgbotapi.Chattable
}

func NewBotFramework(api Sendable) *BotFramework {
	bot := BotFramework{
		api,
		make(map[string]*Command),
		make(chan tgbotapi.Chattable),
	}
	return &bot
}

func (bot *BotFramework) HandleUpdates(ch tgbotapi.UpdatesChannel) {
	for update := range ch {
		go bot.handleUpdate(&update)
	}
}

func (bot *BotFramework) handleUpdate(update *tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("no message")
	}
	if update.Message.IsCommand() {
		return bot.handleCommand(update)
	}
	return errors.New("not handled")
}

func (bot *BotFramework) RegisterCommand(c *Command) {
	bot.commands[c.Name] = c
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	if command, ok := bot.commands["/" + update.Message.Command()]; ok {
		return command.Handler(bot, update)
	}
	return errors.New("command not found")
}
