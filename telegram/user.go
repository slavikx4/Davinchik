package telegram

import (
	log "Davinchik/pkg/logger"
	"Davinchik/storage/cache"
	psql "Davinchik/storage/postresql"
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
)

var CreatingUsers map[int]*User

func init() {
	CreatingUsers = make(map[int]*User)
}

type User struct {
	UserID    int    `json:"user_id" bson:"user_id"`
	UserName  string `json:"user_name" bson:"user_name"`
	ChatID    int64  `json:"chat_id" bson:"chat_id"`
	Name      string `json:"name" bson:"name"`
	Gender    string `json:"gender" bson:"gender"`
	Age       int    `json:"age" bson:"age"`
	Bio       string `json:"bio" bson:"bio"`
	PhotoID   string `json:"photo_id" bson:"photo_id"`
	LastIntro int64  `bson:"last_intro"`

	channelInput  chan *tgbotapi.Update
	channelOutput chan OutputAnswer

	ready bool // для проверки заполнены ли все поля структуры
}

type OutputAnswer struct {
	ChatID  int64
	Text    string
	PhotoID string
}

var commands = map[string]func(*User) OutputAnswer{
	"let`s go": letsGo,
	"❤️":       like,
	"🥴":        unLike,
	"😴":        sleep,
}

func letsGo(user *User) OutputAnswer {
	var introUser User
	if user.Gender == "муж" {

		pipeline := []bson.M{bson.M{"$match": bson.M{"gender": "жен"}}, bson.M{"$sample": bson.M{"size": 1}}}
		cursor, err := cache.Collection.Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Logger.Error.Println("не удалось вытащить мужчину из cache: ", err)
		}

		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			log.Logger.Error.Println("не считалось с курсора: ", err)
		}
		for _, result := range results {
			introUser.ChatID = result["chat_id"].(int64)
			introUser.Name = fmt.Sprintf("%v", result["name"])
			introUser.Age = int(result["age"].(int32))
			introUser.Bio = fmt.Sprintf("%v", result["bio"])
			introUser.PhotoID = fmt.Sprintf("%v", result["photo_id"])
		}
	} else {
		pipeline := []bson.M{bson.M{"$match": bson.M{"gender": "муж"}}, bson.M{"$sample": bson.M{"size": 1}}}
		cursor, err := cache.Collection.Aggregate(context.Background(), pipeline)
		if err != nil {
			log.Logger.Error.Println("не удалось вытащить женщину из cache: ", err)
		}

		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			log.Logger.Error.Println("не счиалось с курсора: ", err)
		}
		for _, result := range results {
			introUser.ChatID = result["chat_id"].(int64)
			introUser.Name = fmt.Sprintf("%v", result["name"])
			introUser.Age = int(result["age"].(int32))
			introUser.Bio = fmt.Sprintf("%v", result["bio"])
			introUser.PhotoID = fmt.Sprintf("%v", result["photo_id"])
		}
	}
	updateUserInCache(user, introUser.ChatID)
	text := introUser.Name + ", " + strconv.Itoa(introUser.Age) + " лет. " + introUser.Bio
	answer := OutputAnswer{
		ChatID:  user.ChatID,
		Text:    text,
		PhotoID: introUser.PhotoID,
	}
	return answer
}

func like(u *User) OutputAnswer {
	var answer OutputAnswer

	if _, err := bot.Send(tgbotapi.NewPhotoShare(u.LastIntro, u.PhotoID)); err != nil {
		log.Logger.Error.Printf("не удалось отправить фотографию пользователю ", u.LastIntro, " : ", err)
	}
	if _, err := bot.Send(tgbotapi.NewMessage(u.LastIntro, fmt.Sprintf("Вы понравились❤️ @"+u.UserName+"\n"+
		u.Name+", "+strconv.Itoa(u.Age)+". "+u.Bio))); err != nil {
		log.Logger.Error.Println("не удалось отправить сообщение lastIntro: ", err)
	}

	answer = letsGo(u)
	answer.Text = "Ваш лайк отправлен, давай посмотрим на следующего:\n " + answer.Text
	return answer
}

func sleep(user *User) OutputAnswer {
	var answer OutputAnswer
	answer = OutputAnswer{
		ChatID: user.ChatID,
		Text: `посидим подождём...

Вот твоя анкета: ` + user.Name + ", " + strconv.Itoa(user.Age) + ". " + user.Bio,
		PhotoID: user.PhotoID,
	}
	return answer
}

func unLike(user *User) OutputAnswer {
	var answer OutputAnswer
	answer = letsGo(user)
	return answer
}

func (u *User) HandlerCommand() {
	command := <-u.channelInput
	answer := commands[command.Message.Text](u)
	u.channelOutput <- answer
}

func (u *User) ChannelInputHandler(update *tgbotapi.Update) {
	u.channelInput <- update
}

func (u *User) GetChannelOutput() chan OutputAnswer {
	return u.channelOutput
}

func updateUserInCache(user *User, introChatID int64) {
	filter := bson.M{"user_id": user.UserID}
	update := bson.M{"$set": bson.M{"last_intro": introChatID}}

	if _, err := cache.Collection.UpdateOne(context.Background(), filter, update); err != nil {
		log.Logger.Error.Println("не удалось изменить lastIntro: ", err)
	}
}

func setUserInCache(user *User) error {
	if _, err := cache.Collection.InsertOne(context.Background(), user); err != nil {
		return err
	}
	log.Logger.Process.Println("пользователь записан в mongoDB")
	return nil
}

func GetUserInCache(userID int) (*User, error) {

	filter := bson.M{"user_id": userID}

	var user User

	if err := cache.Collection.FindOne(context.Background(), filter).Decode(&user); err != nil {

		log.Logger.Error.Println("пользователь не считан из cache: ", err)
		return nil, err
	}
	log.Logger.Process.Println("пользователь считан из cache ")

	return &user, nil
}

func GetUserInCacheMulti(filter interface{}) ([]*User, error) {
	var users []*User

	cur, err := cache.Collection.Find(context.Background(), filter)
	if err != nil {
		return users, err
	}

	for cur.Next(context.Background()) {
		var u User
		if err := cur.Decode(&u); err != nil {
			return users, err
		}
		users = append(users, &u)
	}

	if err := cur.Err(); err != nil {
		return users, err
	}
	if err := cur.Close(context.Background()); err != nil {
		return users, err
	}
	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func DeleteUserInCache(userID int) error {
	filter := bson.M{"user_id": userID}

	res, err := cache.Collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("no user were deleted")
	}
	return nil
}

func AddUserInDataBase(u *User) error {

	q := "INSERT INTO users (user_id, chat_id, name, gender, age, bio, file_id, user_name) VALUES ($1,$2, $3, $4, $5, $6, $7, $8)"

	if _, err := psql.DB.Exec(context.Background(), q, u.UserID, u.ChatID, u.Name, u.Gender, u.Age, u.Bio, u.PhotoID, u.UserName); err != nil {
		e := fmt.Sprintf("не удалось установить значение в dataBase: %v", err)
		return errors.New(e)
	}
	log.Logger.Process.Println("В DataBase установлен новый поользователь")
	return nil
}

func CheckUserInBase(userID int) (*User, bool) {
	q := "SELECT user_id, chat_id, name, gender, age, bio, file_id, user_name FROM users where user_id = $1"

	var user User

	if err := psql.DB.QueryRow(context.Background(), q, userID).Scan(
		&user.UserID,
		&user.ChatID,
		&user.Name,
		&user.Gender,
		&user.Age,
		&user.Bio,
		&user.PhotoID,
		&user.UserName,
	); err != nil {
		log.Logger.Error.Println("не удалось считать пользователя: ", err)
		return nil, false
	}

	return &user, true
}
