## Проектная работа «Превьювер изображений»
---
Запуск сервиса:
```
$ cd /tmp
$ git clone --branch develop git@github.com:spendmail/previewer.git previewer
$ cd previewer
$ make run
```
---
Проверка работы:
```
wget http://localhost:8888/fill/300/200/cdn.pixabay.com/photo/2015/04/23/22/00/tree-736885__480.jpg
file tree-736885__480.jpg
```
---
Не реализовано:
- LRU-кэш
- Проксирование HTTP-заголовков
- Контейнеризация
- Тесты