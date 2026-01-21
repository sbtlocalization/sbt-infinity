<!--
SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>

SPDX-License-Identifier: GPL-3.0-only
-->

# SBT Infinity Tools

**✅ Українська** | [☑️ English](./README.en.md)

Тут зібрані наші допоміжні інструменти для роботи з іграми на базі Infinity Engine, такими як Baldur’s Gate, Baldur’s Gate II, Planescape: Torment та ін.

## Де взяти

Найсвіжіша версія доступна для стягування [звідси](https://github.com/sbtlocalization/sbt-infinity/releases/latest).

## Список команд і можливості

### Робота з запакованими файлами гри

Усі ресурси гри запаковані в `BIF`-файли на диску. Для роботи з ними маємо такі команди:

- `bif types` — список усіх підтримуваних типів ресурсів.
- `bif list` — перелік усіх ресурсів (і якому з `BIF`-файлів вони належать).
- `bif extract` — видобування ресурсів на диск (можна обмежити за типами та назвами).

### Робота з діалогами

- `dialog list` — перелік усіх доступних діалогів
- `dialog export` — збереження діалогів у форматі JSON Canvas.

### Робота з текстовими рядками

- `text list` — перелік текстових рядків (можна фільтрувати).
- `text export` — збереження у форматі `.xlsx`.

### Підтримка форматів WeiDU

- `tra import` — створення `TRA`-файла з `XLSX`-таблиці
- `tra update` — оновлення рядків існуючого `TRA`-файла з `CSV`.

### Інше

- `2da show` — перегляд `2DA`-таблиць.
- `csv diff` — генерація `CSV` з різницею між двома іншими `CSV`.
- `extract-bam` та `update-bam` — робота [з текстурами](docs/BAM-update.md)

## Ліцензія

Без напрацювань з відкритим кодом від купи людей, робити локалізацію ігор було б значно складніше, якщо взагалі можливо. Відкриваючи код нашого маленького інструмента, ми хочемо бодай якось віддячити за це спільноті й сподіваємося, що комусь це стане в пригоді.

Тому весь код у цьому репозиторії доступний під ліцензією [GPL 3.0](./LICENSES/GPL-3.0-only.txt), а всі графічні ресурси (окрім оригінальних ресурсів з гри) — під ліцензією <a href="https://creativecommons.org/licenses/by-sa/4.0/">CC BY-SA 4.0</a>&nbsp;<img src="https://mirrors.creativecommons.org/presskit/icons/cc.svg" alt="" width="16px" height="16px"><img src="https://mirrors.creativecommons.org/presskit/icons/by.svg" alt="" width="16px" height="16px"><img src="https://mirrors.creativecommons.org/presskit/icons/sa.svg" alt="" width="16px" height="16px">

<a href="https://github.com/sbtlocalization/sbt-infinity">SBT Infinity Tools</a> © 2025 by <a href="https://sbt.localization.com.ua">SBT Localization</a>
