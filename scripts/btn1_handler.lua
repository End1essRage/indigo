local function handle()
    log("Скрипт запущен!btn1_handler.lua User ID: " .. ctx.user.id)

    -- Сохранение данных в кэш
    cache_set("user:123", [[{"name":"John","age":30}]])

    -- Получение данных из кэша
    local user_data = cache_get("user:123")
    if user_data ~= nil then
        log("User data: " .. user_data)
        send_message(ctx.chat_id, "Button ndler test, " .. user_data)
    end
    
end

handle()