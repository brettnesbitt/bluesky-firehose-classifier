package repositories

// Repository defines the methods for interacting with the database.
type Repository interface {
	Insert(data interface{}) error
	FindAll() ([]interface{}, error)
	FindByID(id string) (interface{}, error)
	Update(id string, data interface{}) error
	Delete(id string) error
}
