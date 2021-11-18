package app

type UrlStorage interface {
	Store(string) (string, error)
	Load(string) (string, bool, error)
}
