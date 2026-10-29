# Telegram-бот на Go для управления фильмами через Radarr
Этот Telegram-бот создан для управления фильмами через Radarr, менеджер коллекции фильмов. С его помощью можно искать, добавлять, редактировать, удалять и упорядочивать фильмы в вашей библиотеке Radarr. Бот написан на Go, потребляет менее 10 МБ оперативной памяти, не хранит состояние на диске (кроме журналов ошибок), а размер Docker-образа не превышает 10 МБ в сжатом виде. Поддерживаются архитектуры `arm32v7`, `arm64v8` и `x86_64`/`amd64`.

Бот построен на [golift/starr](https://github.com/golift/starr/) и [go-telegram-bot-api/telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api/) без каких-либо дополнительных зависимостей. Интерфейс бота полностью переведён на русский язык.

## Возможности и команды

<img src="screenshots/menu.png?raw=true" alt="q1" title="menu" width="300" />

### Запуск бота
<img src="screenshots/start.png?raw=true" alt="q1" title="start" width="300" />

### Поиск и добавление фильмов
``/q [фильм]`` или просто название фильма в сообщении: поиск фильма.\
Когда фильм найден, бот предлагает добавить его в библиотеку Radarr с разными настройками отслеживания. Если у вас одна корневая папка и один профиль качества, бот выберет их автоматически. Если вариантов несколько, бот попросит выбрать корневую папку и профиль качества. Если в Radarr заданы теги, их тоже можно выбрать.

<img src="screenshots/add_links.png?raw=true" alt="q1" title="add movie" width="300" />
<img src="screenshots/add_inline.png?raw=true" alt="q2" title="add movie" width="300" />
<img src="screenshots/add_confirmation.png?raw=true" alt="q3" title="add movie" width="300" />
<img src="screenshots/add_monsea.png?raw=true" alt="q4" title="add movie" width="300" />

### Управление фильмами
``/library [фильм]`` или ``/l [фильм]``: управление фильмами в библиотеке. Можно изменить профиль качества фильма (если в Radarr настроено больше одного) и теги. Также можно включать и отключать отслеживание, запускать поиск фильма и удалять его. Название необязательно: если его не указать, откроется меню фильтров.

<img src="screenshots/library.png?raw=true" alt="q1" title="library" width="300" />
<img src="screenshots/library_movie.png?raw=true" alt="q1" title="library movie" width="300" />


### Удаление фильмов
``/delete [фильм]`` или ``/d [фильм]``: удаление фильмов из библиотеки Radarr. Будьте осторожны: вместе с фильмом удаляются и связанные файлы. Название необязательно: если его не указать, все фильмы будут показаны кнопками, и можно выбрать несколько.

<img src="screenshots/delete_confirmation.png?raw=true" alt="q1" title="delete" width="300" />

### Отмена команд
``/clear`` или ``/cancel`` или ``/stop``:
сбрасывает все ранее введённые команды и состояние бота. Можно использовать в любой момент.

### Управление библиотекой
- ``/up`` или ``/upcoming``: список фильмов, которые выйдут в ближайшие 30 дней
- ``/rss``: запустить RSS-синхронизацию
- ``/searchmonitored``: поиск всех отслеживаемых фильмов
- ``/updateall``: обновить метаданные и пересканировать файлы и папки для всех фильмов


### Информация о системе
- ``/free`` или ``/diskspace``: свободное место на дисках, подключённых к серверу Radarr
- ``/system``: конфигурация Radarr
- ``/id`` или ``/getid``: показать ваш Telegram ID

## Установка и настройка
Бота можно собрать самостоятельно из исходного кода или использовать Docker-образ из GitHub Container Registry и Docker Hub:
- GitHub [ghcr.io/woiza/telegram-bot-radarr](https://github.com/woiza/telegram-bot-radarr/pkgs/container/telegram-bot-radarr)
- Docker Hub [woiza/telegram-bot-radarr](https://hub.docker.com/repository/docker/woiza/telegram-bot-radarr/)

Для работы бота нужно задать семь обязательных переменных окружения. Подробности смотрите в примере Docker Compose ниже. Перед запуском получите токен Telegram-бота и API-ключ Radarr. Также определите, кому разрешён доступ к боту (ID пользователя Telegram). Можно указать несколько пользователей списком ID. Как получить эти данные, описано в официальной документации:
- [Токен Telegram-бота](https://core.telegram.org/bots/tutorial/)
- [API-ключ Radarr](https://wiki.servarr.com/en/radarr/settings#security/)



### Сборка Docker-образа
```
docker buildx build --push --platform linux/amd64,linux/arm64,linux/arm/v7 --tag <repo>/<image>:<tag> .
```


### Пример Docker Compose
```
services:
    telegram-bot-radarr:
        image: woiza/telegram-bot-radarr
        mem_limit: 128M
        container_name: telegram-bot-radarr
#        depends_on:
#            - radarr
        restart: always
        environment:
            - RBOT_TELEGRAM_BOT_TOKEN=1460...:AAHlBW_mabVg...
            - RBOT_BOT_ALLOWED_USERIDS=123,987,-567 # ID пользователей Telegram, ID групп отрицательные
            - RBOT_BOT_MAX_ITEMS=10 # количество элементов на странице
            - RBOT_BOT_IGNORE_TAGS=false # true/false; true = бот не будет спрашивать теги (удобно при автотегировании)
            - RBOT_RADARR_PROTOCOL=http # http или https
            - RBOT_RADARR_PORT=7878
            - RBOT_RADARR_HOSTNAME=192.168.2.2 # IP-адрес или имя хоста
            - RBOT_RADARR_BASE_URL=/radarr # необязательно, например /radarr, в зависимости от настроек Radarr
            - RBOT_RADARR_API_KEY=1010d7...
```
### Команды для /setcommands в BotFather

```
q - поиск фильма
library - список всех фильмов (ВНИМАНИЕ: может быть большим)
delete - удаление фильма (ВНИМАНИЕ: список может быть большим)
clear - сбросить все ранее отправленные команды
free - свободное место на дисках
up - фильмы, которые выйдут в ближайшие 30 дней
rss - запустить RSS-синхронизацию
searchmonitored - поиск всех отслеживаемых фильмов
updateall - обновить метаданные и пересканировать файлы и папки
system - конфигурация Radarr
id - показать ваш Telegram ID
```

## Участие в разработке
Вы можете помочь проекту: отправляйте pull request'ы, сообщайте об ошибках и предлагайте улучшения. Любой вклад приветствуется!




## Пиво
Если вам нравится наша работа, угостите нас освежающим напитком.

<a href="https://paypal.me/telegramarrbots?country.x=EUR" target="_blank">
  <img src="pp.png?raw=true" alt="q1" title="donate" width="200">
</a>




## Лицензия
Этот Telegram-бот распространяется по [лицензии MIT](https://opensource.org/license/mit/).
