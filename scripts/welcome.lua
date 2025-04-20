local function handle()
    log("Скрипт запущен! User ID: " .. ctx.user.id)

    log("Secret: " .. reveal("SOME_SECRET"))

    send_message(ctx.chat_id, "Hello, " .. ctx.user.name)
end

handle()