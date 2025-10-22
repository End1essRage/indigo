-- Сохранение профиля из формы
local chat_id = ctx.chat_id
local user_id = tostring(ctx.user_id)
local form_data = ctx.form_data

-- Получаем существующий профиль
local profile, err = storage_get_one("profiles", query_condition("user_id", "=", user_id))

local profile_data = {
    user_id = user_id,
    name = form_data.name,
    age = tonumber(form_data.age),
    city = form_data.city,
    bio = form_data.bio or "",
    updated_at = os.time()
}

if err then
    send(chat_id, "❌ Ошибка при поиске прфоиля")
else
    if profile then
        -- Обновляем существующий
        storage_update_by_id("profiles", profile.id, profile_data)
        send(chat_id, "✅ Профиль обновлен!")
    else
         -- Создаем новый
        profile_data.created_at = os.time()
        local ok, id = storage_create("profiles", profile_data)
        if ok then
            send(chat_id, "✅ Профиль создан!")
        else
            send(chat_id, "❌ Ошибка при сохранении профиля")
        end
    end
end

-- Кешируем профиль
cache_set("profile:" .. user_id, json_encode(profile_data))