package main

import (
	log "Davinchik/pkg/logger"
	"Davinchik/telegram"
)

func main() {
	log.Logger.Process.Println("запуск программы")

	log.Logger.Process.Println("попытка запустить бота")
	telegram.StartBot()
}
