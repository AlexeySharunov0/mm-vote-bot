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

    // Создание пространства "polls" с автоинкрементным id
    _, err := conn.Eval(`
        if not box.space.polls then
            local polls = box.schema.space.create('polls', {if_not_exists = true})
            polls:format({
                {name = 'id', type = 'unsigned'},          -- Уникальный ID голосования
                {name = 'creator_id', type = 'unsigned'},  -- ID создателя голосования
                {name = 'question', type = 'string'},      -- Вопрос голосования
                {name = 'options', type = 'array'},        -- Массив вариантов ответов
                {name = 'created_at', type = 'string'},    -- Дата создания голосования
                {name = 'is_active', type = 'boolean'}     -- Флаг активности голосования
            })
            polls:create_index('primary', {type = 'TREE', parts = {1, 'unsigned'}, if_not_exists = true})

            -- Создаем sequence для автоинкремента ID голосований
            if not box.sequence.poll_id_seq then
                box.schema.sequence.create('poll_id_seq', {if_not_exists = true})
            end
            
            polls:create_index('id', {parts = {1, 'unsigned'}, sequence = 'poll_id_seq', if_not_exists = true})
        end
    `, []interface{}{})

    if err != nil {
        log.Printf("Ошибка при создании пространства для голосований: %v", err)
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

    // Подготовка данных для вставки
    poll := []interface{}{
        uint(creatorID),                         // Преобразуем creatorID к unsigned
        question,                                // Текст вопроса
        options,                                 // Массив вариантов ответов
        time.Now().Format(time.RFC3339),         // Дата создания
        true,                                    // Флаг is_active
    }

    // Используем Eval с корректным форматом передачи данных
    _, err := conn.Eval(`
        return box.space.polls:auto_increment({...})
    `, []interface{}{poll})

    if err != nil {
        log.Printf("Ошибка при вставке голосования: %v", err)
        return err
    }

    log.Println("Голосование успешно добавлено")
    return nil
}

func getPollResults(conn *tarantool.Connection, pollID int) error {
    log.Printf("Получение результатов голосования с ID %d...\n", pollID)

    // Запрос всех голосов для данного голосования
    resp, err := conn.Select("votes", "primary", 0, 1000, tarantool.IterEq, []interface{}{pollID})
    if err != nil {
        log.Printf("Ошибка при получении голосов: %v", err)
        return err
    }

    if len(resp.Data) == 0 {
        log.Println("Голосов для этого голосования не найдено")
        return nil
    }

    // Подсчет результатов
    results := make(map[int]int) // Карта для хранения количества голосов по индексу варианта
    for _, row := range resp.Data {
        vote := row.([]interface{})
        optionIndex := int(vote[2].(uint64)) // Индекс варианта ответа
        results[optionIndex]++
    }

    // Вывод результатов
    log.Println("Результаты голосования:")
    for optionIndex, count := range results {
        log.Printf("Вариант %d: %d голосов\n", optionIndex, count)
    }

    return nil
}

func insertVote(conn *tarantool.Connection, pollID int, userID int, optionIndex int) error {
    log.Println("Вставка голоса участника...")

    // Проверяем, существует ли уже голос для этого пользователя и голосования
    resp, err := conn.Select("votes", "primary", 0, 1, tarantool.IterEq, []interface{}{pollID, userID})
    if err != nil {
        log.Printf("Ошибка при проверке существования голоса: %v", err)
        return err
    }

    if len(resp.Data) > 0 {
        log.Println("Голос уже существует для этого пользователя и голосования")
        return nil // Пропускаем вставку
    }

    // Вставка записи в пространство "votes"
    vote := []interface{}{
        uint(pollID),       // Преобразуем pollID к unsigned
        uint(userID),       // Преобразуем userID к unsigned
        optionIndex,        // Индекс выбранного варианта
    }
    _, err = conn.Call("box.space.votes:insert", []interface{}{vote})
    if err != nil {
        log.Printf("Ошибка при вставке голоса: %v", err)
        return err
    }

    log.Println("Голос успешно добавлен")
    return nil
}

func endPoll(conn *tarantool.Connection, pollID int) error {
    log.Printf("Завершение голосования с ID %d...\n", pollID)

    // Обновление поля is_active на false
    _, err := conn.Update("polls", "primary", []interface{}{uint(pollID)}, []interface{}{
        []interface{}{"=", uint(5), false}, // Поле is_active находится на 5-й позиции (индексация с 0)
    })
    if err != nil {
        log.Printf("Ошибка при завершении голосования: %v", err)
        return err
    }

    log.Println("Голосование успешно завершено")
    return nil
}

func deletePoll(conn *tarantool.Connection, pollID int) error {
    log.Printf("Удаление голосования с ID %d...\n", pollID)

    // Поиск всех голосов для данного голосования
    resp, err := conn.Select("votes", "primary", 0, 1000, tarantool.IterEq, []interface{}{uint(pollID)})
    if err != nil {
        log.Printf("Ошибка при поиске голосов: %v", err)
        return err
    }

    if len(resp.Data) == 0 {
        log.Println("Голосов для этого голосования не найдено")
    } else {
        // Удаление каждого голоса
        for _, row := range resp.Data {
            vote := row.([]interface{})
            userKey := vote[1].(uint64) // Получаем user_id из записи

            // Удаляем запись по составному ключу [pollID, userKey]
            _, err := conn.Delete("votes", "primary", []interface{}{uint(pollID), uint(userKey)})
            if err != nil {
                log.Printf("Ошибка при удалении голоса для пользователя %d: %v", userKey, err)
                return err
            }
        }
        log.Println("Все голоса успешно удалены")
    }

    // Удаление записи из пространства polls
    _, err = conn.Delete("polls", "primary", []interface{}{uint(pollID)})
    if err != nil {
        log.Printf("Ошибка при удалении голосования: %v", err)
        return err
    }

    log.Println("Голосование и связанные голоса успешно удалены")
    return nil
}

func main() {
    // Подключение к базе данных
    conn, err := connectToDB()
    if err != nil {
        log.Fatal("Не удалось подключиться к базе данных: ", err)
    }
    defer conn.Close()

    // Создание пространств
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

    // Просмотр результатов
    err = getPollResults(conn, 1)
    if err != nil {
        log.Fatal("Не удалось получить результаты голосования: ", err)
    }

    // Завершение голосования
    err = endPoll(conn, 1)
    if err != nil {
        log.Fatal("Не удалось завершить голосование: ", err)
    }

    // Удаление голосования
    err = deletePoll(conn, 1)
    if err != nil {
        log.Fatal("Не удалось удалить голосование: ", err)
    }

    log.Println("Программа завершена успешно")
}
