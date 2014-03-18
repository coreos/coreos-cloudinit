package datasource

type Datasource interface {
	Fetch() ([]byte, error)
	Type()  string
}
