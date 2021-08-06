package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Mock struct {
	Code    int
	Headers map[string]string
	Body    []byte
}

const (
	NotFound    = "404"
	UUID        = "%v_uuid4%"
	SameUUID    = "%uuid4%"
	INT         = "%increment%"
	SameINT     = "%int%"
	MONGOID     = "%v_mongoid%"
	SameMONGOID = "%mongoid%"
	TIME        = "%time%"
	DATE        = "%date%"
)

var (
	mock                   map[string]*Mock
	dirFlag                *string
	portFlag, nextFlag, nn *int
	l                      sync.RWMutex
	nl                     sync.Mutex
)

func main() {

	rand.Seed(time.Now().UnixNano())

	flg := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Обработка флагов командной строки
	dirFlag = flg.String("d", "./", "путь к каталогу с файлами")
	portFlag = flg.Int("p", 8888, "порт, на котором запустится мок")
	nextFlag = flg.Int("n", 1, "первое число для последовательности")

	flg.Parse(os.Args[1:])
	nn = nextFlag

	mock = make(map[string]*Mock)
	mock[NotFound] = &Mock{
		Code:    404,
		Headers: nil,
		Body:    nil,
	}

	println("dir=", *dirFlag, ", port=", *portFlag)

	handler := &H{}
	http.ListenAndServe(":"+strconv.Itoa(*portFlag), handler)

}

type H struct {
}

func (h *H) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	l.RLock()
	m, ok := mock[req.RequestURI]
	l.RUnlock()

	if !ok {
		m, err = makeMock(req.RequestURI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			l.RLock()
			m = mock[NotFound]
			l.RUnlock()
		} else {
			l.Lock()
			mock[req.RequestURI] = m
			l.Unlock()
		}
	}

	if m.Headers != nil && len(m.Headers) > 0 {
		for k, v := range m.Headers {
			resp.Header().Add(k, v)
		}
	}
	resp.WriteHeader(m.Code)
	if m.Body != nil {
		b := fillVars(m)
		resp.Write(b)
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
			ss := strings.SplitN(s, ": ", 2)
			if len(ss) == 2 {
				m.Headers[ss[0]] = ss[1]
			}
		}

		m.Body = []byte(strings.Join(two, "\n"))
	} else {
		m.Body = []byte(strings.Join(one, "\n"))
	}

}

func fillVars(m *Mock) []byte {
	ret := m.Body

	now := time.Now()

	bUUID := []byte(UUID)
	for bytes.Contains(ret, bUUID) {
		ret = bytes.Replace(ret, bUUID, []byte(uuid.New().String()), 1)
	}

	bSameUUID := []byte(SameUUID)
	if bytes.Contains(ret, bSameUUID) {
		ret = bytes.ReplaceAll(ret, bSameUUID, []byte(uuid.New().String()))
	}

	bINT := []byte(INT)
	for bytes.Contains(ret, bINT) {
		r := rand.Int31()
		fmt.Printf("r=%d\n", r)
		ret = bytes.Replace(ret, bINT, []byte(strconv.FormatInt(int64(r), 10)), 1)
	}

	bSameINT := []byte(SameINT)
	if bytes.Contains(ret, bSameINT) {
		r := rand.Int31()
		fmt.Printf("r=%d\n", r)
		ret = bytes.ReplaceAll(ret, bSameINT, []byte(strconv.FormatInt(int64(r), 10)))
	}

	// bNEXT := []byte(NEXT)
	// for bytes.Contains(ret, bNEXT) {
	// 	nl.Lock()
	// 	ret = bytes.Replace(ret, bNEXT, []byte(strconv.FormatInt(int64(*nextFlag), 10)), 1)
	// 	*nextFlag++
	// 	nl.Unlock()
	// }
	//
	// bSameNEXT := []byte(SameNEXT)
	// if bytes.Contains(ret, bSameNEXT) {
	// 	nl.Lock()
	// 	ret = bytes.ReplaceAll(ret, bSameNEXT, []byte(strconv.FormatInt(int64(*nextFlag), 10)))
	// 	*nextFlag++
	// 	nl.Unlock()
	// }

	bMONGOID := []byte(MONGOID)
	for bytes.Contains(ret, bMONGOID) {
		nl.Lock()
		ret = bytes.Replace(ret, bMONGOID, []byte(NewMongoID()), 1)
		*nextFlag++
		nl.Unlock()
	}

	bSameMONGOID := []byte(SameMONGOID)
	if bytes.Contains(ret, bSameMONGOID) {
		nl.Lock()
		ret = bytes.ReplaceAll(ret, bSameMONGOID, []byte(NewMongoID()))
		*nextFlag++
		nl.Unlock()
	}

	// 2006-01-02T15:04:05Z07:00

	bTIME := []byte(TIME)
	if bytes.Contains(ret, bTIME) {
		nl.Lock()
		ret = bytes.ReplaceAll(ret, bTIME, []byte(now.Format("15:04:05")))
		*nextFlag++
		nl.Unlock()
	}

	bDATE := []byte(DATE)
	if bytes.Contains(ret, bDATE) {
		nl.Lock()
		ret = bytes.ReplaceAll(ret, bDATE, []byte(now.Format("2006-01-02")))
		*nextFlag++
		nl.Unlock()
	}

	return ret
}
