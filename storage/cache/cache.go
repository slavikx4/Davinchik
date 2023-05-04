package cache

import (
	log "Davinchik/pkg/logger"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	Collection *mongo.Collection
)

func init() {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Logger.Error.Fatalln("не удалось создать клиента для подключения к mongodb: ", err)
	}
	Collection = client.Database("BASE").Collection("users")
	if err := client.Ping(context.Background(), nil); err != nil {
		panic(err)
	} else {
		log.Logger.Process.Println("успешное подключение к mongodb")
	}
	Collection.DeleteMany(context.Background(), bson.M{})
}
