# URL shortener API 🛠
REST API для сокращение ссылок, написанное на Go с использованием фреймворка [chi](https://github.com/go-chi/chi). В качестве СУБД используется PostgreSQL.

## API referneces

Для неавторизаованных пользователей доступен только GET-запрос:

- `GET /{alias}`. При успешном запросе произойдет временный редирект на url из БД по этому алиасу

В качестве механизма аутентификации используется BaseAuth.
Аутентифицированные пользователи могут добавлять и удалять url и их алиасы:

- `POST /url` в теле запроса нужно указать url и его алиас:

    ```json
    {
        "url":"https://google.go",
        "alias": "zxc",
    }
    ```
    В случае успешного выполнения запроса вернется json-ответ с таким содержимым:

    ```json
    {
        "status":"OK",
        "alias":{alias},
    }
    ```
    В случае какой-либо ошибки вернется json-ответ с таким содержимым:
    ```json
    {
        "status":"Error",
        "error":{errorMessage},
    }
    ```

- `DELETE /url/{alias}` удалит пару url-alias из БД. Доступен только аутентифицированным пользователям.

    В случае успешного запроса вернется json-ответ с таким содержимым:
    ```json
    {
        "status":"OK",
        "alias":{alias},
        "deletedId":{deletedId},
    }
    ```

    В случае какой-то ошибки вернется json-ответ с таким содержимым:
    ```json
    {
        "status":"Error",
        "error":{errorMessage},
    }
    ```

## Локальный запуск 🎩

### Настройка переменных окружения 🌱
Для начала нужно создать файл `.env.local`. Пример заполнения файла находится в `.env.example`:. 
- `DATABASE_URL` - полный путь к базе данных. С этими данными создаться база в контейнере. Прошу обратить внимание, что хост **менять не нужно**. `db` - это название сервиса в docker compose.
- `CONFIG_PATH` - путь до конфига. Можно указать такой путь: `"./config/local.yaml"`
- `DB_NAME` - имя БД
- `DB_USERNAME` - юзернейм от БД
- `DB_PASSWORD` - пароль от БД 

**!ВАЖНО!** нужно, чтобы значение у пароля, имени и юзера были такие же, как и в `DATABASE_URL`.

### Билд образа

Чтобы сбилдить образ, в корне проекта нужно выполнить команду:
```shell
docker-compose --env-file=.env.local build
```

### Запуск сервера
После успешного билда, нужно запустить сервер и базу данных, выполнив команду:
```shell
docker compose --env-file=.env.local up -d
```
Это запустит сервер.

Адрес сервера будет `http://127.0.0.1:8082`. Если что-то пошло не так, то вот [тред](https://stackoverflow.com/questions/62002249/docker-container-sending-request-to-http), где расписаны адреса под разные ОС.

Далее нужно накатить миграции с помощью `goose`. Если он не установлен, установить этой командой (если не установлен go - установите :)):

```shell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Для того, чтобы накатить миграции, в корне проекта надо выполнить команду:

```shell
goose -dir db/migrations postgres "postgresql://DB_USERNAME:DB_PASSWORD@127.0.0.1:5432/DB_NAME?sslmode=disable" up
```

## Локальный запуск тестов 

*!Приложение должно быть запущено!*

Для локального запуска тестов нужно зайти внутрь контейнера приложения такой командой:

```shell
docker-compose exec app sh
```

Внутри перейти в папку, например, с интеграционным тестами и запустить их:
```go
/app # cd tests/
/app/tests # go test url_shortener_test.go
```

Чтобы запустить все тесты в проекте, в корне нужно выполнить:

```go
app # go test ./...
```

Запуск smoke-тестов:

```go
app # go test --tags=smoke ./...
```

Запуск итеграционных тестов:

```go
app # go test ./...
```