# Инструкция

Данный проект состоит из двух компонент: трекера и клиента. Трекер нужен для того, чтобы клиент при скачивании файла мог получить список пиров, которые раздают этот документ. Клиент запускается локально и в течение всей работы слушает входящие запросы, раздает те файлы которые есть у этого пользователя. Также у клиента есть две команды upload и download. Первая возвращает торрент файл документа и регистрирует его в трекере (чтобы его можно было скачивать).
Вторая может по торрент файлу скачивает у других пользователей нужный файл. 

## Трекер

1. Трекер реализован на питоне с помощью фремворка Flask, т.к. этот способ выглядит наименее простым.
2. Поднять с `localhost:5000` или кастомным хостом:
```
python3 tracker/server.py
python3 tracker/server.py 0.0.0.0 9545
```
4. Хранит пиры внутри памяти, поэтому при прекращении работы все данные будут утеряны.
5. Обработка запросов:
```
POST http://localhost:5000/hash 
{
  "listening_port": 300,
  "listening_host": "0.0.0.0"
}
```
Возвращает уникальный хэш (uuid), который нужен для индентификации документов (каждому раздаваемому документу соответсвует свой уникальный хэш). И добавляет пир(хост + порт) в список известных пиров для этого хэша, то есть те пиры, которые раздают этот документ. 
```
POST localhost:5000/peers
{
  "hash": "ca74aa76-e687-45e6-87c2-8fa9fc23d063",
  "listening_port": 300,
  "listening_host": "0.0.0.0"
}
```
Возвращает список всех пиров, которые раздают документ с этим хэшем и добавляет запрашиваемый хост и порт в список известных хостов.

## Клиент

1. Примеры конфигурации клиента находяться в cfgs:
```
{
  "self": {
    "service": "../data/test1", # Относительный путь к папка где храняться торрент файлы (и файлы для раздачи)
    "host": "0.0.0.0", # Хост на котором клиент слушает другие запросы
    "port": 1000 # Порт на котором клиент слушает другие запросы
  },
  "tracker": {
    "host": "http://0.0.0.0", # Хост трекера
    "port": 5000 # Порт трекера
  }"
}
```
2. Запуск локально. Аргумент это относительный путь к конфигу:
```
python3 client/main.py cfgs/cfg.json
```
3. Команда upload:
  ```
  upload path_to_file path_to_torrent
  ```
  По файлу создает торрент файл. Ожидаемый вывод: SUCCESSFUL UPLOAD
4. Команда download:
```
download path_to_torrent path_to_file
```
Загружает файл по торрент файлу. Ожидаемый вывод: SUCCESSFUL DOWNLOAD 
5. Процесс загрузки может прерываться, его можно восстановить запустив команду еще раз.
6. Клиент можно выключать, все торрент файлы и файлы храняться в данных
7. Общение между клиентами происходит напрямую по сокетам. 
8. Каждый сегмент проверяется на хэш сумму, что позволяет немного защититься от злоумышленников. 

## Пример торрент файла:
```
{
  "FileInfo": {
    "Size": 2048,
    "PartSize": 256,
    "Parts": [
      "4ebea4f2cc3c93321608b94eb98cf2c1280d7072",
      "0a552d04e3435c346332454b1024d41d1e802abe",
      "693b04c64f7f1633a49057ca51431a0bb4bc9650",
      "26c9a81bba906a7e09b2bd13aaabb1a6be205b9b",
      "e863d1678c6f6ffb2165b53842e66e0d14a05650",
      "7aaa8e1d2d46a7fada13c5acd084323b3021d9ba",
      "8299079c05ecf7cf39880ed609ac0b4bcfc84316",
      "6225a56e16b52710fbd64a3120d03584e7c82a29"
    ],
    "Hash": "29964bae-53e9-4c3f-b620-9c583aeb64b5"
  }
}
```
Hash -> индентификатор документа (получается с помощью регистрации в трекере)

## Очевидные недоработки
1. In-memory хранилище в трекере. Можно хотя бы поднять редис
2. Нет прогресса загрузки, что не очень удобно для загрузки больших данных
3. Клиенты при скачивании файлов не пересматривают пиры с которых они берут документы (из-за этого не видят новые пиры)
4. Плохое логгирование
5. Нет тестирования
