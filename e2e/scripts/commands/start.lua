-- Инициализация пользователя
local chat_id = ctx.chat_id
local user_id = ctx.user.id
local username = ctx.user.name or "anonymous"

-- Создаем нового пользователя
local user_data = {
    user_id = user_id,
    chat_id = chat_id,
    username = username,
    created_at = os.time(),
    is_active = true,
    commands_count = 0
}

local ok, id = storage_create("users", user_data)

if ok then
    log("New user created with id: " .. id)
    send_message(chat_id, "👋 Добро пожаловать! Ваш профиль создан.")
    
    -- Кешируем ID пользователя
    cache_set("user:" .. user_id, id)
else
    send_message(chat_id, "❌ Ошибка при создании профиля")
end

-- Показываем главное меню
local mesh = {
    Rows = {
        {
            {Text = "📊 Тесты", Data = "show_tests"},
            {Text = "👤 Профиль", Data = "show_profile"}
        }
    }
}

send_keyboard(chat_id, "Главное меню:", mesh)
