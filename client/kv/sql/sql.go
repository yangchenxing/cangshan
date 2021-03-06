package sqlkv

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/kv"
	"github.com/yangchenxing/cangshan/client/sql"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("SQLKV", new(SQLKV))
}

type SQLKV struct {
	sync.Mutex
	DB            *sql.DB
	CreateQueries []string
	GetQuery      string
	SetQuery      string
	CleanQuery    string
	CleanInterval time.Duration
	initialized   bool
}

func (k *SQLKV) Initialize() error {
	if k.DB == nil {
		return errors.New("Missing DB in SQLKV")
	}
	if k.GetQuery == "" {
		return errors.New("Missing GetQuery")
	} else if k.SetQuery == "" {
		return errors.New("Missing SetQuery")
	}
	return nil
}

func (k *SQLKV) initializeTable() error {
	k.Lock()
	defer k.Unlock()
	if !k.initialized {
		k.initialized = true
		if len(k.CreateQueries) > 0 {
			for _, query := range k.CreateQueries {
				if _, err := k.DB.Exec(query); err != nil {
					return fmt.Errorf("Create SQLKV table fail: %s", err.Error())
				}
			}
		}
		if k.CleanQuery != "" {
			go k.autoClean()
		}
	}
	return nil
}

func (k *SQLKV) Ping() error {
	return k.DB.Ping()
}

func (k *SQLKV) Get(key string) ([]byte, error) {
	if err := k.initializeTable(); err != nil {
		return nil, err
	}
	row := k.DB.QueryRow(k.GetQuery, key)
	var value []byte
	if err := row.Scan(&value); err == sql.ErrNoRows {
		return nil, kv.ErrNotFound
	} else if err != nil {
		return nil, err
	} else {
		return value, nil
	}
}

func (k *SQLKV) Set(key string, value []byte, maxAge time.Duration) error {
	if err := k.initializeTable(); err != nil {
		return err
	}
	_, err := k.DB.Exec(k.SetQuery, key, value, time.Now().Add(maxAge).Unix())
	return err
}

func (k *SQLKV) autoClean() {
	for {
		if _, err := k.DB.Exec(k.CleanQuery, time.Now().Unix()); err != nil {
			logging.Error("Clean SQLKV fail: %s", err.Error())
		}
		time.Sleep(k.CleanInterval)
	}
}
