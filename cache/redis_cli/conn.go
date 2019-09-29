package redir_cli

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

var (
	pool      *redis.Pool
	redisHost = "192.168.31.186:6381"
	redisPass = "testupload"
)

func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisHost)

			if err != nil {
				glog.Error("redis dial error ", err.Error())
				return nil, err
			}

			return c, nil
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
}

func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}
