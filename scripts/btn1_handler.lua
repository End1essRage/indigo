local function handle()
    log("Скрипт запущен!btn1_handler.lua User ID: " .. ctx.user.id)

    send_message(ctx.chat_id, "Button 1 handler test, " .. ctx.user.from_name)
end

handle()