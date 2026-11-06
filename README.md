# Социальная сеть

Соцсеть в духе Facebook: бэкенд на **Go** (HTTP API + WebSocket + SQLite) и
фронтенд на **Next.js (React)**, каждый в своём Docker-образе.

Возможности: подписки (публичные/приватные профили, запросы на подписку), профили,
посты и комментарии с изображениями и тремя уровнями приватности, группы с
приглашениями и запросами на вступление, события, личный и групповой чат в
реальном времени (с эмодзи) и уведомления.

> **Аудиторам:** пошаговые помощники по проверке —
> [audit_helper.md](audit_helper.md) (основная соцсеть) и
> [cross_platform_appimage_audit_helper.md](cross_platform_appimage_audit_helper.md)
> (десктоп-мессенджер).

## Быстрый старт (Docker)

```bash
# из корня репозитория, нужен запущенный Docker Client
docker compose up --build
```

- Фронтенд: http://localhost:3000
- API бэкенда: http://localhost:8080

Остановить: `docker compose down`.

Seed-миграция создаёт демо-аккаунты (пароль `password123`):
`bob@demo.local` (публичный), `carol@demo.local` (приватный), `dave@demo.local` (публичный).

## Просмотр базы данных

Файл SQLite хранится в Docker-томе `backend-data`. Изнутри контейнера бэкенда:

```bash
docker exec -it social-network-backend sqlite3 /app/data/social.db
sqlite> .tables
```

## Структура проекта

```
backend/                 сервер API на Go
  server.go              точка входа (Server): роутер + middleware + запуск
  pkg/
    db/                  База данных: соединение + миграции
      migrations/sqlite/ пронумерованные файлы .up.sql / .down.sql
      sqlite/sqlite.go   открытие БД + применение миграций (golang-migrate)
    models/              репозитории (по файлу на сущность)
    handlers/            HTTP-обработчики (логика App)
    middleware/          CORS, recover, логирование
    session/             сессии на cookie
    ws/                  WebSocket-хаб (чат, presence, уведомления)
    utils/               помощники для JSON и загрузки изображений
frontend/                приложение на Next.js (App Router)
  app/                   маршруты (login, register, лента, профиль, группы, чат, уведомления)
  components/            PostCard, CreatePost, ChatBox, Nav, EmojiPicker, ConfirmDialog, …
  lib/                   API-клиент, контексты WebSocket и авторизации, типы
desktop/                 десктоп-мессенджер на Electron (Win/Linux/macOS)
docker-compose.yml       сервисы backend + frontend
build.sh / build.ps1     сборка и запуск одной командой
```

## Десктоп-приложение

Кросс-платформенный мессенджер на Electron находится в каталоге `desktop/`
(см. [desktop/README.md](desktop/README.md)). Запуск:

```bash
cd desktop
npm install
npm start            # бэкенд должен быть запущен (docker compose up)
```

Сборка установщиков: `npm run dist:win` / `npm run dist:linux` (AppImage) / `npm run dist:mac`.

## Локальная разработка (без Docker)

Бэкенд (нужны Go + C-компилятор для CGO/SQLite):

```bash
cd backend
go run .            # слушает :8080, создаёт data/social.db
```

Фронтенд:

```bash
cd frontend
npm install
npm run dev         # слушает :3000
```

## Разрешённые пакеты

Стандартная библиотека, `gorilla/websocket`, `golang-migrate`, `mattn/go-sqlite3`,
`golang.org/x/crypto/bcrypt`, `google/uuid`.
