This is a mini-mock.

I got bored every time I went to a project that one of my colleagues was sawing at, I asked him to set up our main mock. It's cool and can do a lot, but for positive cases, it would be easier to  have something that can be stored in text files and run from a com.strings. Simply put, a .rest file from Idea, but vice versa.

So this utility was written.

## 

## Launch

```
mm [-d <directory with mock files, by default ./>]\
[-p <port, default 9999>]\
[-n <initial value %increment%, default 1>]
```

## 

## How does it work?

The idea is simple: we put the necessary files in the right paths and give them back when requested. Pay attention to the section/.section file. If you need to have both /section and /section/file links, then make a section/.section file for the first link

Example:

```
./mock
├── fail
├── ok
└── section
    ├── file
    └── .section

1 directory, 4 files
```

we have two response files and a directory. With this arrangement, we can get two successful responses:

- /fail
- /ok
- /section/file
- /section

## 

## How does the response file work?

The device is simple: the first lines, before the empty line - headers.

### 

### Headlines

(2022-01-26) If you specify it among the headers `Status-Code: <n>`, it will not go to the headers, but will be used for its intended purpose.
(2022-08-27) In the m header. specify `include: <filepath>`and then this file will be returned as the body. For an example, see . / mock/file
(20230-03-16) Delay in seconds `X-mm-delay: <n>`(thanks to [Teimur8(https://github.com/teimur8))

### Body

After an empty line - the response body.

### Example

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

## Macros

Macros can be used in the file:

- %v_uuid4% - new UUIDv4
- %uuid4% - UUIDv4 one per request (repeated as many times as the macro is specified)
- %increment% - the number that increases with each request (starts with the number passed in the-n key)
- %int% - a number (repeated as many times as the macro is specified). Will be taken in order from the same sequence as %increment%
- %v_mongoid% - new mongoID
- %mongoid% - mongoID one per request (repeated as many times as the macro is specified)
- %time% - current time in HH format:MM:SS
- %date% - date in the format YYYY-MM-DD
- %v_rnd_int% - random number
- %rnd_int% - random number (repeated as many times as the macro is specified)
