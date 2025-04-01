local function handle()
    log("Скрипт запущен! User ID: " .. ctx.user.id)
    local mesh = {
        Rows = {
            {
                {Text = "Btn1", Script = "btn1.lua"},
                {Text = "Btn2", Data = "custom123"}
            }
        }
    }

    send_keyboard(ctx.chat_id, "Menu", mesh)
end

handle()