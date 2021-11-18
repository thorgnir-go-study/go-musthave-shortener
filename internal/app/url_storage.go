package app

type URLStorage interface {
	Store(string) (string, error)
	Load(string) (string, bool, error)
}
