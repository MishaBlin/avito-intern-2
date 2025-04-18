# Тестовое задание на стажировку Авито

## Выполнено:

1. Написан API для работы с ПВЗ со всеми функциональными требованиями
2. Настроена аутентификация по JWT токенам (/dummyLogin, /register, /login)
3. Настроены миграции для базы данных
4. Покрытие тестами более 75% строк кода
5. Написан один e2e-тест с использованием testcontainers для проверки сценария из ТЗ
6. Настроено базовое логирование запросов
7. Сконфигурирован запуск через docker-compose
8. Настроен prometheus для сбора метрик

## Стек

Go, Postgres, chi-router, testcontainers, prometheus.

## Запуск

1. Создать файл `.env`. Перенести в него данные из `example.env`
2. Выполнить команду `docker-compose up --build -d`
3. Выполнить команду для миграций бд: `docker exec pvz-service "bash" "./scripts/run-migrations.sh"`
