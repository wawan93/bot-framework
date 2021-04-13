package storage

type DB interface {
	Find(kind string, name string, chatID int64) (string, string, error)
}

type MysqlStorage struct {
	factories map[string]Serializable
	db        DB
}

func (m MysqlStorage) RegisterFactories(factories ...Serializable) {
	for _, factory := range factories {
		if _, ok := m.factories[factory.CommandName()]; !ok {
			m.factories[factory.CommandName()] = factory
		}
	}
}

func (m MysqlStorage) Set(kind string, name string, chatID int64, handler Command) {
	panic("implement me")
}

func (m MysqlStorage) Get(kind string, name string, chatID int64) (Command, error) {
	commandName, data, err := m.db.Find(string(kind), name, chatID)
	cmd := m.factories[commandName].Deserialize(data)
	return cmd, err
}

func (m MysqlStorage) Unset(kind string, name string, chatID int64) {
	panic("implement me")
}
