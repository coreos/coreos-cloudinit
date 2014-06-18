package datasource

type Datasource interface {
	ConfigRoot() string
	Fetch() ([]byte, error)
	Type() string
}
