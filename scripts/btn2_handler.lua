local function handle()
    log("Скрипт запущен!btn2_handler.lua User ID: " .. ctx.user.id)

    cache_set("user:1", [[{"name":"abb","age":15}]])
end

handle()