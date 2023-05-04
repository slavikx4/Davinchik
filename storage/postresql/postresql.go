package postresql

import (
	log "Davinchik/pkg/logger"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	DB *pgxpool.Pool
)

func init() {
	url := "YOUR-POSTGRES-URL"
	var err error
	DB, err = pgxpool.New(context.Background(), url)
	if err != nil {
		log.Logger.Error.Fatalln("не удалось подключиться к DataBASE: ", err)
	}
	if err := DB.Ping(context.Background()); err != nil {
		panic(err)
	}
	log.Logger.Process.Println("postgres успешно подключён")
}
