package main

import (
	"flag"
	"fmt"
	"github.com/yuin/gopher-lua"
	"io"
	luajson "layeh.com/gopher-json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Mock struct {
	lua    *lua.LState
	lvalue lua.LValue
}

const DEFAULT_BODY = "It's Works!\n"
const DEFAULT_CODE = 200
const DEFAULT_CONTENT_TYPE = "text/plain; charset=utf-8"

func (m *Mock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dumpRequest(r)

	var lv lua.LValue
	if f, ok := m.lvalue.(*lua.LFunction); ok {
		reqTable := m.lua.NewTable()
		m.lua.SetTable(reqTable, lua.LString("method"), lua.LString(r.Method))

		urlTable := m.lua.NewTable()
		m.lua.SetTable(urlTable, lua.LString("hostname"), lua.LString(r.URL.Hostname()))
		// TODO: convert port to int
		m.lua.SetTable(urlTable, lua.LString("port"), lua.LString(r.URL.Port()))
		m.lua.SetTable(urlTable, lua.LString("path"), lua.LString(r.URL.Path))
		m.lua.SetTable(urlTable, lua.LString("querystring"), lua.LString(r.URL.RawQuery))
		m.lua.SetTable(reqTable, lua.LString("url"), urlTable)
		// TODO: put more infomations into reqTable
		if err := m.lua.CallByParam(lua.P{
			Fn:      f,
			NRet:    1,
			Protect: true,
		}, reqTable); err != nil {
			log.Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Internal Server Error\n")

			return
		}
		lv = m.lua.Get(-1) // returned value
		m.lua.Pop(1)       // remove recived value
	} else {
		lv = m.lvalue
	}

	body := strings.NewReader(DEFAULT_BODY)

	if table, ok := lv.(*lua.LTable); ok {
		code := DEFAULT_CODE
		contentType := DEFAULT_CONTENT_TYPE

		m.lua.ForEach(table, func(key, value lua.LValue) {
			switch key.String() {
			case "code":
				if lcode, ok := value.(lua.LNumber); ok {
					code = int(lcode)
				} else {
					log.Printf("not a number: %s", value)
				}
			case "contentType":
				contentType = value.String()
			case "body":
				body = strings.NewReader(value.String())
			default:
				log.Printf("unknown key: %s\n", key)
			}
		})
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(code)
	} else {
		body = strings.NewReader(lv.String())
	}

	if _, err := io.Copy(w, body); err != nil {
		log.Print(err)
	}
}

func dumpRequest(r *http.Request) {
	var w = os.Stdout

	fmt.Fprintf(w, "[%s] %s\n", time.Now().Format(time.RFC3339), r.RemoteAddr)
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", key, value)
		}
	}
	// TODO: pretty-print if it's JSON
	fmt.Fprintf(w, "\n")
	if _, err := io.Copy(os.Stdout, r.Body); err != nil {
		log.Print(err)
	}
	fmt.Fprintf(w, "\n")
}

var addr string
var scriptfile string

func init() {
	flag.StringVar(&addr, "addr", ":8080", "address to serve")
	flag.StringVar(&scriptfile, "f", "", "script filename")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <lua script>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	L := lua.NewState()
	defer L.Close()

	luajson.Preload(L)
	if err := L.DoString(`json = require('json')`); err != nil {
		log.Fatal(err)
	}

	var lvalue lua.LValue

	if scriptfile != "" {
		if err := L.DoFile(scriptfile); err != nil {
			log.Fatal(err)
		}
		lvalue = L.Get(-1)
	} else if flag.NArg() > 0 {
		code := flag.Arg(0)
		if err := L.DoString(code); err != nil {
			log.Fatal(err)
		}
		lvalue = L.Get(-1)
	} else {
		// TODO: load code from file or stdin
		lvalue = lua.LString("It's Works!\n")
	}

	m := Mock{
		lua:    L,
		lvalue: lvalue,
	}

	log.Printf("serve at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, &m))
}
