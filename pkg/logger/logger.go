package logger

import (
	"log"
	"os"
)

// Logger сам логер, который используется во всём проекте
var (
	Logger *logger
)

func init() {

	//выдаём ключи
	flagsFile := os.O_APPEND | os.O_CREATE
	flagsLog := log.Ldate | log.Ltime | log.Lshortfile

	//создаём дириккторию для хранения файлов
	if err := os.Mkdir("history-loggers", 0777); err != nil {
		log.Fatalln(err)
	}

	//открываем файлы или создаём, если такого нет
	fileProcess, err := os.OpenFile("history-loggers/log_process.txt", flagsFile, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	fileError, err := os.OpenFile("history-loggers/log_error.txt", flagsFile, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	fileChat, err := os.OpenFile("history-loggers/log_chat.txt", flagsFile, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	//создаём логгеры
	loggerInfo := log.New(fileProcess, "PROCESS:	", flagsLog)
	loggerError := log.New(fileError, "ERROR:	", flagsLog)
	loggerChat := log.New(fileChat, "CHAT:	", flagsLog)

	//инициализируем наш логер
	Logger = &logger{
		Process: loggerInfo,
		Error:   loggerError,
		Chat:    loggerChat,
	}
}

// структура объединения логеров
type logger struct {
	Process *log.Logger
	Error   *log.Logger
	Chat    *log.Logger
}
