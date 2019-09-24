package db

import (
	mydb "filestore-server/db/mysql"

	"github.com/golang/glog"
)

//通过用户名和密码完成user表的注册工作
func UserSignUp(userName string, passwd string) bool {
	sql := "insert ignore into tbl_user(`user_name`,`user_pwd`) values (?,?)"
	stmt, err := mydb.DBConn().Prepare(sql)

	if err != nil {
		glog.Error("Failed to insert, err" + err.Error())
		return false
	}

	defer stmt.Close()

	ret, err := stmt.Exec(userName, passwd)
	if err != nil {
		glog.Error("Failed to insert, err:" + err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

//判断密码
func UserSignin(userName string, encPasswd string) bool {
	sql := "select * from tbl_user where user_name=? limit 1"
	stmt, err := mydb.DBConn().Prepare(sql)

	if err != nil {
		glog.Error("Failed to signin, err" + err.Error())
		return false
	}

	defer stmt.Close()

	rows, err := stmt.Query(userName)
	if err != nil {
		glog.Error("Failed to signin, err:" + err.Error())
		return false
	} else if rows == nil {
		glog.Error("userName not found :" + userName)
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encPasswd {
		return true
	}
	return false
}

func UpdateToken(userName string, token string) bool {
	sql := "replace into tbl_user_token(`user_name`,`user_token`) values (?,?)"
	stmt, err := mydb.DBConn().Prepare(sql)

	if err != nil {
		glog.Error("Failed to prepare token, err" + err.Error())
		return false
	}

	defer stmt.Close()

	_, err = stmt.Exec(userName, token)
	if err != nil {
		glog.Error("Failed to update token, err:" + err.Error())
		return false
	}

	return true
}

func IsTokenValid(token string) bool {
	return true
}

type User struct {
	UserName     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Staus        int
}

func GetUserInfo(userName string) (user *User, err error) {
	user = &User{}

	sql := "select user_name,signup_at from tbl_user where user_name=? limit 1"
	stmt, err := mydb.DBConn().Prepare(sql)

	if err != nil {
		glog.Error("Failed to prepare user_info, err" + err.Error())
		return
	}

	defer stmt.Close()

	err = stmt.QueryRow(userName).Scan(&user.UserName, &user.SignupAt)
	if err != nil {
		glog.Error("Failed to query user info, err:" + err.Error())
		return
	}

	return
}
