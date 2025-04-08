# Indigo 🤖  
**Гибкий Lua-скриптовый движок для Telegram-ботов**  

Управляйте поведением бота через YAML-конфиг и Lua-скрипты без перекомпиляции кода.  
Идеально для: быстрого прототипирования, кастомных команд, динамических ответов.

---

## 🚀 Быстрый старт

### 1. Конфигурация
```yaml
bot:
  mode: "polling"  # Или "webhook"
  
commands:
  - name: "start"
    description: "Приветствие"
    handler: "welcome.lua"
  - name: "menu"
    description: "тест меню"
    reply:
      keyboard: "main_menu"

keyboards:
  - name: "main_menu"
    type: "inline" # или "reply"
    message: "main"
    buttons:
      - row:
          - name: "btn1"
            text: "Кнопка 1"
            handler: "btn1_handler.lua"
          - name: "btn2"
            text: "Кнопка 2"
            callback_data: "customData"
            handler: "btn1_handler.lua"
      - row:
          - name: "btn3"
            text: "Кнопка 1"
            handler: "btn1_handler.lua"
          - name: "btn4"
            text: "Кнопка 2"
            handler: "btn1_handler.lua"
```
2. Переменные окружения:
BOT_TOKEN
CONFIG_PATH
3. Пример скрипта (welcome.lua)
```lua
local function handle()
    log("Скрипт запущен! User ID: " .. ctx.user.id)
    send_message(ctx.chat_id, "Hello, " .. ctx.user.from_name)
end

handle()
```
Запуск
```bash
task build_and_run
```

# хранилище
```yaml
# Файловое хранилище
storage:
  type: "file"
  file:
    path: "./data"

# MongoDB
storage:
  type: "mongo"
  mongo:
    uri: "mongodb://localhost:27017"
    db: "bot_db"
```

# клавиатуры
```yaml
keyboards:
  - name: "age_buttons"
    message: "Выберите возрастную группу:"
    buttons:
      - row:
          - text: "18-25"
            data: "18-25"
          - text: "26-35"
            data: "26-35"
```

```lua
function create_dynamic_keyboard()
  local buttons = {
    Rows = {}
  }
  
  for i = 1, 5 do
    table.insert(buttons.Rows, {
      {Text = "Кнопка "..i, Data = "btn"..i}
    })
  end
  
  send_keyboard(ctx.chat_id, "Выберите вариант:", buttons)
end
```

# формы
```yaml
forms:
  - name: "user_reg"
    stages:
      - field: "email"
        message: "Введите email:"
        validation:
          type: "email"
      - field: "age_group"
        message: "Выберите возраст:"
        keyboard: "age_buttons"
    script: "form_complete.lua"
```

```lua
function handle()
  local data = ctx.form_data
  
  if not string.match(data.email, "^.+@.+%..+$") then
    send_message(ctx.user.id, "❌ Неверный email!")
    return
  end
  
  storage_save("users", ctx.user.id, data)
  send_message(ctx.user.id, "✅ Регистрация завершена!")
end
```

# http интеграции
```yaml
http:
  endpoints:
    - path: "/notify"
      method: "POST"
      scheme: "auth_scheme"
      script: "notify.lua"
```

# Lua

контекст выполнения
```lua
ctx = {
  chat_id = 123456789,     -- ID текущего чата
  user = {
    id = 987654321,        -- ID пользователя
    name = "Иван"          -- Имя пользователя
  },
  form_data = {            -- Данные заполненной формы
    name = "Мария",
    email = "test@example.com"
  },
  req_data = {}
  cb_data = {}
  text = "/start"          -- Текст сообщения
}
```

работа с ботом
```lua
-- Отправка сообщения
send_message(chat_id, "Привет, мир!")

-- Отправка клавиатуры
send_keyboard(chat_id, "Выберите:", {
  Rows = {
    {
      {Text = "Да", Data = "yes"},
      {Text = "Нет", Data = "no"}
    }
  }
})
```

работа с кэшом
```lua
cache_set("temp_data", "123")
local value = cache_get("temp_data")
```

работа с хранилищем
```lua
-- Сохранение данных
local success, err = storage_save("users", "user123", {
  name = "Василий",
  age = 30
})

-- Загрузка данных
local data = storage_load("users", "user123")
```

http модуль
```lua
-- Простой GET запрос
local res, err = http_get("https://api.example.com/data")
if res then
    log("Status: " .. res.status)
    log("Body: " .. res.body)
end

-- POST запрос с заголовками
local json_body = [[{"name": "Lua Bot", "version": 1.0}]]
local headers = {
    ["Content-Type"] = "application/json",
    ["X-Custom-Header"] = "lua-request"
}

local post_res = http_post("https://api.example.com/update", json_body, headers)
if post_res then
    cache_set("last_response", post_res.body)
end
```
