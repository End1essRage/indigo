local function handle()
    log("Скрипт запущен!btn3_handler.lua User ID: " .. ctx.user.id)

    -- Получение данных из кэша
    local user_data = cache_get("user:1")
    if user_data ~= nil then
        log("User data: " .. user_data)
        send_message(ctx.chat_id, "Button ndler test, " .. user_data)
    end
end

handle()