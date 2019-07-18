package goToolMSSql2000

import (
	"database/sql"
	"fmt"
	"github.com/Deansquirrel/goToolCommon"
	_ "github.com/alexbrainman/odbc"
	"github.com/kataras/iris/core/errors"
	"time"
)

var dbCache goToolCommon.IObjectManager

var maxIdleConn int
var maxOpenConn int
var maxLifetime time.Duration

type MSSqlConfig struct {
	Server string
	Port   int
	DbName string
	User   string
	Pwd    string
}

const (
	connFormatter = "Driver={SQL Server};Server=%s,%d;Database=%s;Uid=%s;Pwd=%s;Network=DbMsSoCn;"
)

func init() {
	dbCache = goToolCommon.NewObjectManager()

	SetMaxIdleConn(15)
	SetMaxOpenConn(15)
	SetMaxLifetime(time.Second * 180)
}

func SetMaxIdleConn(n int) {
	if n > 0 {
		maxIdleConn = n
	}
}

func SetMaxOpenConn(n int) {
	if n > 0 {
		maxOpenConn = n
	}
}

func SetMaxLifetime(d time.Duration) {
	maxLifetime = d
}

//根据配置获取数据库连接
func GetConn(config *MSSqlConfig) (*sql.DB, error) {
	var conn *sql.DB
	connString := getConnStr(config)
	if dbCache.IsClosed() {
		return nil, errors.New("dbCache is closed")
	}
	if dbCache.HasId(connString) {
		conn := dbCache.GetObject(connString).(*sql.DB)
		if IsValid(conn) {
			return conn, nil
		} else {
			dbCache.Unregister() <- connString
			return GetConn(config)
		}
	}
	conn, err := getConn(connString)
	if err != nil {
		return nil, err
	}
	dbCache.Register() <- goToolCommon.NewObject(connString, conn)
	return GetConn(config)
}

//获取连接字符串
func getConnStr(config *MSSqlConfig) string {
	return fmt.Sprintf(connFormatter, config.Server, config.Port, config.DbName, config.User, config.Pwd)
}

//根据配置获取数据库连接
func getConn(connString string) (*sql.DB, error) {
	conn, err := sql.Open("odbc", connString)
	//fmt.Println(connString)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(maxIdleConn)
	conn.SetMaxOpenConns(maxOpenConn)
	conn.SetConnMaxLifetime(maxLifetime)

	return conn, nil
}

//检查数据库连接是否有效
func IsValid(db *sql.DB) bool {
	if db != nil {
		err := db.Ping()
		if err == nil {
			return true
		}
	}
	return false
}
