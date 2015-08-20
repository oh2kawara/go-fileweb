package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/oh2kawara/go-fileweb/fwlibs"
)

const defaultPort int = 8050

type fwConfig struct {
	// port
	Port int
	// DocumentRoot
	Root string
	// DocumentRoots
	Roots []string
}

var port int
var confFile string

func init() {
	flag.IntVar(&port, "p", defaultPort, "WebServerPort")
	flag.StringVar(&confFile, "conf", "", "ConfigFile(json)")
}

func addRoot(path string) {
	err := fwlibs.AddDocumentRoot(path)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	// コマンドライン解析
	flag.Parse()
	args := flag.Args()

	for _, path := range args {
		addRoot(path)
	}

	// 設定ファイル解析
	if confFile != "" {
		var c fwConfig
		var jsondata []byte
		jsondata, err = ioutil.ReadFile(confFile)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(jsondata, &c)
		if err != nil {
			log.Fatal(err)
		}
		if c.Port > 0 {
			port = c.Port
		}
		if c.Root != "" {
			addRoot(c.Root)
		}
		for _, path := range c.Roots {
			addRoot(path)
		}
	}

	http.HandleFunc("/", fwlibs.Handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
