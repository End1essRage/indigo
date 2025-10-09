-- Примеры использования Query Builder в Lua

-- 1. Создание простого условия
local query1 = query_condition("age", ">", 18)
print("Query 1: " .. query1:toString())  -- age > 18

-- 2. Создание условия с равенством
local query2 = query_condition("status", "=", "active")
print("Query 2: " .. query2:toString())  -- status = active

-- 3. Комбинирование условий с AND
local query3 = query_condition("age", ">=", 18):and(query_condition("age", "<=", 65))
print("Query 3: " .. query3:toString())  -- (age >= 18) AND (age <= 65)

-- 4. Комбинирование условий с OR
local query4 = query_condition("role", "=", "admin"):or(query_condition("role", "=", "moderator"))
print("Query 4: " .. query4:toString())  -- (role = admin) OR (role = moderator)

-- 5. Сложный запрос с вложенными условиями
local ageQuery = query_condition("age", ">=", 18):and(query_condition("age", "<=", 65))
local statusQuery = query_condition("status", "=", "active")
local roleQuery = query_condition("role", "=", "admin"):or(query_condition("role", "=", "moderator"))

local complexQuery = ageQuery:and(statusQuery):and(roleQuery)
print("Complex Query: " .. complexQuery:toString())

-- Альтернативный способ создания сложных запросов
local complexQuery2 = query_and(
    query_condition("status", "=", "active"),
    query_or(
        query_condition("role", "=", "admin"),
        query_condition("role", "=", "moderator")
    )
)

-- =======================
-- Примеры работы с Storage
-- =======================

-- 1. Создание записи
local newUser = {
    name = "John Doe",
    age = 25,
    status = "active",
    role = "user"
}
local success, id = storage_create("users", newUser)
if success then
    print("Created user with ID: " .. id)
end

-- 2. Получение записи по ID
local user = storage_get_by_id("users", id)
print("User name: " .. user.name)

-- 3. Поиск записей с условиями
-- Найти всех активных пользователей старше 18 лет
local query = query_condition("age", ">", 18):and(query_condition("status", "=", "active"))
local users = storage_get("users", 10, query)  -- Получить максимум 10 записей

for i, user in ipairs(users) do
    print(i .. ". " .. user.name .. " (age: " .. user.age .. ")")
end

-- 4. Получение одной записи по условию
local adminQuery = query_condition("role", "=", "admin")
local admin = storage_get_one("users", adminQuery)
if admin.name then
    print("Admin found: " .. admin.name)
end

-- 5. Получение только ID записей
local activeUsersQuery = query_condition("status", "=", "active")
local ids = storage_get_ids("users", 100, activeUsersQuery)
print("Found " .. #ids .. " active users")

-- 6. Обновление записей по условию
local updateData = {
    last_login = os.time(),
    login_count = 1
}
local oldUsersQuery = query_condition("last_login", "<", os.time() - 86400*30)  -- Не логинились 30 дней
local updatedCount = storage_update("users", oldUsersQuery, updateData)
print("Updated " .. updatedCount .. " old users")

-- 7. Обновление по ID
local updateUser = {
    name = "Jane Doe",
    age = 26
}
local ok = storage_update_by_id("users", id, updateUser)
if ok then
    print("User updated successfully")
end

-- 8. Удаление записей по условию
local inactiveQuery = query_condition("status", "=", "inactive"):and(
    query_condition("last_login", "<", os.time() - 86400*90)  -- Не логинились 90 дней
)
local deletedCount = storage_delete("users", inactiveQuery)
print("Deleted " .. deletedCount .. " inactive users")

-- 9. Удаление по ID
local deleted = storage_delete_by_id("users", id)
if deleted then
    print("User deleted successfully")
end

-- =======================
-- Практические примеры
-- =======================

-- Пример 1: Поиск товаров с фильтрами
function findProducts(minPrice, maxPrice, category, inStock)
    local query = query_condition("price", ">=", minPrice):and(
        query_condition("price", "<=", maxPrice)
    )
    
    if category then
        query = query:and(query_condition("category", "=", category))
    end
    
    if inStock then
        query = query:and(query_condition("stock", ">", 0))
    end
    
    return storage_get("products", 50, query)
end

-- Пример 2: Поиск пользователей по различным критериям
function findUsers(options)
    local query = nil
    
    if options.minAge then
        local ageQuery = query_condition("age", ">=", options.minAge)
        query = query and query:and(ageQuery) or ageQuery
    end
    
    if options.maxAge then
        local ageQuery = query_condition("age", "<=", options.maxAge)
        query = query and query:and(ageQuery) or ageQuery
    end
    
    if options.roles and #options.roles > 0 then
        local roleQuery = nil
        for _, role in ipairs(options.roles) do
            local cond = query_condition("role", "=", role)
            roleQuery = roleQuery and roleQuery:or(cond) or cond
        end
        query = query and query:and(roleQuery) or roleQuery
    end
    
    if options.status then
        local statusQuery = query_condition("status", "=", options.status)
        query = query and query:and(statusQuery) or statusQuery
    end
    
    return storage_get("users", options.limit or 100, query)
end

-- Использование
local products = findProducts(10.00, 100.00, "electronics", true)
local users = findUsers({
    minAge = 18,
    maxAge = 65,
    roles = {"admin", "moderator"},
    status = "active",
    limit = 50
})