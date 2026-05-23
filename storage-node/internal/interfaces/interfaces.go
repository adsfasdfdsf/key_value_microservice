package interfaces

type Repository interface{
	Add(key, value string) error
	Get(key string) (string, error)
}