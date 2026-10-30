package bot

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// whMap is a decoded JSON object. Maps are used instead of structs so that
// a changed field type in a newer Radarr version can never break the whole payload.
type whMap = map[string]interface{}

func whObj(m whMap, key string) whMap {
	v, _ := m[key].(map[string]interface{})
	return v
}

func whStr(m whMap, key string) string {
	switch v := m[key].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	}
	return ""
}

func whNum(m whMap, key string) float64 {
	v, _ := m[key].(float64)
	return v
}

func whBool(m whMap, key string) bool {
	v, _ := m[key].(bool)
	return v
}

func whList(m whMap, key string) []interface{} {
	v, _ := m[key].([]interface{})
	return v
}

func whSize(bytes float64) string {
	switch {
	case bytes <= 0:
		return ""
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.2f ГБ", bytes/(1024*1024*1024))
	default:
		return fmt.Sprintf("%.0f МБ", bytes/(1024*1024))
	}
}

// whAdd appends "label: value" to lines when the value is not empty.
func whAdd(lines []string, label, value string) []string {
	if value == "" {
		return lines
	}
	return append(lines, fmt.Sprintf("%s: %s", label, value))
}

func whDeleteReason(reason string) string {
	switch reason {
	case "":
		return ""
	case "manual":
		return "удалён вручную"
	case "missingFromDisk":
		return "файл пропал с диска"
	case "upgrade":
		return "заменён файлом лучшего качества"
	}
	return reason
}

func whMovieTitle(movie whMap) string {
	title := whStr(movie, "title")
	if title == "" {
		return "Неизвестный фильм"
	}
	if year := whStr(movie, "year"); year != "" && year != "0" {
		return fmt.Sprintf("%s (%s)", title, year)
	}
	return title
}

func whAddMethod(method string) string {
	switch method {
	case "":
		return ""
	case "manual":
		return "вручную"
	case "list":
		return "из списка"
	case "collection":
		return "из коллекции"
	}
	return method
}

func buildWebhookMessage(p whMap) string {
	event := whStr(p, "eventType")
	title := whMovieTitle(whObj(p, "movie"))

	var lines []string
	switch event {
	case "Test":
		return "✅ Тестовое уведомление от Radarr получено, бот принимает события."

	case "Grab":
		release := whObj(p, "release")
		lines = append(lines, "⬇️ Начата загрузка", title)
		lines = whAdd(lines, "Качество", whStr(release, "quality"))
		lines = whAdd(lines, "Размер", whSize(whNum(release, "size")))
		lines = whAdd(lines, "Релиз", whStr(release, "releaseTitle"))
		lines = whAdd(lines, "Индексатор", whStr(release, "indexer"))
		lines = whAdd(lines, "Клиент загрузки", whStr(p, "downloadClient"))

	case "Download":
		file := whObj(p, "movieFile")
		if whBool(p, "isUpgrade") {
			lines = append(lines, "🔄 Фильм заменён файлом лучшего качества", title)
		} else {
			lines = append(lines, "✅ Фильм загружен и импортирован", title)
		}
		lines = whAdd(lines, "Качество", whStr(file, "quality"))
		lines = whAdd(lines, "Размер", whSize(whNum(file, "size")))

	case "Rename":
		lines = append(lines, "✏️ Файлы переименованы", title)

	case "MovieAdded":
		lines = append(lines, "➕ Фильм добавлен в библиотеку", title)
		lines = whAdd(lines, "Способ", whAddMethod(whStr(p, "addMethod")))

	case "MovieDelete":
		lines = append(lines, "🗑 Фильм удалён из библиотеки", title)
		if whBool(p, "deletedFiles") {
			lines = append(lines, "Файлы тоже удалены")
		}

	case "MovieFileDelete":
		lines = append(lines, "🗑 Файл фильма удалён", title)
		lines = whAdd(lines, "Причина", whDeleteReason(whStr(p, "deleteReason")))

	case "Health":
		icon := "⚠️"
		if whStr(p, "level") == "error" {
			icon = "❌"
		}
		lines = append(lines, icon+" Проблема в Radarr")
		lines = whAdd(lines, "Сообщение", whStr(p, "message"))
		lines = whAdd(lines, "Подробнее", whStr(p, "wikiUrl"))

	case "HealthRestored":
		lines = append(lines, "✅ Проблема в Radarr устранена")
		lines = whAdd(lines, "Сообщение", whStr(p, "message"))

	case "ApplicationUpdate":
		lines = append(lines, "⬆️ Radarr обновлён")
		lines = whAdd(lines, "Было", whStr(p, "previousVersion"))
		lines = whAdd(lines, "Стало", whStr(p, "newVersion"))

	case "ManualInteractionRequired":
		lines = append(lines, "✋ Требуется ручное вмешательство", title)
		lines = whAdd(lines, "Загрузка", whStr(whObj(p, "downloadInfo"), "title"))

	default:
		lines = append(lines, "ℹ️ Событие Radarr: "+event)
		if title != "Неизвестный фильм" {
			lines = append(lines, title)
		}
	}

	return strings.Join(lines, "\n")
}


// StartWebhook starts the HTTP server that receives Radarr webhook notifications.
// It does nothing when RBOT_WEBHOOK_PORT is not set.
func (b *Bot) StartWebhook() {
	if b.Config.WebhookPort == 0 {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", b.handleWebhook)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", b.Config.WebhookPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Приём уведомлений Radarr запущен на порту %d (путь /webhook)", b.Config.WebhookPort)
		if err := server.ListenAndServe(); err != nil {
			log.Println("Ошибка сервера вебхуков:", err)
		}
	}()
}

func (b *Bot) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if b.Config.WebhookToken != "" {
		token := r.URL.Query().Get("token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(b.Config.WebhookToken)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var payload whMap
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Println("Не удалось разобрать уведомление:", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	text := buildWebhookMessage(payload)
	if runes := []rune(text); len(runes) > 4000 {
		text = string(runes[:4000]) + "…"
	}

	for chatID, allowed := range b.Config.AllowedChatIDs {
		if !allowed {
			continue
		}
		msg := tgbotapi.NewMessage(chatID, text)
		msg.DisableWebPagePreview = true
		b.sendMessage(msg)
	}

	w.WriteHeader(http.StatusOK)
}
