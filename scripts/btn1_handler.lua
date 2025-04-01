local function handle()
    log("Скрипт запущен!btn1_handler.lua User ID: " .. ctx.user.id)

    send_message(ctx.chat_id, "Button ndler test, " .. ctx.user.from_name)
    send_message(ctx.chat_id, "Button handler test cb data is , " .. ctx.cb_data)
end

handle()