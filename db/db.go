package db

import (
	"log"

	"github.com/tarantool/go-tarantool"
)

// Переменная для хранения подключения к Tarantool
var Conn *tarantool.Connection

// Функция для подключения к Tarantool
func InitDB() {
	var err error
	// Здесь указывается IP и порт, где работает твой Tarantool
	Conn, err = tarantool.Connect("127.0.0.1:3301", tarantool.Opts{})
	if err != nil {
		log.Fatalf("Ошибка подключения к Tarantool: %v", err)
	} else {
		log.Println("Подключение к Tarantool успешно!")
	}
}