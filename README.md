## Автор
 Даньшин Семён

## Описание
Часть сервиса аутентификации, отвечающая за генерацию и обновление токенов

## Технологии
* Go
* JWT
* MongoDB

## Маршруты

### GET /tokens/login?guid=\<guid\>
Генерация пары токенов для пользователя с переданным guid

#### Параметры
* guid - идентификатор пользователя

#### Ответ
```json
{
  "status": "success",
  "error": "",
  "data": {
    "access_token": "",
    "refresh_token": ""
  }
}
```

### GET /tokens/refresh

Обновление пары токенов

#### Тело запроса
```json
{
    "access_token": "",
    "refresh_token": ""
}
```

#### Ответ
```json
{
    "status": "success",
    "error": "",
    "data": {
        "access_token": "",
        "refresh_token": ""
    }
}
```
