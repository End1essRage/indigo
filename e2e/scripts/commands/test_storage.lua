-- Комплексный тест хранилища
local chat_id = ctx.chat_id
local test_results = {}

log("Starting storage tests...")

-- Тест 1: Создание документа
local function test_create()
    local test_data = {
        name = "Test User",
        score = 99,
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
    
    local count, err = storage_update("test_collection", query, update_data)
    local should = 5
    if err then 
        table.insert(test_results, "❌ BULK UPDATE: ошибка" .. err)
    else 
        if count == should then
            table.insert(test_results, "✅ BULK UPDATE: выполнено")
        else 
            table.insert(test_results, "❌ BULK UPDATE: неправильное кол-во обновившехся")
        end
    end
   
end

-- Тест 6: Получение только ID
local function test_get_ids()
    local query = query_condition("score", ">", 0)
    local ids = storage_get_ids("test_collection", 10, query)
    local should = 5
    if #ids ~= should then 
        table.insert(test_results, "❌ GET IDS: Получено ID: " .. #ids .. " Должно: " .. should)
    else
        table.insert(test_results, "✅ GET IDS: Получено ID: " .. #ids)
    end
    
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

local function seed_data()
    local count, err = storage_delete("test_collection", query_condition("_id", "!=", "a"))
    if err then
        table.insert(test_results, "❌ CREATE: ошибка удаления старых данных" .. err)
    end

    local test_data = {
        {
            name = "Test User1",
            score = 50,
            city = "KAZ",
            tags = {"test", "demo"},
            created_at = os.time()
        },
        {
            name = "Test User2",
            score = 99,
            city = "KAZ",
            tags = {"test", "demo"},
            created_at = os.time()
        },
        {
            name = "Test User3",
            score = 101,
            city = "MOS",
            tags = {"test", "demo"},
            created_at = os.time()
        },
        {
            name = "Test User4",
            score = 100,
            city = "MOS",
            tags = {"test", "demo"},
            created_at = os.time()
        },
        {
            name = "Test User5",
            score = 49,
            city = "MOS",
            tags = {"test", "demo"},
            created_at = os.time()
        },
    }
    
    local ok1, id1 = storage_create("test_collection", test_data[1])
    local ok2, id2 = storage_create("test_collection", test_data[2])
    local ok3, id3 = storage_create("test_collection", test_data[3])
    local ok4, id4 = storage_create("test_collection", test_data[4])
    local ok5, id5 = storage_create("test_collection", test_data[5])

    if ok1 and ok2 and ok3 and ok4 and ok5 then
        log("created")

        table.insert(test_results, "✅ CREATE: Документы созданы с ID")
        return id
    else
        log("not created")

        table.insert(test_results, "❌ CREATE: Ошибка создания")
        return nil
    end
end

-- Тест 3: Поиск с условиями
local function test_query()
    -- Простое условие
    local simple_query = query_condition("score", ">", 50)
    local should = 3
    local results, err = storage_get("test_collection", 10, simple_query)
    if err then
        table.insert(test_results, "❌ SIMPLE QUERY: Ошибка: " .. err)   
    else
        if should ~= #results then
            table.insert(test_results, "❌ SIMPLE QUERY: Найдено документов: " .. #results .. " Должно: " .. should)     
        else
            table.insert(test_results, "✅ SIMPLE QUERY: Найдено документов: " .. #results)     
        end
    end
                                                           
    
    -- Сложное условие с AND
    local complex_query = query_and(
        query_condition("city", "=", "MOS"),
        query_condition("score", ">=", 100)
    )
    should = 2
    results, err = storage_get("test_collection", 10, complex_query)
    if err then
        table.insert(test_results, "❌ AND QUERY: Ошибка: " .. err)   
    else
        if should ~= #results then
            table.insert(test_results, "❌ AND QUERY:: Найдено документов: " .. #results .. " Должно: " .. should)  
        else
            table.insert(test_results, "✅ AND QUERY: Найдено документов: " .. #results)     
        end
    end
    
    -- Условие с OR
    local or_query = query_or(
        query_condition("city", "=", "KAZ"),
        query_condition("score", ">", 100)
    )
    should = 3
    results, err = storage_get("test_collection", 10, or_query)
    if err then
        table.insert(test_results, "❌ OR QUERY: Ошибка: " .. err)   
    else
        if should ~= #results then
            table.insert(test_results, "❌ OR QUERY: Найдено документов: " .. #results .. " Должно: " .. should)     
        else
            table.insert(test_results, "✅ OR QUERY: Найдено документов: " .. #results)     
        end
    end

    -- Условие возвращающее пустоту
    local emp_query = query_and(
        query_condition("city", "=", "KAZ"),
        query_condition("score", ">", 1000)
    )
    should = 0
    results, err = storage_get("test_collection", 10, emp_query)
    if err then
        table.insert(test_results, "❌ NOITEMS QUERY: Ошибка: " .. err)   
    else
        if should ~= #results then
            table.insert(test_results, "❌ NOITEMS QUERY: Найдено документов: " .. #results .. " Должно: " .. should)     
        else
            table.insert(test_results, "✅ NOITEMS QUERY: Найдено документов: " .. #results)     
        end
    end

    -- пустая квери
    should = 5
    results, err = storage_get("test_collection", 10)
    if err then
        table.insert(test_results, "❌ NO QUERY: Ошибка: " .. err)   
    else
        if should ~= #results then
            table.insert(test_results, "❌ NO QUERY: Найдено документов: " .. #results .. " Должно: " .. should)     
        else
            table.insert(test_results, "✅ NO QUERY: Найдено документов: " .. #results)     
        end
    end
end

-- Выполняем тесты
local test_id = test_create()
if test_id then
    test_get_by_id(test_id)
    test_update(test_id)
    test_delete(test_id)
end


seed_data()
test_query()
test_bulk_update()
test_get_ids()

-- Отправляем результаты
local message = "📊 *Результаты тестов Storage:*\n\n" .. table.concat(test_results, "\n")
send_message(chat_id, message)