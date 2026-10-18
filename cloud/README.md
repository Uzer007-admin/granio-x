# granio-cloud

Компонент независимого проекта **Granio — Гранио**.

**Исходный репозиторий:** [cloud](https://github.com/fyne-io/cloud).
**Цель форка:** Провайдеры синхронизации и облачных данных Granio.

История изменений, лицензии и сведения об исходных авторах сохранены.
Go-модули переведены на собственные пути `github.com/Uzer007-admin/…`;
минимальная версия Go — **1.27.1**. Ядро и общие API подключаются из
[`granio/v2`](https://github.com/Uzer007-admin/granio).

Состояние переноса и результаты проверок приведены в
[реестре экосистемы](https://github.com/Uzer007-admin/granio/blob/dev/granio/docs/ECOSYSTEM.md).
Наличие форка само по себе не означает проверку всех платформ и сценариев.

---

## Документация исходного проекта

Команды установки и ссылки в сохранённом тексте ниже относятся к исходнику.
Для Granio используются новые пути модулей, перечисленные выше.

# Fyne Cloud

A repository that helps to manage cloud integrations in a Fyne app.
Since Fyne 2.3 developers can provide their own cloud storage to a Fyne app,
but if you would like to use standard providers this repository will manage that for you.

There is also the ability to have developers choose cloud provider and configure settings.
Just add the following lines to your app:

```go
import "fyne.io/cloud"

func main() {
   	a := app.NewWithID("io.fyne.cloud.example")
	cloud.Enable(a)

    ...
}
```

You would then present the cloud choice UI using a button or menu item that simply calls:
```go
    cloud.ShowSettings(a, w)
```
Which will render the settings UI as a dialog on the specified window.