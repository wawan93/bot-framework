# Telegram bot framework [![GoDoc](https://godoc.org/github.com/wawan93/bot-framework?status.svg)](https://godoc.org/github.com/wawan93/bot-framework) [![Go Report Card](https://goreportcard.com/badge/github.com/wawan93/bot-framework)](https://goreportcard.com/report/github.com/wawan93/bot-framework)
`tgbot` is an extension for [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) package.
It helps you easily bind functions to handle any messages and callback queries

## Getting started
Install package:
```
go get -u github.com/wawan93/bot-framework
```

## Usage

```go
package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wawan93/bot-framework"
)

type StartCommand struct {
    bot *BotFramework
	Message string
	Field1  struct {
		Field2 int
	}
}

func (s StartCommand) Exec(bot *tgbot.BotFramework, update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)
	msg := tgbotapi.NewMessage(chatID, s.Message)
	_, err := bot.Send(msg)
	return err
}

func (s StartCommand) CommandName() string {
	return "start"
}

func (s StartCommand) Serialize() (string, error) {
	return s.Message, nil
}

func (s StartCommand) Deserialize(data string) tgbot.Command {
	return StartCommand{bot: s.bot, Message: data}
}

func main() {
	token := "123:YOUR-TOKEN"
	api, _ := tgbotapi.NewBotAPI(token)

	u := tgbotapi.NewUpdate(0)
	updates, _ := api.GetUpdatesChan(u)

	// extend api
	
	storage := tgbot.MysqlStorage{}
	
	bot := tgbot.NewBotFramework(api, storage)

	storage.RegisterFactories(StartCommand{bot: bot}, StopCommand{})

	start := StartCommand{Message: "Hello, World!"}

	// bind handler Start for "/start" command in chat 0 (any chat)
	bot.RegisterCommand("/start", start, 0)

	// endless loop handles updates from channel
	bot.HandleUpdates(updates)
}

```
