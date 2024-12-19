package dbaccess

import (
	"database/sql"
	"encoding/json"
	"strings"
	"sync"

	"github.com/go-jose/go-jose/v4"
	"github.com/yyboo586/IAMService/interfaces"
)

var (
	dbJWTOnce sync.Once
	dJWT      *dbJWT
)

type dbJWT struct {
	dbPool *sql.DB
}

func NewDBJWT() interfaces.DBJWT {
	dbJWTOnce.Do(func() {
		dJWT = &dbJWT{
			dbPool: dbPool,
		}
	})
	return dJWT
}

func (j *dbJWT) AddKeySet(setID string, keySet *jose.JSONWebKeySet) (err error) {
	if len(keySet.Keys) == 0 {
		return nil
	}

	format := []string{}
	values := []any{}
	for _, key := range keySet.Keys {
		data, err := json.Marshal(key)
		if err != nil {
			return err
		}

		format = append(format, "(?, ?, ?)")
		values = append(values, key.KeyID, string(data), setID)
	}
	sqlStr := "INSERT INTO t_jwt_keys(id, data, sid) VALUES" + strings.Join(format, ",")

	if _, err = j.dbPool.Exec(sqlStr, values...); err != nil {
		return err
	}

	return nil
}

func (j *dbJWT) GetKeySet(setID string) (kSet *jose.JSONWebKeySet, err error) {
	sqlStr := "SELECT data FROM t_jwt_keys WHERE sid = ? ORDER BY created_at DESC"

	rows, err := j.dbPool.Query(sqlStr, setID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	kSet = &jose.JSONWebKeySet{}
	for rows.Next() {
		var data string
		if err = rows.Scan(&data); err != nil {
			return nil, err
		}

		key := jose.JSONWebKey{}
		if err = json.Unmarshal([]byte(data), &key); err != nil {
			return nil, err
		}

		kSet.Keys = append(kSet.Keys, key)
	}

	return kSet, nil
}

func (j *dbJWT) GetKey(kid string) (key *jose.JSONWebKey, err error) {
	sqlStr := "SELECT data FROM t_jwt_keys WHERE id = ?"

	var data string
	if err = j.dbPool.QueryRow(sqlStr, kid).Scan(&data); err != nil {
		return nil, err
	}

	key = &jose.JSONWebKey{}
	if err = json.Unmarshal([]byte(data), &key); err != nil {
		return nil, err
	}

	return key, nil
}

func (j *dbJWT) AddBlacklist(id string) error {
	sqlStr := "INSERT INTO t_jwt_blacklist(id) VALUES(?)"

	if _, err := j.dbPool.Exec(sqlStr, id); err != nil {
		return err
	}

	return nil
}

func (j *dbJWT) GetBlacklist(id string) (exists bool, err error) {
	sqlStr := "SELECT COUNT(*) FROM t_jwt_blacklist WHERE id = ?"

	var count int
	if err = j.dbPool.QueryRow(sqlStr, id).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
