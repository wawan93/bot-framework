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
		go func() {
			err := bot.handleUpdate(&update)
			if err != nil {
				if update.Message != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
			}
		}()
	}
}

func (bot *BotFramework) handleUpdate(update *tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("no message")
	}
	if update.Message.IsCommand() {
		return bot.handleCommand(update)
	}
	if update.Message.Text != "" {
		err := bot.handleKeyboardCommand(update)
		if err == nil {
			return nil
		}
		if err.Error() != "command not found" {
			return err
		}
	}
	return errors.New("not handled")
}

func (bot *BotFramework) RegisterCommand(c *Command) error {
	if c.Name[0] != '/' {
		return errors.New("command must start with slash")
	}

	bot.commands[c.Name] = c
	return nil
}

func (bot *BotFramework) handleCommand(update *tgbotapi.Update) error {
	if command, ok := bot.commands["/" + update.Message.Command()]; ok {
		return command.Handler(bot, update)
	}
	return errors.New("command not found")
}

func (bot *BotFramework) RegisterKeyboardCommand(c *Command) error {
	if c.Name[0] == '/' {
		return errors.New("keyboard command must not start with slash")
	}
	bot.commands[c.Name] = c
	return nil
}

func (bot *BotFramework) handleKeyboardCommand(update *tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("no message")
	}
	if command, ok := bot.commands[update.Message.Text]; ok {
		return command.Handler(bot, update)
	}
	return errors.New("command not found")
}