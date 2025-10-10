-- Поиск пользователей по имени
local chat_id = ctx.chat_id

send_message(chat_id, "Введите имя для поиска:")

-- Сохраняем состояние в кеш
cache_set("search:mode:" .. chat_id, "name")
cache_set("search:active:" .. chat_id, "true")