package storage

type URLStorage interface {
	Store(string) (string, error)
	Load(string) (string, bool, error)
}
