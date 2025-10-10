-- examples/complex_queries.lua
-- Примеры сложных запросов к хранилищу

-- 1. Поиск пользователей с фильтрацией
local function find_active_users_in_city(city, min_age)
    local query = query_and(
        query_and(
            query_condition("city", "=", city),
            query_condition("age", ">=", min_age)
        ),
        query_condition("is_active", "=", true)
    )
    
    return storage_get("profiles", 100, query)
end

-- 2. Поиск с OR условиями
local function find_users_by_cities(cities)
    local query = nil
    for i, city in ipairs(cities) do
        local city_query = query_condition("city", "=", city)
        if query then
            query = query_or(query, city_query)
        else
            query = city_query
        end
    end
    
    return storage_get("profiles", 50, query)
end

-- 3. Статистика по пользователям
local function get_user_stats()
    -- Все пользователи
    local all_users = storage_get_ids("users", 1000, query_condition("user_id", ">", 0))
    
    -- Активные за последние 24 часа
    local day_ago = os.time() - 86400
    local active_users = storage_get("users", 1000, 
        query_condition("last_seen", ">", day_ago)
    )
    
    -- Новые за последнюю неделю
    local week_ago = os.time() - 604800
    local new_users = storage_get("users", 1000,
        query_condition("created_at", ">", week_ago)
    )
    
    return {
        total = #all_users,
        active_24h = #active_users,
        new_week = #new_users
    }
end

-- 4. Пакетная обработка
local function deactivate_old_users()
    local month_ago = os.time() - 2592000
    local query = query_condition("last_seen", "<", month_ago)
    
    -- Получаем ID старых пользователей
    local old_user_ids = storage_get_ids("users", 1000, query)
    
    -- Обновляем их статус
    storage_update("users", query, {
        is_active = false,
        deactivated_at = os.time()
    })
    
    -- Очищаем их из кеша
    for _, id in ipairs(old_user_ids) do
        local user = storage_get_by_id("users", id)
        if user then
            cache_set("user:" .. user.user_id, nil)
        end
    end
    
    return #old_user_ids
end