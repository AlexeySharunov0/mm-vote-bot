-- Это базовая конфигурация для инициализации Tarantool
box.cfg {
    listen = 3301  -- Указываем порт для прослушивания
}

-- Пример создания пространства "polls"
polls = box.schema.space.create('polls', {
    if_not_exists = true,
    temporary = false,
})

polls:format({
    {name = 'id', type = 'unsigned'},
    {name = 'creator_id', type = 'unsigned'},
    {name = 'question', type = 'string'},
    {name = 'options', type = 'array'},
    {name = 'created_at', type = 'unsigned'},
})

-- Создание индекса для "polls"
polls:create_index('primary', {
    type = 'tree',
    parts = {'id'},
    if_not_exists = true
})


