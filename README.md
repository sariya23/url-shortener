# URL shortener API 🛠
REST API для сокращение ссылок, написанное на Go с использованием фреймворка [chi](https://github.com/go-chi/chi). В качестве СУБД используется PostgreSQL.

## API refernces

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
    В случае усрешного выполнения запроса вернется json-ответ с таким содержимым:

    ```
    {
        "status":"OK",
        "alias":{alias_from_body},
    }
    ```
    В случае какой-либо ошибки вернется json-ответ с таким содержимым:
    ```
    {
        "status":"Error",
        "error":{errorMessage},
    }
    ```

- `DELETE /url/{alias}` удалит пару url-alias из БД. Доступен только Аутентифицированным пользователям.

    В случае успешного запроса вернется json-ответ с таким содержимым:
    ```
    {
        "status":"OK",
        "alias":{alias},
        "deletedId":{deletedId},
    }
    ```

    В случае какой-то ошибки вернется json-ответ с таким содержимым:
    ```
    {
        "status":"Error",
        "error":{errorMessage},
    }
    ```