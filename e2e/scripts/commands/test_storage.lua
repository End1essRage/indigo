-- Комплексный тест хранилища
local chat_id = ctx.chat_id
local test_results = {}

log("Starting storage tests...")

-- Тест 1: Создание документа
local function test_create()
    local test_data = {
        name = "Test User",
        score = 100,
        tags = {"test", "demo"},
        created_at = os.time()
    }
    
    local ok, id = storage_create("test_collection", test_data)
    if ok then
        log("created")

        table.insert(test_results, "✅ CREATE: Документ создан с ID: " .. id)
        return id
    else
        log("not created")

        table.insert(test_results, "❌ CREATE: Ошибка создания")
        return nil
    end
end

-- Тест 2: Получение по ID
local function test_get_by_id(id)
    local data, err = storage_get_by_id("test_collection", id)
    log(data)
    log(err)
    if err then
        log(test_results, "❌ GET BY ID: Ошибка: " .. err)
        table.insert(test_results, "❌ GET BY ID: Ошибка: " .. err)
    elseif data then
        log(test_results, "✅ GET BY ID: Найден документ: " .. data.name)
        table.insert(test_results, "✅ GET BY ID: Найден документ: " .. data.name)
    else
        log(test_results, "❌ GET BY ID: Документ не найден")
        table.insert(test_results, "❌ GET BY ID: Документ не найден")
    end
end

-- Тест 3: Поиск с условиями
local function test_query()
    -- Простое условие
    local simple_query = query_condition("score", ">", 50)
    local results = storage_get("test_collection", 10, simple_query)
    table.insert(test_results, "✅ SIMPLE QUERY: Найдено документов: " .. #results)
    
    -- Сложное условие с AND
    local complex_query = query_and(
        query_condition("score", ">=", 100),
        query_condition("name", "=", "Test User")
    )
    results = storage_get("test_collection", 10, complex_query)
    table.insert(test_results, "✅ COMPLEX AND QUERY: Найдено: " .. #results)
    
    -- Условие с OR
    local or_query = query_or(
        query_condition("score", "<", 50),
        query_condition("score", ">", 150)
    )
    results = storage_get("test_collection", 10, or_query)
    table.insert(test_results, "✅ OR QUERY: Найдено: " .. #results)
end

-- Тест 4: Обновление
local function test_update(id)
    local update_data = {
        score = 150,
        updated_at = os.time()
    }
    
    storage_update_by_id("test_collection", id, update_data)
    
    local updated = storage_get_by_id("test_collection", id)
    if updated and updated.score == 150 then
        table.insert(test_results, "✅ UPDATE: Документ обновлен")
    else
        table.insert(test_results, "❌ UPDATE: Ошибка обновления")
    end
end

-- Тест 5: Массовое обновление
local function test_bulk_update()
    local query = query_condition("score", "<", 200)
    local update_data = {
        category = "updated",
        bulk_updated = true
    }
    
    storage_update("test_collection", query, update_data)
    table.insert(test_results, "✅ BULK UPDATE: Выполнено")
end

-- Тест 6: Получение только ID
local function test_get_ids()
    local query = query_condition("score", ">", 0)
    local ids = storage_get_ids("test_collection", 5, query)
    table.insert(test_results, "✅ GET IDS: Получено ID: " .. #ids)
end

-- Тест 7: Удаление
local function test_delete(id)
    local ok, err = storage_delete_by_id("test_collection", id)
    if not ok then
        log("error while deleting" .. err)
    end

    local deleted, err = storage_get_by_id("test_collection", id)
    if err then 
        table.insert(test_results, "❌ DELETE: Ошибка удаления" .. err)
    else
        if not deleted then
            table.insert(test_results, "✅ DELETE: Документ удален")
        else
            table.insert(test_results, "❌ DELETE: Ошибка удаления не удалилось")
        end
    end
end

-- Выполняем тесты
local test_id = test_create()
if test_id then
    test_get_by_id(test_id)
    test_query()
    test_update(test_id)
    test_bulk_update()
    test_get_ids()
    test_delete(test_id)
end

-- Отправляем результаты
local message = "📊 *Результаты тестов Storage:*\n\n" .. table.concat(test_results, "\n")
send_message(chat_id, message)