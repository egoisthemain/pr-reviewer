Сервис автоматического назначения ревьюверов для Pull Request’ов внутри команд.
Реализует HTTP API согласно предоставленной OpenAPI-спецификации.

Проект выполнен на языке Go, использует PostgreSQL, поднимается через docker-compose, содержит миграции и логику для работы с командами, пользователями и PR.

Функционал: 

Управление командами:
1. Создание команды с участниками
2. Активность/деактивация пользователя

Работа с Pull Request:
1. Создание PR (автоматическое назначение до 2 активных ревьюверов)
2. Переназначение ревьювера
3. Получение PR для конкретного ревьювера
4. Merge PR (идемпотентный)

Правила:
1. Ревьюверами могут быть только активные пользователи
2. После MERGE переназначать ревьюверов нельзя
3. Переназначение выбирает нового ревьювера из команды старого
4. Если нет доступных кандидатов → ошибка NO_CANDIDATE

Используемые технологии: 
Go
gorilla/mux — роутер
postgres/sql — работа с БД
Docker + docker-compose

Принятые решения в процессе:
1. Пользователь хранится внутри команды, поэтому поиск пользователя производится через перебор всех команд.
2. Количество ревьюверов фиксировано (до 2)
Проверяется в коде, а не через SQL constraint.
3. Идемпотентность merge
Если PR уже MERGED — просто вернуть состояние без ошибок.
4. Переназначение ревьювера
Новый ревьювер выбирается случайно из активных, кроме старого.
5. Ошибки из OpenAPI возвращаются в JSON:
"error": "PR_MERGED"
"error": "NOT_ASSIGNED"
6. Миграции применяются автоматически при запуске сервиса.

Запуск:
docker compose up --build
Сервис поднимется на http://localhost:8080

Тесты:
Ниже приведён минимум curl для проверки всех кейсов:

СОздание команды: 
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id":"u1","username":"Egor","is_active":true},
      {"user_id":"u2","username":"Alice","is_active":true},
      {"user_id":"u3","username":"Bob","is_active":true},
      {"user_id":"u4","username":"John","is_active":false}
    ]
  }'

Создание PR:
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr1",
    "pull_request_name": "Fix bug",
    "author_id": "u1"
  }'

Получить PR ревьюера:
curl "http://localhost:8080/users/getReview?user_id=u2"

Merge PR:
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr1"}'

Повторны Merge:
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "pr1"}'

Egor Tugaev
Тестовое задание Avito — 2025


