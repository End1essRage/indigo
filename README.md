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

2. Пример скрипта (welcome.lua)

local function handle()
    log("Скрипт запущен! User ID: " .. ctx.user.id)
    send_message(ctx.chat_id, "Hello, " .. ctx.user.from_name)
end

handle()
