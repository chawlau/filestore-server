package db

import (
	mydb "filestore-server/db/mysql"
	"time"

	"github.com/golang/glog"
)

//UserFile 用户文件表结构体
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

func OnUserFileUploadFinished(userName, fileHash, fileName string, fileSize int64) bool {
	sql := "insert ignore into tbl_user_file (`user_name`, `file_sha1`," +
		"`file_name`, `file_size`, `upload_at`) values (?,?,?,?,?)"
	stmt, err := mydb.DBConn().Prepare(sql)

	if err != nil {
		glog.Error("update tbl_user_file prepare failed " + err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(userName, fileHash, fileName, fileSize, time.Now())

	if err != nil {
		glog.Error("update tbl_user_file failed " + err.Error())
		return false
	}
	return true
}

//批量获取用户文件信息
func QueryUserFileMetas(userName string, limit int) (userFiles []*UserFile, err error) {
	sql := "select file_sha1,file_name,file_size,upload_at,last_update from " +
		"tbl_user_file where user_name=? limit ?"

	stmt, err := mydb.DBConn().Prepare(sql)
	if err != nil {
		glog.Info("QueryUserFileMetas err ", err.Error())
		return
	}

	defer stmt.Close()

	row, err := stmt.Query(userName, limit)
	if err != nil {
		glog.Info("QueryUserFileMetas err ", err.Error())
		return
	}

	for row.Next() {
		uFile := &UserFile{}

		err = row.Scan(&uFile.FileHash, &uFile.FileName, &uFile.FileSize,
			&uFile.UploadAt, &uFile.LastUpdated)

		if err != nil {
			glog.Info("query failed err" + err.Error())
			break
		}

		userFiles = append(userFiles, uFile)
	}
	return
}
