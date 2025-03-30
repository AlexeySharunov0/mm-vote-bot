package db

import "log"

// Функция для создания пространства для голосований и голосов участников
func CreatePollSchema() {
	_, err := Conn.Eval(`
		box.cfg{}

		-- Создаем пространство для голосований (polls)
		polls = box.schema.space.create("polls", { if_not_exists = true })
		polls:format({
			{name = "id", type = "unsigned"},        -- Уникальный идентификатор голосования
			{name = "creator_id", type = "unsigned"},-- ID пользователя, который создал голосование
			{name = "question", type = "string"},    -- Текст вопроса голосования
			{name = "options", type = "array"},      -- Массив вариантов ответов
			{name = "created_at", type = "unsigned"} -- Время создания голосования
		})
		-- Создаем уникальный индекс по полю id для голосований
		polls:create_index("primary", {parts = {1, "unsigned"}, if_not_exists = true})

		-- Создаем пространство для голосов (votes)
		votes = box.schema.space.create("votes", { if_not_exists = true })
		votes:format({
			{name = "poll_id", type = "unsigned"},  -- ID голосования
			{name = "user_id", type = "unsigned"},  -- ID пользователя
			{name = "option_index", type = "unsigned"} -- Индекс выбранного варианта
		})
		-- Создаем составной индекс по полям poll_id и user_id для предотвращения повторного голосования
		votes:create_index("primary", {parts = {1, "unsigned", 2, "unsigned"}, if_not_exists = true})
	`, []interface{}{})

	if err != nil {
		log.Fatalf("Ошибка создания таблиц: %v", err)
	}
}