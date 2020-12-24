package db

//
//import (
//	"github.com/dgraph-io/badger/v2"
//)
//
//type badgerDb struct {
//	db *badger.DB
//}
//
//func Badger(db *badger.DB) IDatabase {
//	res := &badgerDb{
//		db: db,
//	}
//	return res
//}
//
//func (d *badgerDb) Has(id []byte) bool {
//	return d.db.View(func(txn *badger.Txn) (e error) {
//		_, e = txn.Get(id)
//		return
//	}) == nil
//}
//func (d *badgerDb) Get(id []byte) (data []byte, err error) {
//	err = d.db.View(func(txn *badger.Txn) error {
//		item, e := txn.Get(id)
//		if e != nil {
//			return e
//		}
//		e = item.Value(func(val []byte) error {
//			data = val
//			return nil
//		})
//		if e == nil {
//			return e
//		}
//
//		return nil
//	})
//	return
//}
//
//func (d *badgerDb) Set(id []byte, data []byte) error {
//	return d.db.Update(func(txn *badger.Txn) error {
//		return txn.Set(id, data)
//	})
//}
