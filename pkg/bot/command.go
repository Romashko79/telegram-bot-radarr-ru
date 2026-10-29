package bot

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/woiza/telegram-bot-radarr/pkg/utils"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

func (b *Bot) handleCommand(update tgbotapi.Update, r *radarr.Radarr) {

	chatID, err := b.getChatID(update)
	if err != nil {
		fmt.Printf("Не удалось обработать команду: %v", err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "")

	switch update.Message.Command() {

	case "q", "query", "add", "Q", "Query", "Add":
		b.setActiveCommand(chatID, AddMovieCommand)
		b.processAddCommand(update, chatID, r)

	case "movies", "library", "l":
		b.setActiveCommand(chatID, LibraryMenuCommand)
		b.processLibraryCommand(update, chatID, r)

	case "delete", "remove", "Delete", "Remove", "d":
		b.setActiveCommand(chatID, DeleteMovieCommand)
		b.processDeleteCommand(update, chatID, r)

	case "clear", "cancel", "stop":
		b.clearState(update)
		msg.Text = "Все команды сброшены"
		b.sendMessage(msg)

	case "diskspace", "disk", "free", "rootfolder", "rootfolders":
		rootFolders, err := r.GetRootFolders()
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = utils.PrepareRootFolders(rootFolders)
		msg.ParseMode = "MarkdownV2"
		msg.DisableWebPagePreview = true
		b.sendMessage(msg)

	case "up", "upcoming":
		calendar := radarr.Calendar{
			Start:       time.Now(),
			End:         time.Now().AddDate(0, 0, 30), // 30 days
			Unmonitored: *starr.True(),
		}
		upcoming, err := r.GetCalendar(calendar)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		if len(upcoming) == 0 {
			msg.Text = "В ближайшие 30 дней релизов нет"
			b.sendMessage(msg)
			break
		}
		b.sendUpcoming(upcoming, &msg)

	case "rss", "RSS":
		command := radarr.CommandRequest{
			Name:     "RssSync",
			MovieIDs: []int64{},
		}
		_, err := r.SendCommand(&command)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = "RSS-синхронизация запущена"
		b.sendMessage(msg)

	case "searchmonitored":
		movies, err := r.GetMovie(0)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		var monitoredMoviesIDs []int64
		for _, movie := range movies {
			if movie.Monitored {
				monitoredMoviesIDs = append(monitoredMoviesIDs, movie.ID)
			}
		}
		command := radarr.CommandRequest{
			Name:     "MoviesSearch",
			MovieIDs: monitoredMoviesIDs,
		}
		_, err = r.SendCommand(&command)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = "Поиск отслеживаемых фильмов запущен"
		b.sendMessage(msg)

	case "updateAll", "updateall":
		movies, err := r.GetMovie(0)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		var allMoviesIDs []int64
		for _, movie := range movies {
			allMoviesIDs = append(allMoviesIDs, movie.ID)
		}
		command := radarr.CommandRequest{
			Name:     "RefreshMovie",
			MovieIDs: allMoviesIDs,
		}
		_, err = r.SendCommand(&command)
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = "Обновление всех фильмов запущено"
		b.sendMessage(msg)

	case "system", "System", "systemstatus", "Systemstatus":
		status, err := r.GetSystemStatus()
		if err != nil {
			msg.Text = err.Error()
			fmt.Println(err)
			b.sendMessage(msg)
			break
		}
		msg.Text = prettyPrint(status)
		b.sendMessage(msg)

	case "getid", "id":
		msg.Text = fmt.Sprintf("Ваш ID пользователя: %d", chatID)
		b.sendMessage(msg)

	default:
		msg.Text = fmt.Sprintf("Привет, %v!\n", update.Message.From)
		msg.Text += "Доступные команды:\n\n"
		msg.Text += "/q [фильм] - поиск фильма\n"
		msg.Text += "/library [фильм] - управление фильмами\n"
		msg.Text += "/delete [фильм] - удаление фильма\n"
		msg.Text += "/clear - сбросить все отправленные команды\n"
		msg.Text += "/free - свободное место на дисках\n"
		msg.Text += "/up - ближайшие релизы на 30 дней вперёд\n"
		msg.Text += "/rss - запустить RSS-синхронизацию\n"
		msg.Text += "/searchmonitored - поиск всех отслеживаемых фильмов\n"
		msg.Text += "/updateall - обновить метаданные и пересканировать файлы и папки\n"
		msg.Text += "/system - конфигурация Radarr\n"
		msg.Text += "/id - ваш Telegram ID"
		b.sendMessage(msg)
	}
}
