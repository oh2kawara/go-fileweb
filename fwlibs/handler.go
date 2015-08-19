package fwlibs

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// error interface
type fwhError struct{ message string }

func (e fwhError) Error() string { return e.message }

// ========== ========== ========== ==========

// http.Handler interface
type fwHandler struct {
	// ドキュメントルートスライス
	docRoot []string
	// リクエスト
	req *http.Request
	// レスポンス
	rsp http.ResponseWriter
}

// http.Handler.ServeHTTP
func (h fwHandler) ServeHTTP(rsp http.ResponseWriter, req *http.Request) {
	// 初期化
	h.req = req
	h.rsp = rsp
	// パス
	uri := "/" + req.URL.Path
	//log.Println(uri)
	var fi os.FileInfo
	var err error
	for _, dr := range h.docRoot {
		fname := filepath.Join(dr, uri)
		fi, err = os.Stat(fname)
		if !os.IsNotExist(err) {
			if fi.IsDir() {
				// 末尾が"/"で無ければ"/"をつけてリダイレクト
				last := uri[len(uri)-1]
				if last != '/' {
					req.URL.Path += "/"
					header := rsp.Header()
					header.Set("location", req.URL.String())
					rsp.WriteHeader(http.StatusMovedPermanently)
					return
				}
				// ディレクトリ
				for _, idxname := range [2]string{"index.html", "index.htm"} {
					fnameidx := filepath.Join(fname, idxname)
					fi, err = os.Stat(fnameidx)
					if !os.IsNotExist(err) && !fi.IsDir() {
						h.responseFile(fnameidx)
						return
					}
				}
			} else {
				h.responseFile(fname)
				return
			}
		}
	}
	h.response404()
}

func (h fwHandler) response404() {
	h.rsp.WriteHeader(http.StatusNotFound)
	_, err := h.rsp.Write([]byte("File Not Found."))
	if err != nil {
		log.Fatal(err)
	}
}

func (h fwHandler) response500(err error) {
	h.rsp.WriteHeader(http.StatusInternalServerError)
	mess := "Internal Server Error"
	if err != nil {
		mess += ":" + err.Error()
	}
	_, err = h.rsp.Write([]byte(mess))
	if err != nil {
		log.Fatal(err)
	}
}

func (h fwHandler) responseFile(fname string) {
	var err error
	var fp *os.File
	var fi os.FileInfo
	// log.Println(fname)
	// ファイルを出力
	for {
		fp, err = os.Open(fname)
		if err != nil {
			break
		}
		fi, err = fp.Stat()
		if err != nil {
			break
		}
		fileLen := fi.Size()
		fileData := make([]byte, fileLen)
		_, err = fp.Read(fileData)
		if err != nil {
			break
		}
		err = fp.Close()
		header := h.rsp.Header()
		header.Set("content-type", GetMimeTypeByFilename(fname))
		header.Set("content-length", strconv.FormatInt(fileLen, 10))
		h.rsp.WriteHeader(http.StatusOK)
		h.rsp.Write(fileData)
		return
	}
	if fp != nil {
		err = fp.Close()
	}
	h.response500(err)
}

// ========== ========== ========== ==========

// ドキュメントルートスライス
var docRoot []string

// ユーザホームの取得
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// ドキュメントルートの追加
func AddDocumentRoot(path string) error {
	t := path[0:1]
	if t == "~" {
		home := userHomeDir()
		path = filepath.Join(home, path[1:])
	}
	fi, err := os.Stat(path)
	if !os.IsNotExist(err) {
		if fi.IsDir() {
			docRoot = append(docRoot, path)
			return nil
		}
		return fwhError{path + " is not directory."}
	}
	return err
}

// http.Handlerの生成
func Create() (http.Handler, error) {
	var err error
	var _dr []string
	if len(docRoot) == 0 {
		var cwd string
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
		_dr = []string{cwd}
	} else {
		_dr = make([]string, len(docRoot))
		copy(_dr, docRoot)
	}
	return fwHandler{docRoot: _dr}, nil
}
