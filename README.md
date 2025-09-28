<!--
SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>

SPDX-License-Identifier: GPL-3.0-only
-->

# SBT Infinity Tools

**✅ Українська** | [☑️ English](./README.en.md)

Тут зібрані наші допоміжні інструменти для роботи з іграми на базі Infinity Engine, такими як Baldur’s Gate, Baldur’s Gate II, Planescape: Torment та ін.

Читайте про локалізацію текстур у Planescape: Torment у нашій [повній статті](https://sbt.localization.com.ua/article/lokalizatsiia-tekstur-u-planescape-torment).

## Збирання проєкту

```
go build
```
🙂

## Видобуток кадрів з анімацій у форматі BAM v2

Анімації в розширених виданнях (Enhanced Editions) згаданих вище ігор зберігаються у форматі [BAM v2](https://gibberlings3.github.io/iesdp/file_formats/ie_formats/bam_v2.htm). Це бінарний формат, що містить в собі інформацію про кадри, кожний з яких складається з одного або декількох блоків. Самі блоки при цьому зберігаються в окремих файлах формату `mosXXXX.PVRZ`, де `XXXX` — це номер файла з чотирьох цифр.

Ці PVRZ-файли нерідко являють собою текстурні атласи, в котрих одне фінальне зображення порізане на купу шматочків для ефективнішого зберігання. 

Для експорту всіх окремих кадрів з BAM-анімації у вигляді PNG-файлів є команда `extract-bam`:
```
./sbt-inf extract-bam path/to/config.toml
```

### Конфігурація

Тут файл `config.toml` містить в собі інформацію про те, які вхідні файли обробляти й куди класти результат. Наприклад:
```toml
[Input]
bam = "animation.BAM"

[InputMos]
1000 = "mos2000.PNG"
1001 = "mos2001.PNG"

[Output]
extract = "output"
```
Рядок `bam` вказує на оригінальну анімацію, а розділ `InputMos` містить список усіх PVRZ, від яких залежить BAM-файл, експортованих як PNG. Для видобутку того й іншого найкраще використовувати інструмент під назвою [Near Infinity](https://github.com/NearInfinityBrowser/NearInfinity/wiki).

А `extract` — це шлях до теки, в яку збережуться результати видобутку.

Всі шляхи можуть буть як абсолютні, так і відносні (до розташування самого TOML-файла).

## Оновлення текстурних атласів для анімацій

Попри назву, команда `update-bam` не оновлює саму анімацію, але оновлює текстури, що з нею повʼязані. (Маєте на думці кращу назву — зробіть PR 😉).

Для цього в конфігураційному файлі треба також вказати список оновлених (перемальованих) кадрів:
```toml
[Input]
bam = "animation.BAM"

[InputMos]
1000 = "mos2000.PNG"
1001 = "mos2001.PNG"

[NewFrames]
1 = "Frame1.png"
6 = "Frame6.png"
7 = "Frame7.png"

[Output]
update = "override"
```

Зверніть увагу, що не обовʼязково зазначати всі кадри. Якщо якісь з них пропущені, програма натомість братиме блоки з «оригінальних» текстур, вказаних в `InputMos`.

Використання команди аналогічне до попередньої:
```
./sbt-inf update-bam path/to/config.toml
```

## Приклади

Ви можете знайти наші конфігураційні файли та перемальовані кадри для **Planescape: Torment: Enhanced Edition** у теці [`inputs`](./inputs/).

## Ліцензія

Без напрацювань з відкритим кодом від купи людей, робити локалізацію ігор було б значно складніше, якщо взагалі можливо. Відкриваючи код нашого маленького інструмента, ми хочемо бодай якось віддячити за це спільноті й сподіваємося, що комусь це стане в пригоді.

Тому весь код у цьому репозиторії доступний під ліцензією [GPL 3.0](./LICENSES/GPL-3.0-only.txt), а всі графічні ресурси (окрім оригінальних ресурсів з гри) — під ліцензією <a href="https://creativecommons.org/licenses/by-sa/4.0/">CC BY-SA 4.0</a>&nbsp;<img src="https://mirrors.creativecommons.org/presskit/icons/cc.svg" alt="" width="16px" height="16px"><img src="https://mirrors.creativecommons.org/presskit/icons/by.svg" alt="" width="16px" height="16px"><img src="https://mirrors.creativecommons.org/presskit/icons/sa.svg" alt="" width="16px" height="16px">

<a href="https://github.com/sbtlocalization/sbt-infinity">SBT Infinity Tools</a> © 2025 by <a href="https://sbt.localization.com.ua">SBT Localization</a>
