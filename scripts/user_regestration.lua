local function handle()
    local user_id = ctx.user.id
    local form_data = ctx.form_data
    
    -- Проверяем полученные данные
    if not form_data then
        log("Ошибка: данные формы не получены")
        send_message(user_id, "❌ Произошла ошибка при обработке формы")
        return
    end
    
    -- Логируем полученные данные
    log("Данные формы пользователя "..user_id..":")
    for k, v in pairs(form_data) do
        log(" - "..k..": "..tostring(v))
    end
    
    -- Форматируем данные для вывода
    local message = string.format(
        "✅ Регистрация завершена!\n\n"..
        "Ваши данные:\n"..
        "👤 Имя: %s\n"..
        "🎂 Возрастная группа: %s\n"..
        "📧 Email: %s\n"..
        "📢 Рассылка: %s",
        form_data.user_name,
        form_data.user_age,
        form_data.user_email,
        form_data.newsletter == "yes" and "подписаны" or "не подписаны"
    )
    
    -- Сохраняем в хранилище
    local success, err = storage_save("user_registrations", tostring(user_id), form_data)
    if not success then
        log("Ошибка сохранения: "..err)
        send_message(user_id, "❌ Ошибка при сохранении данных")
        return
    end
    
    -- Отправляем результат пользователю
    send_message(user_id, message)
    
    -- Дополнительные действия
    if form_data.newsletter == "yes" then
        send_message(user_id, "📬 Вы подписаны на нашу рассылку!")
    end
    
    log("Форма пользователя "..user_id.." успешно обработана")
end

handle()