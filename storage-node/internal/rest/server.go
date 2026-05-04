package rest

type Storage interface{
	Add(key, value string) error
	Get(key string) (string, error)
}

type Server struct {

}