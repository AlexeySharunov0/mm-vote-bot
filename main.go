package main

import (
	"log"
	"os"
	"time"
	//"strings"

	"github.com/tarantool/go-tarantool"
)

func init() {
	// Открытие или создание файла логов
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла для логов: %v", err)
}
	// Устанавливаем логгер, чтобы выводить логи в файл и консоль
	log.SetOutput(logFile)	
	// Настроим формат вывода лога
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func connectToDB() (*tarantool.Connection, error) {
	log.Println("Подключение к базе данных Tarantool...")

	// Настройка подключения с параметрами
	opts := tarantool.Opts{
		User: "admin", // Указываем пользователя admin, так как guest не имеет прав для создания пространства
		Pass: "anksoonamoon", // Указываем пароль для пользователя admin
	}

	// Подключение к серверу tarantool с параметрами
	conn, err := tarantool.Connect("localhost:3301", opts)
	if err != nil {
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return nil, err
	}
	log.Println("Успешно подключено к базе данных Tarantool")
	return conn, nil
}

func createPollsSpace(conn *tarantool.Connection) error {
    log.Println("Создание пространства для голосований...")

    // Создание пространства "polls" с использованием Eval
    _, err := conn.Eval(`
        if box.space.polls == nil then
            box.schema.space.create('polls', {if_not_exists = true})
        end
    `, []interface{}{})
    if err != nil {
        log.Printf("Ошибка при создании пространства для голосований: %v", err)
        return err
    }

    // Создание индекса для поля id
    _, err = conn.Eval(`
        if box.space.polls.index.primary == nil then
            box.space.polls:create_index('primary', {type = 'TREE', parts = {1}})
        end
    `, []interface{}{})
    if err != nil {
        log.Printf("Ошибка при создании индекса для голосований: %v", err)
        return err
    }

    log.Println("Пространство для голосований успешно создано или уже существует")
    return nil
}

func createVotesSpace(conn *tarantool.Connection) error {
    log.Println("Создание пространства для голосов участников...")

    // Создание пространства "votes" с использованием Eval
    _, err := conn.Eval(`
        if box.space.votes == nil then
            box.schema.space.create('votes', {if_not_exists = true})
        end
    `, []interface{}{})
    if err != nil {
        log.Printf("Ошибка при создании пространства для голосов: %v", err)
        return err
    }

    // Создание составного индекса для полей poll_id и user_id
    _, err = conn.Eval(`
        if box.space.votes.index.primary == nil then
            box.space.votes:create_index('primary', {type = 'TREE', parts = {1, 2}})
        end
    `, []interface{}{})
    if err != nil {
        log.Printf("Ошибка при создании индекса для голосов: %v", err)
        return err
    }

    log.Println("Пространство для голосов успешно создано или уже существует")
    return nil
}


func insertPoll(conn *tarantool.Connection, creatorID int, question string, options []string) error {
	log.Println("Вставка нового голосования...")

	// Вставка записи в пространство "polls"
	poll := []interface{}{creatorID, question, options, time.Now().Format(time.RFC3339)}
	_, err := conn.Call("box.space.polls:insert", []interface{}{poll})
	if err != nil {
		log.Printf("Ошибка при вставке голосования: %v", err)
		return err
	}

	log.Println("Голосование успешно добавлено")
	return nil
}

func insertVote(conn *tarantool.Connection, pollID int, userID int, optionIndex int) error {
	log.Println("Вставка голоса участника...")

	// Вставка записи в пространство "votes"
	vote := []interface{}{pollID, userID, optionIndex}
	_, err := conn.Call("box.space.votes:insert", []interface{}{vote})
	if err != nil {
		log.Printf("Ошибка при вставке голоса: %v", err)
		return err
	}

	log.Println("Голос успешно добавлен")
	return nil
}

func main() {
	// Подключение к базе данных
	conn, err := connectToDB()
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных: ", err)
	}
	defer conn.Close()

	// Создание пространств (spaces)
	err = createPollsSpace(conn)
	if err != nil {
		log.Fatal("Не удалось создать пространство для голосований: ", err)
	}

	err = createVotesSpace(conn)
	if err != nil {
		log.Fatal("Не удалось создать пространство для голосов участников: ", err)
	}

	// Вставка нового голосования
	err = insertPoll(conn, 1, "Какой язык программирования лучший?", []string{"Go", "Python", "Java", "C++"})
	if err != nil {
		log.Fatal("Не удалось вставить новое голосование: ", err)
	}

	// Вставка голосов
	err = insertVote(conn, 1, 101, 0) // Пользователь 101 голосует за "Go"
	if err != nil {
		log.Fatal("Не удалось вставить голос: ", err)
	}

	err = insertVote(conn, 1, 102, 1) // Пользователь 102 голосует за "Python"
	if err != nil {
		log.Fatal("Не удалось вставить голос: ", err)
	}

	// Завершение работы программы
	log.Println("Программа завершена успешно")
}
