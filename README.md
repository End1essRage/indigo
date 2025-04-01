# Indigo 🤖  
**Гибкий Lua-скриптовый движок для Telegram-ботов**  

Управляйте поведением бота через YAML-конфиг и Lua-скрипты без перекомпиляции кода.  
Идеально для: быстрого прототипирования, кастомных команд, динамических ответов.

---

## 🚀 Быстрый старт

### 1. Конфигурация (`config/config.yml`)
```yaml
bot:
  token: "<token>"
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
2. Пример скрипта (welcome.lua)
```lua
local function handle()
    log("Скрипт запущен! User ID: " .. ctx.user.id)
    send_message(ctx.chat_id, "Hello, " .. ctx.user.from_name)
end

handle()
```
Запуск
```bash
make build_and_run
```
