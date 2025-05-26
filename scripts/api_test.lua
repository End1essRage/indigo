local function handle()
    log("Скрипт запущен! api_test")
    if ctx.req_data ~= nil then
        log("data not null")
        log(ctx.req_data)
        log("data is " .. ctx.req_data.user_id)
        
    end

    
end

handle()