<!--
SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>

SPDX-License-Identifier: GPL-3.0-only
-->

# Конфігурація

У більшості випадків доволі незручно вказувати повний шлях до гри через параметри командного рядка для кожної команди. Натомість значно легше створити конфігураційний файл!

Цей файл може мати довільне імʼя (яке треба вказувати через `--config <filename>`), однак, якщо цього не зробити, `sbt-inf` автоматично шукатиме файл `sbt-inf.toml` у поточній директорії.

## Формат файлу

### Перелік ігор

Головний розділ файлу містить перелік ігор зі шляхами до відповідних `chitin.key`-файлів. Наприклад:

```toml
[Games]
bg1 = "/Users/gooroo/Documents/Baldur's Gate Enhanced Edition/chitin.key"
pst = "/Users/gooroo/Library/Application Support/Steam/steamapps/common/Project P/chitin.key"
```

Ключі зліва від `=` використовуватимуться у ваших подальших командах на кшталт `sbt-inf dialog ls --game bg1`.

> [!IMPORTANT]
> Насправді, можна навіть не зазначати `--game` окремо, особливо якщо у вашому файлі вказана лише одна гра. Варто лише памʼятати, що в цьому випадку `sbt-inf` використовуватиме **перший ключ за абеткою!** Тобто якщо ви маєте спочатку `pst`, а потім `bg1`, інструмент використовуватиме саме `bg1`.
>

### Параметри окремих ігор

Додатково в конфігураційному файлі можна вказати певні параметри для конкретних ігор. Для цього треба створити окремий розділ, назва якого збігається з ключем гри в переліку. Наприклад, якщо в переліку ви маєте `bg1`, то можете написати наступне:

```toml
[bg1]
dialog_site_base_url = "https://my-site.org/dialogs/bg1"
```

Параметри, що підтримуються наразі:
- `dialog_site_base_url` – те саме, що ключ `--dlg-base-url` для команди `sbt-inf text export`.

## Повний приклад

(Я використовую macOS, тому шляхи вказані через `/`. На Windows відповідно будуть `\\`).

**`sbt-inf.toml`**
```toml
[Games]
bg1 = "/Users/gooroo/Documents/Baldur's Gate Enhanced Edition/chitin.key"
bg2 = "/Users/gooroo/Documents/Baldur's Gate II Enhanced Edition/chitin.key"
iwd = "/Users/gooroo/Documents/Icewind Dale Enhanced Edition/chitin.key"
pst = "/Users/gooroo/Library/Application Support/Steam/steamapps/common/Project P/chitin.key"

[bg1]
dialog_site_base_url = "https://my-site.org/dialogs/bg1"

[bg2]
dialog_site_base_url = "https://my-site.org/dialogs/bg2"

[iwd]
dialog_site_base_url = "https://my-site.org/dialogs/iwd"

[pst]
dialog_site_base_url = "https://my-site.org/dialogs/pst"
```
