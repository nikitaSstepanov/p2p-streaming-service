package entity

type row interface {
    Scan(dst ...interface{}) error
}