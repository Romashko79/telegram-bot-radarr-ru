package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// BotConfig ...
type Config struct {
	TelegramBotToken string
	AllowedChatIDs   map[int64]bool
	MaxItems         int
	IgnoreTags       bool
	RadarrProtocol   string
	RadarrHostname   string
	RadarrPort       int
	RadarrAPIKey     string
	RadarrBaseUrl    string
}

func LoadConfig() (Config, error) {
	var config Config

	config.TelegramBotToken = os.Getenv("RBOT_TELEGRAM_BOT_TOKEN")
	allowedUserIDs := os.Getenv("RBOT_BOT_ALLOWED_USERIDS")
	botMaxItems := os.Getenv("RBOT_BOT_MAX_ITEMS")
	botIgnoreTags := os.Getenv("RBOT_BOT_IGNORE_TAGS")
	config.RadarrProtocol = os.Getenv("RBOT_RADARR_PROTOCOL")
	config.RadarrHostname = os.Getenv("RBOT_RADARR_HOSTNAME")
	radarrPort := os.Getenv("RBOT_RADARR_PORT")
	config.RadarrAPIKey = os.Getenv("RBOT_RADARR_API_KEY")
	config.RadarrBaseUrl = os.Getenv("RBOT_RADARR_BASE_URL")

	// Validate required fields
	if config.TelegramBotToken == "" {
		return config, errors.New("Переменная RBOT_TELEGRAM_BOT_TOKEN пуста или не задана")
	}
	if allowedUserIDs == "" {
		return config, errors.New("Переменная RBOT_BOT_ALLOWED_USERIDS пуста или не задана")
	}
	if botMaxItems == "" {
		return config, errors.New("Переменная RBOT_BOT_MAX_ITEMS пуста или не задана")
	}
	if botIgnoreTags == "" {
		return config, errors.New("Переменная RBOT_BOT_IGNORE_TAGS пуста или не задана")
	}
	// Normalize and validate RBOT_RADARR_PROTOCOL
	config.RadarrProtocol = strings.ToLower(config.RadarrProtocol)
	if config.RadarrProtocol != "http" && config.RadarrProtocol != "https" {
		return config, errors.New("Переменная RBOT_RADARR_PROTOCOL должна быть http или https")
	}
	if config.RadarrHostname == "" {
		return config, errors.New("Переменная RBOT_RADARR_HOSTNAME пуста или не задана")
	}
	if radarrPort == "" {
		return config, errors.New("Переменная RBOT_RADARR_PORT пуста или не задана")
	}
	if config.RadarrAPIKey == "" {
		return config, errors.New("Переменная RBOT_RADARR_API_KEY пуста или не задана")
	}

	// Parsing RBOT_BOT_MAX_ITEMS as a number
	maxItems, err := strconv.Atoi(botMaxItems)
	if err != nil {
		return config, errors.New("Переменная RBOT_BOT_MAX_ITEMS должна быть числом")
	}
	config.MaxItems = maxItems

	// Parsing RBOT_BOT_IGNORE_TAGS as a boolean
	ignoreTags, err := strconv.ParseBool(botIgnoreTags)
	if err != nil {
		return config, errors.New("Переменная RBOT_BOT_IGNORE_TAGS должна быть true или false")
	}
	config.IgnoreTags = ignoreTags

	// Parsing RBOT_BOT_ALLOWED_USERIDS as a list of integers
	userIDs := strings.Split(allowedUserIDs, ",")
	parsedUserIDs := make(map[int64]bool)
	for _, id := range userIDs {
		parsedID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return config, fmt.Errorf("Переменная RBOT_BOT_ALLOWED_USERIDS содержит нецелое значение: %s", err)
		}
		parsedUserIDs[parsedID] = true
	}
	config.AllowedChatIDs = parsedUserIDs

	// Parsing RBOT_RADARR_PORT as a number
	port, err := strconv.Atoi(radarrPort)
	if err != nil {
		return config, errors.New("Переменная RBOT_RADARR_PORT должна быть числом")
	}
	config.RadarrPort = port

	return config, nil
}
