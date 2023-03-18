# mm

Это мини-мок. 

Меня задолбало каждый раз при переходе к проекту, который пилил кто-то из коллег, просить у него настройки нашего основного мока. Он крутой и может многое, но для положительных кейсов было бы проще что=то, что можно хранить в текстовых файлах и запускать из ком.строки. Проще говоря .rest файл от Idea, но наоборот.

Так и была написала эта утилита.

## Запуск

```bash
mm [-d <каталог с мок-файлам, по-умолчанию ./>]\
    [-p <порт, по-умолчанию 9999>]\ 
    [-n <начальное значение %increment%, по-умолчанию 1>]
```

## Как это работает?

Идея простая: кладем по нужным путям нужные файлы и отдаем их при запросе. 
Обратите внимание на файл section/.section. Если вам нужно иметь ссылки и /section, и /section/file, то для первой ссылки надо сделать файл section/.section

Пример:

```
./mock
├── fail
├── ok
└── section
    ├── file
    └── .section

1 directory, 4 files
```

у нас есть два файла ответа и каталог. При таком расположении мы можем получить два успешных ответа:
- /fail
- /ok
- /section/file
- /section

## Как устроен файл ответа?

Устройство простое: первые строки, до пустой строки - заголовки.

### Заголовки

(2022-01-26) Если указать среди заголовков `Status-Code: <n>`, то он не пойдет в заголовки, а будет использован по назначению.
(2022-08-27) В заголовке м. указать `include: <filepath>` и тогда в качестве тела будет отдан этот файл. Пример см. в ./mock/file
(20230-03-16) Задержка в секундах `X-mm-delay: <n>` (спасибо [Teimur8](https://github.com/teimur8))

### Тело 

После пустой строки - тело ответа.

### Пример

```
Status-Code: 202
Content-type: application/json
X-test-header: abc
X-mm-delay: 5

{
	"Description":"Вызов /section/file",
	"One": "один",
	"Two": 2
}
```

## Макросы

В файле могут быть использованы макросы:

- %v_uuid4% - новый uuidV4
- %uuid4% - uuidV4 один на запрос (повторится столько раз, сколько будет указан макрос)
- %increment% - увеличивающееся при каждом запросе число (начинается с числа, переданого в ключе -n) 
- %int% - число (повторится столько раз, сколько будет указан макрос). Будет взято по порядку из той же последовательности, что и %increment% 
- %v_mongoid% - новый mongoID 
- %mongoid% - mongoID один на запрос (повторится столько раз, сколько будет указан макрос)
- %time% - текущее время в формате ЧЧ:ММ:СС
- %date% - дата в формате ГГГГ-ММ-ДД
- %v_rnd_int% - случайное число
- %rnd_int% - случайное число (повторится столько раз, сколько будет указан макрос)

