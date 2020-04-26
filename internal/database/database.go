package database

var Database = map[string]string{
	"RandomUser1": "test1",
	"RandomUser2": "test2",
}

func New() string {
	return "database"
}
