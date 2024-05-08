# Руководство к запуску программы


## Параметры запуска


Операционная система - Linux (Ubuntu 22.04)


Для запуска **обязателен** файл .env с оформлением, указанным в .env.example. 

Запустить пиложение можно несколькими способами:
1. Через Docker (приоритетно). Нужно, чтобы переменная окружения DB_HOST = db. Затем, командой `docker compose up --build` запустить.
2. На локально машине. Нужно, чтобы переменная окружения DB_HOST = localhost. Затем, командой `go run main.go` запустить.


**Приложение слушает порт 8080.** 

## Конечные точки


### Создание команды
- **Метод**: POST
- **URL**: `/api/add-command`
- **Заголовок**:
    * `Content-Type`: application/json
    * `Authorization`: -
#### Параметры:
- `script` (string, обязательный): Содержание команды

#### Тело запроса:
```json
{
    "script": "Содержание вашей команды"
}
```
#### Пример запроса:
```
curl -d '{"script": "mkdir aa; cd aa; mkdir pp; mkdir jj; ls"}' \
-H "Content-Type: application/json" \
-X POST http://localhost:8080/api/add-command
```


### Получение списка команд
- **Метод**: GET
- **URL**: `/api/show-commands`
- **Заголовок**:
    * `Content-Type`: application/json
    * `Authorization`: ---
#### Параметры: ---


### Получение команды по id
- **Метод**: POST
- **URL**: `/api/show-commands/{id}`
- **Заголовок**:
    * `Content-Type`: application/json
    * `Authorization`: ---
#### Параметры: ---


## Ошибки
- `400 Bad Request` - неверный запрос 
- `405 Method Not Allowed` - неверный метод запроса





