package db

//IDatabase interface for Key Value Database
type IDatabase interface {
	Has(id []byte) bool
	Get(id []byte) ([]byte, error)
	Set(id []byte, data []byte) error
	Dump() string
}
