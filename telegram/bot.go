package telegram

import (
	log "Davinchik/pkg/logger"
	"flag"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"strconv"
	"time"
)

const (
	WebURL = "YOUR-WEB"
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("❤️"),
		tgbotapi.NewKeyboardButton("🥴"),
		tgbotapi.NewKeyboardButton("😴"),
	),
)
var letsGOKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("let`s go"),
	),
)
var registrationKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("зарегестрироваться"),
	),
)

var bot *tgbotapi.BotAPI

func StartBot() {

	log.Logger.Process.Println("попытка установить токен")

	log.Logger.Process.Println("токен установлен")

	log.Logger.Process.Println("попытка создать бота")
	var err error
	bot, err = tgbotapi.NewBotAPI(mustToken())
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
		if err := http.ListenAndServe(mustToken(), nil); err != nil {
			log.Logger.Error.Fatalln("Не удалось начать прослуживание и обслуживание: ", err)
		}
	}()
	log.Logger.Process.Println("успешно началось прослушивание и обслуживание")

	gettingUpdates()
}

func gettingUpdates() {
	log.Logger.Process.Println("бот начал получать обновления")
	updates := bot.ListenForWebhook("/")

	for update := range updates {
		go func(update *tgbotapi.Update) {
			userID := update.Message.From.ID
			log.Logger.Chat.Println("поступило новое обновление от ", userID, " : ", update.Message.Text)

			if user, err := GetUserInCache(userID); err == nil {
				switch update.Message.Text {
				case "let`s go", "🥴", "❤️", "😴":
					break
				default:
					sendMessageError(update)
					return
				}
				user.channelInput = make(chan *tgbotapi.Update)
				user.channelOutput = make(chan OutputAnswer)
				log.Logger.Process.Println("обновление передано в приёмный канал обработчик пользователю: ", userID)
				go func(u *User, b *tgbotapi.BotAPI) {
					output := u.GetChannelOutput()
					for ans := range output {
						log.Logger.Chat.Println("Новое сообщение отправленно ", u.ChatID, " ", u.Name, " : ", ans.Text)
						sendMessageAnswer(ans)
						close(output)
					}
				}(user, bot)
				go user.HandlerCommand()
				time.Sleep(time.Second * 1)
				user.ChannelInputHandler(update)

			} else if user, ok := CheckUserInBase(userID); ok {
				if err := setUserInCache(user); err != nil {
					log.Logger.Error.Printf("не удалось установить в cache: ", err)
					sendMessageError(update)
				} else {
					answer := tgbotapi.NewMessage(user.ChatID, msgReUpdate)
					answer.ReplyMarkup = letsGOKeyboard
					sendMessageReUpdate(&answer)
				}
			} else {
				user, ok := CreatingUsers[userID]
				if ok {
					if user.ready {
						if err := AddUserInDataBase(user); err != nil {
							log.Logger.Error.Println(err)
							sendMessageError(update)
						} else {
							log.Logger.Process.Println("зарегестрирован новый пользователь ", userID)
							if err := setUserInCache(user); err != nil {
								log.Logger.Error.Printf("не удалось установить в cache: ", err)
								sendMessageError(update)
							} else {
								answer := tgbotapi.NewMessage(user.ChatID, msgMenu)
								answer.ReplyMarkup = letsGOKeyboard
								sendMessageMenu(&answer)
							}
						}
						delete(CreatingUsers, userID)
					} else {
						switch {
						case user.Name == "":
							sendMessageEnterGender(update)
							user.Name = update.Message.Text
						case user.Gender == "":
							sendMessageEnterAge(update)
							user.Gender = update.Message.Text
						case user.Age == 0:
							sendMessageEnterBio(update)
							user.Age, _ = strconv.Atoi(update.Message.Text)
						case user.Bio == "":
							user.Bio = update.Message.Text
							sendMessageEnterPhoto(update)
						case user.PhotoID == "":
							for _, photoSize := range *update.Message.Photo {
								fileConf := tgbotapi.FileConfig{FileID: photoSize.FileID}
								file, err := bot.GetFile(fileConf)
								if err != nil {
									log.Logger.Error.Println("нет фотографии: ", err)
								}
								user.PhotoID = file.FileID
								break
							}
							user.ready = true
							ans := tgbotapi.NewMessage(user.ChatID, "Регестрируемся?")
							ans.ReplyMarkup = registrationKeyboard
							bot.Send(tgbotapi.NewMessage(ans.ChatID, ans.Text))
						}
					}
				} else {
					CreatingUsers[userID] = &User{
						UserID:        userID,
						UserName:      update.Message.From.UserName,
						ChatID:        update.Message.Chat.ID,
						channelInput:  make(chan *tgbotapi.Update),
						channelOutput: make(chan OutputAnswer),
					}
					sendMessageHello(update)
					sendMessageEnterName(update)
				}

			}
			return
		}(&update)
	}
}

func sendMessageAnswer(ans OutputAnswer) {
	answer := tgbotapi.NewMessage(ans.ChatID, ans.Text)
	answer.ReplyMarkup = mainKeyboard
	if ans.PhotoID != "" {
		if _, err := bot.Send(tgbotapi.NewPhotoShare(ans.ChatID, ans.PhotoID)); err != nil {
			log.Logger.Error.Printf("не удалось отправить фотографию пользователю ", ans.ChatID, " : ", err)
		}
	}
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", ans.ChatID, " : ", err)
	} else {
		log.Logger.Chat.Println("отправлено сообщение пользователю", ans.ChatID, " : ", ans.Text)
	}
}

func sendMessageEnterPhoto(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgEnterPhoto)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func sendMessageMenu(answer *tgbotapi.MessageConfig) {
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", answer.ChatID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", answer.ChatID, " : ", answer.Text)
}

func sendMessageHello(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgHello)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func sendMessageEnterBio(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgEnterBio)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func sendMessageEnterAge(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgEnterAge)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func sendMessageEnterGender(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgEnterGender)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func sendMessageEnterName(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgEnterName)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func sendMessageReUpdate(answer *tgbotapi.MessageConfig) {
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", answer.ChatID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", answer.ChatID, " : ", msgReUpdate)
}

func sendMessageError(update *tgbotapi.Update) {
	answer := tgbotapi.NewMessage(update.Message.Chat.ID, msgError)
	if _, err := bot.Send(answer); err != nil {
		log.Logger.Error.Printf("не удалось отправить сообщение пользователю ", update.Message.Chat.ID, " : ", err)
	}
	log.Logger.Chat.Println("отправлено сообщение пользователю", update.Message.From.ID, " : ", answer.Text)
}

func mustToken() string {
	token := flag.String("telegram-bot-token", "", "necessary input telegram bot token")
	flag.Parse()
	if *token == "" {
		log.Logger.Error.Fatalln("!не установлен токен телеграм бота!")
	}
	return *token
}

func mustPort() string {
	port := flag.String("port", "", "necessary input port for Listener")
	flag.Parse()
	if *port == "" {
		log.Logger.Error.Fatalln("!не установлен порт для прослушивания!")
	}
	return *port
}
