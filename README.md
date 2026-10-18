# Расширения Granio

Общие компоненты интерфейса и интеграции независимого фреймворка Granio.

**Основной исходник:** https://github.com/fyne-io/fyne-x.
**Цель форка:** совместное развитие дополнительных виджетов, терминала,
описаний интерфейса, облачных провайдеров и определения локали.
Истории исходных библиотек и имеющиеся LICENSE/AUTHORS сохранены.
Полный реестр источников и коммитов — [SOURCES.json](SOURCES.json).

| Пакет | Назначение | Исходник |
| --- | --- | --- |
| `widget`, `layout`, `theme`, `dialog`, `menu`, `wrapper`, `data` | Компоненты интерфейса | [fyne-x](https://github.com/fyne-io/fyne-x) |
| `terminal` | Терминальный виджет, ANSI, PTY/ConPTY | [terminal](https://github.com/fyne-io/terminal) |
| `ui` | Чтение, сериализация и редактирование описаний интерфейса | [refyne](https://github.com/fyne-io/refyne) |
| `cloud` | Облачные провайдеры и настройки | [cloud](https://github.com/fyne-io/cloud) |
| `locale` | Определение системного языка | [go-locale](https://github.com/fyne-io/go-locale) |

Минимальная версия Go — 1.27.1. Все пакеты используют один Go-модуль.
Импорт терминала: `github.com/Uzer007-admin/granio-x/terminal`.
Другие пакеты доступны по тому же префиксу: `/ui`, `/cloud`, `/locale`.

Установка согласованной версии:

```sh
go get github.com/Uzer007-admin/granio-x@22a6db2b4c9ad36d73460eb3d8e626caf68ad2df
go test -race -tags migrated_fynedo,ci ./...
```

Приложения и демонстрации перенесены в
[granio-examples](https://github.com/Uzer007-admin/granio-examples):
`terminal`, `extensions`, `cloud`, `localization`.
Библиотека терминала остаётся доступна для встраивания, в том числе в Studio.
Старые импорты `granio-terminal`, `granio-ui`, `granio-cloud`, `granio-locale`
следует заменить путями из таблицы. [Материалы исходного проекта](UPSTREAM.md).
