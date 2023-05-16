package main

import (
	"bytes"
	"flag"
	"fmt"
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
	TokenKey = "X-Mm-Auth-Token"
)

type Macro struct {
	Macro      []byte
	IsVariadic bool
	F          func() []byte
}

type Mock struct {
	Code        int
	Headers     map[string]string
	Body        []byte
	FileModTime time.Time
}

var (
	mock                         map[string]*Mock
	dirFlag, pidFlag, tokensFlag *string
	portFlag, nextFlag           *int
	l                            sync.RWMutex
	nl                           sync.Mutex
	macros                       []Macro
	partSplitRE, sep             *regexp.Regexp
	tokens                       []string
	needToken                    bool
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
	sep = regexp.MustCompile(`[/?]`)

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
	dirFlag = flg.String("d", "./", "path to mock files dir")
	portFlag = flg.Int("p", 8888, "port")
	nextFlag = flg.Int("n", 1, "first number for sequence")
	pidFlag = flg.String("pid", "", "path to pid-file")
	tokensFlag = flg.String("t", "", "list of access tokens, separated by ,")

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
	if pidFlag != nil && len(*pidFlag) > 0 {
		println("pid-file=" + *pidFlag)
		lockFile, err := os.Create(*pidFlag)
		defer func() {
			if err = lockFile.Close(); err != nil {
				fmt.Fprint(os.Stderr, err.Error())
			}
			if err = os.Remove(*pidFlag); err != nil {
				fmt.Fprint(os.Stderr, err.Error())
			}
		}()
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		} else {
			_, err = lockFile.WriteString(strconv.Itoa(os.Getpid()))
			if err != nil {
				fmt.Fprint(os.Stderr, err.Error())
			}

		}
	}
	if tokensFlag != nil && len(*tokensFlag) > 0 {
		tokens = strings.Split(*tokensFlag, ",")
		needToken = true
	}

	handler := &H{}
	if err := http.ListenAndServe(":"+strconv.Itoa(*portFlag), handler); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
}

func getM(uri string) (*Mock, bool) {
	l.RLock()
	defer l.RUnlock()
	m, ok := mock[uri]

	if !ok {
		for _, u := range getCases(uri) {
			if m, ok = mock[u]; ok {
				return m, ok
			}
		}
	}

	return m, ok
}

func getCases(uri string) []string {

	var r []string
	r = append(r, uri)
	uu := strings.Split(uri, "/")
	for i := len(uu) - 1; i > 1; i-- {
		r = append(r, strings.Join(uu[:i], "/"))

	}
	return r
}

func isValidToken(t string) bool {
	if len(t) == 0 {
		return false
	}
	for _, tt := range tokens {
		if t == tt {
			return true
		}
	}
	return false
}

func (h *H) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	lp := randString(8)

	fmt.Println(lp + "::" + req.RequestURI + "::call")

	if needToken {
		if !isValidToken(req.Header.Get(TokenKey)) {
			l.RLock()
			m := mock[NotFound]
			l.RUnlock()
			resp.WriteHeader(m.Code)
			if _, err = resp.Write(m.Body); err != nil {
				fmt.Fprint(os.Stderr, err.Error())
			}
			return
		}
	}

	uri := strings.TrimRight(req.RequestURI, "/")
	if req.Method == "GET" {
		uu := strings.SplitN(uri, "?", 2)
		uri = uu[0]
	}

	m, ok := getM(uri)

	if ok && !mockCheck(m, uri) {
		fmt.Println(lp + "::" + uri + "::recreate mock")
		ok = false
	}

	if !ok {
		fmt.Println(lp + "::" + uri + "::mock not found, try get file")
		m, err = makeMock(uri)
		if err != nil {
			fmt.Println(lp + "::" + uri + "::new mock was not create => 404")
			l.RLock()
			m = mock[NotFound]
			l.RUnlock()
		} else {
			fmt.Println(lp + "::" + uri + "::new mock was create")
			l.Lock()
			mock[uri] = m
			l.Unlock()
		}
	}

	if m.Headers != nil && len(m.Headers) > 0 {
		for k, v := range m.Headers {
			fmt.Println(lp + "::" + uri + "::add header::" + k + "=>" + v)
			resp.Header().Add(k, v)
		}
	}

	HeaderHandlerDelay(lp+"::"+uri, m.Headers)

	resp.WriteHeader(m.Code)
	if m.Body != nil {
		b := fillVars(m)
		if _, err = resp.Write(b); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
		b = bytes.ReplaceAll(b, []byte("\n"), []byte("\\n"))
		fmt.Println(lp + "::" + uri + "::body::" + string(b))
	}
}

// проверяем время изменения файла
func mockCheck(m *Mock, uri string) bool {
	for _, u := range getCases(uri) {
		stat, err := os.Stat(*dirFlag + u)
		if err != nil {
			return false
		}

		return m.FileModTime.Equal(stat.ModTime())
	}
	return false
}

func makeMock(uri string) (*Mock, error) {
	var (
		err  error
		body []byte
	)
	ret := Mock{Code: 200}

	for _, u := range getCases(uri) {
		if _, err = os.ReadDir(*dirFlag + u); err == nil {
			partOfURI := strings.Split(uri, "/")
			u += "/." + partOfURI[len(partOfURI)-1]
			fmt.Println("it's dir. New path will be " + u)
		}

		body, err = os.ReadFile(*dirFlag + u)
		if err != nil {
			continue
		}

		stat, err := os.Stat(*dirFlag + u)
		if err != nil {
			continue
		}

		ret.FileModTime = stat.ModTime()

		fill(&ret, body)

		return &ret, nil
	}

	return nil, err
}

func fill(m *Mock, body []byte) {
	var (
		one, two []string
		twoBody  []byte
		flag     bool
		err      error
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
	for _, s := range one {
		ss := splitH(s)
		if len(ss) == 2 {
			switch {
			case strings.EqualFold(ss[0], "Status-Code"):
				if code, err := strconv.Atoi(strings.TrimSpace(ss[1])); err != nil {
					fmt.Fprint(os.Stderr, "неверный формат Status-Code:"+s)
				} else {
					m.Code = code
				}
			case strings.EqualFold(ss[0], "include"):
				twoBody, err = os.ReadFile(*dirFlag + "/" + ss[1])
				if err != nil {
					fmt.Fprint(os.Stderr, err.Error()+"filename="+*dirFlag+"/"+ss[1])
					return
				}

			default:
				m.Headers[ss[0]] = ss[1]
			}
		}
	}

	switch {
	case len(twoBody) > 0:
		m.Body = twoBody
	case len(two) > 0:
		m.Body = []byte(strings.Join(two, "\n"))
	default:
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

// Header Handlers
func HeaderHandlerDelay(lp string, headers map[string]string) {
	for k, v := range headers {
		if strings.EqualFold(k, "X-mm-delay") {
			delaySec, _ := strconv.Atoi(v)
			if delaySec > 0 {
				fmt.Println(lp+"::"+"delaySec=", delaySec)
				dur := time.Duration(delaySec) * time.Second
				time.Sleep(dur)
				break
			}
		}
	}
}
