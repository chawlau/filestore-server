package main

import (
	"filestore-server/handler"
	"fmt"
	"net/http"
)

func main() {
	//flag.Parse()
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/multimeta", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)

	//秒传接口
	http.HandleFunc("/file/fastupload", handler.TryFastUploadHandler)

	//分块上传
	http.HandleFunc("/file/mpupload/init", handler.InitalMultipartUploadHandler)
	http.HandleFunc("/file/mpupload/uppart", handler.UploadPartHandler)
	http.HandleFunc("/file/mpupload/complete", handler.CompleteUploadHandler)

	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SigninHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))
	err := http.ListenAndServe(":9097", nil)
	if err != nil {
		fmt.Println("Failed to start server, err ", err.Error())
	}
}
