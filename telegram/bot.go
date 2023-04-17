package telegram

import (
	log "Davinchik/src/logger"
	"flag"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
)

const WebURL = "WEB-URL"

func Srartbot() {

	log.Logger.Process.Println("попытка установить токен")
	//telegramToken := mustToken()
	telegramToken := "6164489666:AAHdU3elOisXac9UCXxdEj3tVHAVvlS1Dmo"
	log.Logger.Process.Println("токен установлен")

	log.Logger.Process.Println("попытка создать бота")
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Logger.Error.Fatalln("!не удалоь создать бота: ", err)
	}
	log.Logger.Process.Println("бот успешно создан")

	log.Logger.Process.Println("попытка установить вебхук")
	webhook := tgbotapi.NewWebhook(WebURL)
	if _, err := bot.SetWebhook(webhook); err != nil {
		log.Logger.Error.Fatalln("!не удалось установить вебхук: ", err)
	}
	log.Logger.Process.Println("вебхук успешно установлен")

	log.Logger.Process.Println("попытка запустить прослушивание и обслуживание запросов")
	go func() {
		//if err := http.ListenAndServe(setPort(), nil); err != nil {
		if err := http.ListenAndServe(":5000", nil); err != nil {
			log.Logger.Error.Fatalln("Не удалось начать прослуживание и обслуживание: ", err)
		}
	}()
	log.Logger.Process.Println("успешно началось прослушивание и обслуживание")

	gettingUpdates(bot)
}

func gettingUpdates(bot *tgbotapi.BotAPI) {
	log.Logger.Process.Println("бот начал получать обновления")
	updates := bot.ListenForWebhook("/")

	for update := range updates {
		user_id := update.Message.From.ID
		log.Logger.Process.Printf("поступило новое обновление от %v", user_id)
		user, ok := Users[user_id]
		if ok {
			user.ChannelInputHandler(&update)
			log.Logger.Process.Println("обновление передано в приёмный канал обработчик пользователю: %v", user_id)
		} else {
			NewUser(user_id, update.Message.Chat.ID)
			log.Logger.Process.Println("создался новый пользователь")
			// тут следует как-то побёдить на новое сообщение от пользователя
		}
	}
}

func mustToken() string {
	token := flag.String("telegram-bot-token", "", "necessary input telegram bot token")
	flag.Parse()
	if *token == "" {
		log.Logger.Error.Fatalln("!не установлен токен телеграм бота!")
	}
	return *token
}

func setPort() string {
	port := flag.String("port", "", "necessary input port for Listener")
	flag.Parse()
	if *port == "" {
		log.Logger.Error.Fatalln("!не установлен порт для прослушивания!")
	}
	return *port
}
