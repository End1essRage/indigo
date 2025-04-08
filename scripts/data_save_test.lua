local function handle()

-- Сохранение формы
local form_data = {
    name = "Василий",
    age = 30,
    contacts = {
        {type = "email", value = "test@example.com"},
        {type = "phone", value = "+79991234567"}
    },
    preferences = {
        newsletter = true,
        theme = "dark"
    }
}

local success, err = storage_save("user_forms", "form_123", form_data)
if not success then
    log("Ошибка сохранения: " .. err)
end

-- Загрузка данных
local loaded_data, err = storage_load("user_forms", "form_123")
if loaded_data then
    log("Загруженные данные:")
    for k,v in pairs(loaded_data) do
        log(k .. " = " .. tostring(v))
    end
end
end

handle()
