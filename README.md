This is a mini-mock.

I got bored every time I went to a project that one of my colleagues was sawing at, I asked him to set up our main mock. It's cool and can do a lot, but for positive cases, it would be easier to  have something that can be stored in text files and run from a com.strings. Simply put, a .rest file from Idea, but vice versa.

So this utility was written.

## 

## Launch

```
Usage of mm:
  -d string
        path to mock files dir (default "./")
  -n int
        first number for sequence (default 1)
  -p int
        port (default 8888)
  -pid string
        path to pid-file
  -t string
        list of access tokens, separated by ,
```

### Environment

You can set params in environment. **Environment variables has first priority**  

- MM_PORT - env key for port
- MM_DIR -  env key for dir
- MM_NEXT - env key for first number for sequence
- MM_PID -  env key for path to pid-file

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

- (2022-01-26) If you specify it among the headers `Status-Code: <n>`, it will not go to the headers, but will be used for its intended purpose.
- (2022-08-27) In the m header. specify `include: <filepath>`and then this file will be returned as the body. For an example, see . / mock/file
- (2023-03-16) Delay in seconds `X-mm-delay: <n>`(thanks to [Teimur8](https://github.com/teimur8))
- (2025-10-06) Add COND string in headers. 
  - you can add string COND in mock-file (example `./mock/cond`)
  - cond-string should like [https://github.com/nikonor/cond](https://github.com/nikonor/cond)
  - variables are taken from the query string
    - **IMPORTENT**: mm works with first parameter only (?one=1&two=2&one=3 => one==1 )
  - mm answers 200-Ok only if COND will be true
- (2025-10-08) Add IF string in header (example: ./mock/if and ./mock/if_[1,2,22])
  - FORMAT: `IF: <cond_string> => <fileName>`
- (2025-10-08) Add OPTION string in header (example: ./mock/rr and ./mock/rr_[1,2,3])
  - FORMAT `OPTION: <fileName>`
  - If you list several OPTION type lines, the files specified in them will be output in turn
- (2025-10-08) You can return body with params from query-string (example: ./mock/if and ./mock/if_[1,2,22])
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
