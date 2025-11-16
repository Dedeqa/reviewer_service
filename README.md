# Reviewer Service (skeleton)

Микросервис для автоматического назначения ревьюеров на Pull Request’ы внутри команд.

---

## Описание

Сервис позволяет:

- Управлять пользователями и командами.
- Создавать Pull Request’ы (PR) и автоматически назначать до двух ревьюеров из команды автора.
- Переназначать ревьюеров на случайного активного участника команды.
- Получать список PR’ов, назначенных конкретному пользователю.
- Менять статус PR на `MERGED`. После merge изменение состава ревьюверов запрещено.

Доступ к сервису осуществляется исключительно через HTTP API.

---

## Стек

- Язык: Go
- База данных: PostgreSQL
- HTTP Router: Gorilla Mux
- Контейнеризация: Docker / docker-compose

---

## Структура проекта

```
reviewer-service/
├── cmd/
│   └── server/
│       └── main.go                 # Точка входа
├── internal/
│   ├── db/
│   │   └── db.go                  # Работа с БД
│   ├── handlers/                  # HTTP-обработчики
│   │   ├── helpers.go
│   │   ├── prs.go
│   │   ├── teams.go
│   │   └── users.go
│   ├── models/                    # Модели (User, Team, PR)
│   │   ├── pr.go
│   │   ├── team.go
│   │   └── user.go
│   ├── services/                  # Бизнес-логика
│   │   ├── pr_service.go
│   │   └── reviewer_service.go
│   └── utils/                    
│       └── reviewer_service.go    # Вспомогательная функция для работы с контекстами
├── migrations/             
│   └── 001_create_schema.sql      # Схема БД
├── Dockerfile                     # Многоступенчатая сборка
├── docker-compose.yml             # Конфигурация Docker-Compose
└── Makefile                       # Автоматизация сборки и запуска
```

---

## Запуск

1. Клонировать репозиторий:

```bash
git clone github.com/Dedeqa/reviewer_service
cd reviewer-service
```

2. Запустить сервис через Makefile или Docker-compose:

```bash
make up / docker-compose up --build
```

3. Сервис доступен на http://localhost:8080.
