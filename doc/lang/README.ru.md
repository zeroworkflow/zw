# ZeroWorkflow

[🇬🇧 EN](../../README.md)

<img src="../../assets/image/logo/light_logo.png" alt="Логотип ZeroWorkflow" width="310"/>

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org/)
[![Лицензия](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](LICENSE)
[![Версия](https://img.shields.io/badge/Version-1.1.0-purple?style=flat-square)](https://github.com/derxanax/ZeroWorkflow/releases)

> Набор инструментов для разработчиков на базе ИИ для оптимизации и автоматизации рабочих процессов

## 🧶 О проекте

ZeroWorkflow — это коллекция утилит командной строки на базе искусственного интеллекта, предназначенных для автоматизации стандартных задач разработки. Создан на Go для максимальной производительности и кросс-платформенной совместимости.

**Ключевые особенности:**
- **ИИ-ассистент в чате** - Интерактивные беседы с ИИ с подсветкой синтаксиса
- **Модульная архитектура** - Легко расширять новыми командами
- **Красивый интерфейс терминала** - Отображение markdown с подсветкой кода
- **Кросс-платформенность** - Развертывание в виде единого бинарного файла

## 🌐 Поддерживаемые платформы

[![Linux](https://img.shields.io/badge/Linux-FCC624?style=flat-square&logo=linux&logoColor=black)](https://www.linux.org/)
[![macOS](https://img.shields.io/badge/macOS-000000?style=flat-square&logo=apple&logoColor=white)](https://www.apple.com/macos/)
[![Windows](https://img.shields.io/badge/windows-0078D6?style=flat-square&logo=Windows&logoColor=white)](https://www.microsoft.com/windows/)

## ⚡️ Быстрый старт

### Установка

#### Быстрая установка 
```bash
# Скачать и установить последнюю версию
curl -L https://github.com/zeroworkflow/zw/releases/download/v1.0.9/install.sh | bash
```

### Настройка

1.  **Автоматическая настройка**:
    *   Установщик создает файл-шаблон `.env`
    *   Отредактируйте `.env` и добавьте свой токен ИИ

2.  **Получите ваш токен ИИ**:
    *   Посетите [Z.ai](https://chat.z.ai), чтобы получить ваш API-токен
    *   Добавьте его в файл `.env` или установите как переменную окружения

## 🛠 Команды

### `zw ask` - ИИ-ассистент

Интерактивный ИИ-ассистент с отображением markdown и подсветкой синтаксиса.

**Примеры:**
```bash
# Задать вопрос
zw ask "Как создать REST API на Go?"

# Прикрепить файлы для контекста
zw ask "Проверь мой код" --file main.go
zw ask "Объясни эту функцию" -f utils.go

# Интерактивный режим
zw ask -i
```

**Особенности:**
- [I] Подсветка синтаксиса для блоков кода
- ! Отображение markdown
- @ Режим интерактивного диалога
- ! **Поддержка контекста из файлов** - Прикрепляйте файлы для анализа ИИ
- [I] Красивое форматирование в терминале
- ! Безопасная обработка файлов с ограничением по размеру

## 💼 Структура проекта

```text
ZeroWorkflow/
├── src/                   # Исходный код
│   ├── cmd/    # Команды CLI
│   │   ├── root.go  # Настройка корневой команды
│   │   └── ask.go         
│   ├── internal/    # Внутренние пакеты
│   │   ├── ai/    # Реализация клиента ИИ
│   │   │   └── client.go 
│   │   └── renderer/      
│   │       └── markdown.go
│   └── main.go  # Точка входа приложения
├── assets/                
│   └── image/logo/        
├── doc/                   
│   ├── lang/              
│   └── ask.md             
├── go.mod     # Определение модуля Go
└── .env                   
```

## 🪵 Разработка

### Требования
- Go 1.21 или выше
- Терминал с поддержкой 256 цветов

### Сборка из исходного кода

#### Используя Makefile
```bash
# Клонировать репозиторий
git clone https://github.com/derxanax/ZeroWorkflow.git
cd ZeroWorkflow

# Сборка
make build

# Глобальная установка
make install

# Запуск тестов
make test

# Сборка для разработки с обнаружением гонок данных
make dev
```

### Добавление новых команд
1.  Создайте новый файл команды в `src/cmd/`
2.  Реализуйте логику команды
3.  Зарегистрируйте в корневой команде
4.  Добавьте документацию

## 📄 Лицензия

Этот проект лицензирован по лицензии MIT - подробности смотрите в файле [LICENSE](LICENSE).

---

<div align="center">
  <strong>Создано ☃️ <a href="https://github.com/derxanax">@derxanax</a></strong>
</div>