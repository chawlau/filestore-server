package handler

import (
	rPool "filestore-server/cache/redis_cli"
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

func InitalMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileSize, err := strconv.Atoi(r.Form.Get("filesize"))

	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//生成分块上传的初始化信息
	upInfo := &MultipartUploadInfo{
		FileHash:   fileHash,
		FileSize:   fileSize,
		UploadID:   userName + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(fileSize) / (5 * 1024 * 1024))),
	}

	//初始化信息写到redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//userName := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	filePath := "./static/file/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(filePath), 0744)
	fd, err := os.Create(filePath)
	defer fd.Close()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}

	buf := make([]byte, 1024*1024)

	readLen := 0
	for {
		n, err := r.Body.Read(buf)
		readLen += n
		fd.Write(buf[:n])
		if err != nil {
			glog.Info("read buf falied", err.Error())
			break
		}
	}

	glog.Info("chunkIndex ", chunkIndex, " size total ", readLen)

	//更新redis缓存状态
	replay, err := redis.Values(rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1))

	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	upId := r.Form.Get("uploadid")
	userName := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	fileSize := r.Form.Get("filesize")
	fileName := r.Form.Get("filename")

	//通过uploadid查询redis并判断是否所有分块上传完成
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upId))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}

	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount += 1
		}
	}

	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-1, "invalid request", nil).JSONBytes())
		return
	}

	//合并分块

	//更新唯一文件表和用户文件表
	fSize, _ := strconv.Atoi(fileSize)
	dblayer.OnFileUploadFinished(fileHash, fileName, int64(fSize), "")
	dblayer.OnUserFileUploadFinished(userName, fileHash, fileName, int64(fSize))

	//响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
