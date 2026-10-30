package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golift.io/starr"
	"golift.io/starr/radarr"

	"github.com/woiza/telegram-bot-radarr/pkg/bot"
	"github.com/woiza/telegram-bot-radarr/pkg/config"
)

func main() {
	fmt.Println("Запуск бота...")

	// get config from environment variables
	config, err := config.LoadConfig()
	if err != nil {
		// Handle error: configuration is incomplete or invalid
		log.Fatal(err)
	}

	b, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Fatal("Ошибка при запуске бота: ", err)
	}

	fmt.Printf("Авторизован как %v\n", b.Self.UserName)

	radarrConfig := starr.New(config.RadarrAPIKey, fmt.Sprintf("%v://%v:%v%v", config.RadarrProtocol, config.RadarrHostname, config.RadarrPort, config.RadarrBaseUrl), 0)
	radarrServer := radarr.New(radarrConfig)

	botInstance := bot.New(&config, b, radarrServer)

	// Receive Radarr webhook notifications (only if the webhook port is configured)
	botInstance.StartWebhook()

	// Channel for receiving updates from the bot API
	updates := make(chan tgbotapi.Update)
	defer close(updates)

	// Start a goroutine to fetch updates from the bot API and send to the updates channel
	go func() {
		lastUpdateID := 0
		for {
			updateConfig := tgbotapi.NewUpdate(lastUpdateID + 1)
			updateConfig.Timeout = 60

			updatesChannel := b.GetUpdatesChan(updateConfig)
			if err != nil {
				log.Println("Ошибка получения канала обновлений:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for update := range updatesChannel {
				updates <- update // Send updates to the updates channel
				lastUpdateID = update.UpdateID
			}
		}
	}()

	// Start a goroutine to handle updates concurrently
	go botInstance.HandleUpdates(updates)

	// This can be a long-running process to handle incoming updates
	select {}
}
