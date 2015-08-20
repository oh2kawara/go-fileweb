package fwlibs

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ドキュメントルートスライス
var docRoot []string

var zeroTime time.Time

func init() {
	zeroTime = time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
}

// ========== ========== ========== ==========

// error interface
type fwhError struct{ message string }

func (e fwhError) Error() string { return e.message }

// ========== ========== ========== ==========

// http.Handler interface
type fwHandler struct {
	// リクエスト
	req *http.Request
	// レスポンス
	rsp http.ResponseWriter
	// コンテンツファイル
	fp *os.File
}

func (h fwHandler) ServeHTTP() {
	// パス
	uri := "/" + h.req.URL.Path
	//log.Println(uri)
	var fi os.FileInfo
	var err error
	for _, dr := range docRoot {
		fname := filepath.Join(dr, uri)
		fi, err = os.Stat(fname)
		if !os.IsNotExist(err) {
			if fi.IsDir() {
				// // 末尾が"/"で無ければ"/"をつけてリダイレクト
				// last := uri[len(uri)-1]
				// if last != '/' {
				// 	h.req.URL.Path += "/"
				// 	redirect_url := h.req.URL.String()
				// 	log.Println(redirect_url)
				// 	http.Redirect(h.rsp, h.req, redirect_url, http.StatusMovedPermanently)
				// 	return
				// }
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
	http.NotFound(h.rsp, h.req)
}

func (h fwHandler) responseFile(fname string) {
	fp, err := os.Open(fname)
	if err != nil {
		http.Error(h.rsp, err.Error(), http.StatusInternalServerError)
		return
	}
	h.fp = fp

	defer func() {
		h.fp.Close()
		h.fp = nil
	}()

	h.rsp.Header().Set("content-type", GetMimeTypeByFilename(fname))
	http.ServeContent(h.rsp, h.req, fname, zeroTime, h.fp)
}

// ========== ========== ========== ==========

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

// ドキュメントルートのセットアップ。
// 空であればカレントディレクトリの追加
func SetupDocumentRoot() error {
	if len(docRoot) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		docRoot = append(docRoot, cwd)
	}
	return nil
}

// http.Handlerの生成
func Handler(rsp http.ResponseWriter, req *http.Request) {
	h := fwHandler{rsp: rsp, req: req}
	h.ServeHTTP()
}
