package database

import (
	"database/sql"
)

const Connection = "x:x@tcp(prototype2-auroradb.cpcbsvtonq3r.eu-central-1.rds.amazonaws.com:3306)/x"
const DbEngine = "mysql"

var instance *sql.DB

func GetConnection() (db *sql.DB , err error){
	if instance == nil {
		x, err := sql.Open(DbEngine, Connection)
		if err != nil {
			return nil, err
		}

		instance = x
	}

	return instance, nil
}