local function handle()
    log("Скрипт запущен! User ID: " .. ctx.user.id)

    send_message(ctx.chat_id, "Hello, " .. ctx.user.from_name)
end

handle()