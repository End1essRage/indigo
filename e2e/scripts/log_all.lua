-- Логирование всех действий
local action_type = ctx.type -- "command", "text", "button", etc.
local user_id = ctx.user_id
local chat_id = ctx.chat_id

local log_entry = {
    type = action_type,
    user_id = user_id,
    chat_id = chat_id,
    timestamp = os.time(),
    data = ctx.data or ctx.text or ctx.command
}

-- Сохраняем в хранилище
storage_create("activity_log", log_entry)

-- Увеличиваем счетчик в кеше
local counter_key = "stats:actions:" .. action_type
local current = cache_get(counter_key)
if current then
    cache_set(counter_key, tostring(tonumber(current) + 1))
else
    cache_set(counter_key, "1")
end

log("Action logged: " .. action_type .. " from user " .. user_id)