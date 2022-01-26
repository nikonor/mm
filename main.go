package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	NotFound = "404"
)

type Macro struct {
	Macro      []byte
	IsVariadic bool
	F          func() []byte
}

type Mock struct {
	Code    int
	Headers map[string]string
	Body    []byte
}

var (
	mock               map[string]*Mock
	dirFlag            *string
	portFlag, nextFlag *int
	l                  sync.RWMutex
	nl                 sync.Mutex
	macros             []Macro
	partSplitRE        *regexp.Regexp
)

type H struct {
}

func uuidV4() []byte     { return []byte(uuid.New().String()) }
func getRndInt() []byte  { return []byte(strconv.FormatInt(int64(rand.Int31()), 10)) }
func getMongoId() []byte { return []byte(newMongoID()) }
func getNextInt() []byte {
	nl.Lock()
	defer nl.Unlock()
	ret := []byte(strconv.FormatInt(int64(*nextFlag), 10))
	*nextFlag++
	return ret
}
func getTime() []byte { return []byte(time.Now().Format("15:04:05")) }
func getDate() []byte { return []byte(time.Now().Format("2006-01-02")) }

func init() {
	rand.Seed(time.Now().UnixNano())

	partSplitRE = regexp.MustCompile(`:\s*`)

	macros = append(macros, Macro{
		Macro:      []byte("%v_uuid4%"),
		IsVariadic: true,
		F:          uuidV4,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%uuid4%"),
		IsVariadic: false,
		F:          uuidV4,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%increment%"),
		IsVariadic: true,
		F:          getNextInt,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%int%"),
		IsVariadic: false,
		F:          getNextInt,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%rnd_int%"),
		IsVariadic: false,
		F:          getRndInt,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%v_rnd_int%"),
		IsVariadic: true,
		F:          getRndInt,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%mongoid%"),
		IsVariadic: false,
		F:          getMongoId,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%v_mongoid%"),
		IsVariadic: true,
		F:          getMongoId,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%time%"),
		IsVariadic: true,
		F:          getTime,
	})
	macros = append(macros, Macro{
		Macro:      []byte("%date%"),
		IsVariadic: true,
		F:          getDate,
	})
}

func main() {

	flg := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Обработка флагов командной строки
	dirFlag = flg.String("d", "./", "путь к каталогу с файлами")
	portFlag = flg.Int("p", 8888, "порт, на котором запустится мок")
	nextFlag = flg.Int("n", 1, "первое число для последовательности")

	if err := flg.Parse(os.Args[1:]); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return
	}

	mock = make(map[string]*Mock)
	mock[NotFound] = &Mock{
		Code:    404,
		Headers: nil,
		Body:    nil,
	}

	println("dir=", *dirFlag, ", port=", *portFlag)

	handler := &H{}
	if err := http.ListenAndServe(":"+strconv.Itoa(*portFlag), handler); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
}

func (h *H) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	lp := randString(8)

	fmt.Println(lp + "::" + req.RequestURI + "::call")

	l.RLock()
	m, ok := mock[req.RequestURI]
	l.RUnlock()

	if !ok {
		fmt.Println(lp + "::" + req.RequestURI + "::mock not found, try get file")
		m, err = makeMock(req.RequestURI)
		if err != nil {
			fmt.Println(lp + "::" + req.RequestURI + "::new mock was not create => 404")
			l.RLock()
			m = mock[NotFound]
			l.RUnlock()
		} else {
			fmt.Println(lp + "::" + req.RequestURI + "::new mock was create")
			l.Lock()
			mock[req.RequestURI] = m
			l.Unlock()
		}
	}

	if m.Headers != nil && len(m.Headers) > 0 {
		for k, v := range m.Headers {
			fmt.Println(lp + "::" + req.RequestURI + "::add header::" + k + "=>" + v)
			resp.Header().Add(k, v)
		}
	}
	resp.WriteHeader(m.Code)
	if m.Body != nil {
		b := fillVars(m)
		if _, err = resp.Write(b); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
		b = bytes.ReplaceAll(b, []byte("\n"), []byte("\\n"))
		fmt.Println(lp + "::" + req.RequestURI + "::body::" + string(b))
	}
}

func makeMock(uri string) (*Mock, error) {
	ret := Mock{Code: 200}

	body, err := ioutil.ReadFile(*dirFlag + uri)
	if err != nil {
		return &ret, err
	}

	fill(&ret, body)

	return &ret, nil
}

func fill(m *Mock, body []byte) {
	var (
		one, two []string
		flag     bool
	)
	for _, s := range strings.Split(string(body), "\n") {
		if !flag && len(s) == 0 {
			flag = true
			continue
		}
		if flag {
			two = append(two, s)
		} else {
			one = append(one, s)
		}
	}

	m.Headers = make(map[string]string)
	if len(two) > 0 {
		for _, s := range one {
			ss := splitH(s)
			if len(ss) == 2 {
				if strings.EqualFold(ss[0], "Status-Code") {
					if code, err := strconv.Atoi(strings.TrimSpace(ss[1])); err != nil {
						fmt.Fprint(os.Stderr, "неверный формат Status-Code:"+s)
					} else {
						m.Code = code
					}
				} else {
					m.Headers[ss[0]] = ss[1]
				}
			}
		}

		m.Body = []byte(strings.Join(two, "\n"))
	} else {
		m.Body = []byte(strings.Join(one, "\n"))
	}

}

func splitH(s string) []string {
	return partSplitRE.Split(s, 2)
}

func fillVars(m *Mock) []byte {
	ret := m.Body

	for _, macro := range macros {
		if bytes.Contains(ret, macro.Macro) {
			if macro.IsVariadic {
				for bytes.Contains(ret, macro.Macro) {
					ret = bytes.Replace(ret, macro.Macro, macro.F(), 1)
				}
			} else {
				ret = bytes.ReplaceAll(ret, macro.Macro, macro.F())
			}
		}
	}

	return ret
}

func randString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
