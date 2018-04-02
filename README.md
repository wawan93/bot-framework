# Telegram bot framework [![GoDoc](https://godoc.org/github.com/wawan93/my-bot-framework?status.svg)](https://godoc.org/github.com/wawan93/my-bot-framework) [![Go Report Card](https://goreportcard.com/badge/github.com/wawan93/my-bot-framework)](https://goreportcard.com/report/github.com/wawan93/my-bot-framework)
`tgbot` is an extension for [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) package.
It helps you easily bind functions to handle any messages and callback queries

## Getting started
Install package:
```
go get -u github.com/wawan93/my-bot-framework
```

## Usage 
```go
package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wawan93/my-bot-framework"
)

func Start(bot *tgbot.BotFramework, update *tgbotapi.Update) error {
	chatID := bot.GetChatID(update)
	msg := tgbotapi.NewMessage(chatID, "Hello, World!")
	_, err := bot.Send(msg)
	return err
}

func main() {
	token := "123:YOUR-TOKEN"
	api, _ := tgbotapi.NewBotAPI(token)

	u := tgbotapi.NewUpdate(0)
	updates, _ := api.GetUpdatesChan(u)
  
	// extend api
	bot := tgbot.NewBotFramework(api)
  
	// bind handler Start for "/start" command in chat 0 (any chat)
	bot.RegisterCommand("/start", Start, 0)

	// endless loop handles updates from channel
	bot.HandleUpdates(updates)
}

```
