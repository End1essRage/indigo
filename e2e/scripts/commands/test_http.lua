-- Тестирование HTTP модуля
local chat_id = ctx.chat_id
local results = {}

-- Тест 1: GET запрос
local resp, err = http_get("https://api.github.com/users/github", {
    ["User-Agent"] = "TestBot/1.0"
})

if resp then
    results[1] = "✅ GET: Статус " .. resp.status .. ", получено " .. #resp.body .. " байт"
else
    results[1] = "❌ GET: " .. (err or "неизвестная ошибка")
end

-- Тест 2: POST запрос
local post_body = json_encode({
    test = "data",
    timestamp = os.time()
})

local post_resp, post_err = http_post("https://httpbin.org/post", {
    ["Content-Type"] = "application/json"
}, post_body)

if post_resp then
    results[2] = "✅ POST: Статус " .. post_resp.status
else
    results[2] = "❌ POST: " .. (post_err or "неизвестная ошибка")
end

-- Тест 3: Custom метод
local custom_resp, custom_err = http_do("PUT", "https://httpbin.org/put", "test data", {
    ["Content-Type"] = "text/plain"
})

if custom_resp then
    results[3] = "✅ PUT: Статус " .. custom_resp.status
else
    results[3] = "❌ PUT: " .. (custom_err or "неизвестная ошибка")
end

local message = "🌐 *Результаты тестов HTTP:*\n\n" .. table.concat(results, "\n")
send(chat_id, message, nil)