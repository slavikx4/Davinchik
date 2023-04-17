package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

var Users map[int]*User

type User struct {
	UserID        int
	ChatID        int64
	Name          string
	IsName        bool //установленно ли имя
	Bio           string
	IsBio         bool // установленна ли биография
	Age           int
	IsAge         bool // установлен ли возраст
	PhotoPath     string
	IsPhoto       bool // устоновленно ли фото у пользователя
	ExpectedPhoto bool // ожидается ли получение фото от пользователя

	channelInput  chan *tgbotapi.Update
	channelOutput chan OutputAnswer
}

func NewUser(userID int, chatID int64) *User {
	return &User{
		UserID: userID,
		ChatID: chatID,
	}
}

type OutputAnswer struct {
	ChatID    int
	Text      string
	PhotoPath string // нужно ещё подумать
}

func (u *User) ChannelInputHandler(update *tgbotapi.Update) {
	u.channelInput <- update
}

func (u *User) GetChannelOutput() chan OutputAnswer {
	return u.channelOutput
}
