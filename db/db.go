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
	// Здесь указываем правильные данные для подключения с авторизацией
	Conn, err = tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		User: "admin",  // Убедитесь, что имя пользователя верное
		Pass: "anksoonamoon",  // Убедитесь, что пароль правильный
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к Tarantool: %v", err)
	} else {
		log.Println("Подключение к Tarantool успешно!")
	}
}