-- Тестирование кеша
local chat_id = ctx.chat_id
local results = {}

-- Тест 1: Установка значения
cache_set("test:key1", "value1")
results[1] = "✅ SET: Значение установлено"

-- Тест 2: Получение значения
local value = cache_get("test:key1")
if value == "value1" then
    results[2] = "✅ GET: Значение получено: " .. value
else
    results[2] = "❌ GET: Ошибка получения"
end

-- Тест 3: Сложные данные (JSON)
local complex_data = {
    user = "test",
    score = 100,
    tags = {"a", "b", "c"}
}
local json_data = json_encode(complex_data)
cache_set("test:complex", json_data)
local item = cache_get("test:complex")
local itemTbl = json_decode(item)
if itemTbl then 
    if itemTbl.user == "test" then
        results[3] = "✅ SET JSON: Сложные данные сохранены, получены и декодирвоаны"
    else
        results[3] = "❌ SET JSON: Сложные данные не получены или не декодированы"
    end    
else
    results[3] = "❌ SET JSON: Сложные данные не получены или не декодированы"
end

-- Тест 4: Получение несуществующего ключа
local missing = cache_get("test:missing")
if not missing then
    results[4] = "✅ GET MISSING: Корректно вернул nil"
else
    results[4] = "❌ GET MISSING: Неожиданное значение"
end

-- Тест 5: Перезапись значения
cache_set("test:key1", "new_value")
local new_value = cache_get("test:key1")
if new_value == "new_value" then
    results[5] = "✅ OVERWRITE: Значение перезаписано"
else
    results[5] = "❌ OVERWRITE: Ошибка перезаписи"
end

local message = "💾 *Результаты тестов Cache:*\n\n" .. table.concat(results, "\n")
send(chat_id, message, nil)