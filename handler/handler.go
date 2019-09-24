package handler

import (
	"encoding/json"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

const (
	passwdSalt = "#890"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	defer glog.Flush()
	if r.Method == "GET" {
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Println("failed to get data err ", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "./static/tmp/" + head.Filename,
			UpLoadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Println("Failed to create file ", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("Failed to save file ", err.Error())
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished")
}

//获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]

	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fMeta)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("OpenError")
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}

//处理注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	defer glog.Flush()
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(data)
		return
	}

	r.ParseForm()

	userName := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(userName) < 3 || len(passwd) < 5 {
		w.Write([]byte("invalid parameter"))
		return
	}

	encPasswd := util.Sha1([]byte(passwd + passwdSalt))
	suc := dblayer.UserSignUp(userName, encPasswd)

	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	defer glog.Flush()
	r.ParseForm()
	fmt.Println("SigninHandler")
	userName := r.Form.Get("username")
	passwd := r.Form.Get("password")

	glog.Info("userName ", userName, " password ", passwd)
	encPasswd := util.Sha1([]byte(passwd + passwdSalt))

	pwdChecked := dblayer.UserSignin(userName, encPasswd)

	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	token := GenToken(userName)
	ret := dblayer.UpdateToken(userName, token)
	if !ret {
		w.Write([]byte("FAILED"))
		return
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: userName,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userName := r.Form.Get("username")
	//token := r.Form.Get("token")

	/*
		isValidToken := dblayer.IsTokenValid(token)
		if !isValidToken {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	*/

	user, err := dblayer.GetUserInfo(userName)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}

	w.Write(resp.JSONBytes())
}

func GenToken(userName string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(userName + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}
