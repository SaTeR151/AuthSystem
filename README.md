# AuthSystem
## При деавторизации guid пользователя удаляется из БД!

## Для запуска приложения достаточно ввести в корневом каталоге команду
```
docker-compose -f docker-compose.yaml up -d
```


## Для получения документации можно перейти по ссылке
[http://localhost:8080/swagger](http://localhost:8080/swagger)

## Тестовая конфигурация
При запуске приложения в базе данных будет два тестовых GUID пользователей
```
090bb747-d6d3-4067-a1da-2b83726eb24d
```
```
2df8716b-d385-4b7e-aae9-4618996c438a
```
## Для обновления тестовых вариантов БД можность использовать команду
```
go run ./cmd/migrate/migrate.go down; go run ./cmd/migrate/migrate.go up
```
