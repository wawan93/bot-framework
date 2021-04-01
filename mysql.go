package tgbot

type DB interface {
	Find(kind string, name string, chatID int64) (string, string, error)
}

type MysqlStorage struct {
	factories map[string]CommonHandler
	db        DB
}

func (m MysqlStorage) RegisterFactories(factories ...CommonHandler) {
	for _, factory := range factories {
		if _, ok := m.factories[factory.CommandName()]; !ok {
			m.factories[factory.CommandName()] = factory
		}
	}
}

// kind | command_name | data | name | chat_id

func (m MysqlStorage) Set(kind string, name string, chatID int64, handler CommonHandler) {
	panic("implement me")
}

func (m MysqlStorage) Get(kind string, name string, chatID int64) (CommonHandler, error) {
	commandName, data, err := m.db.Find(kind, name, chatID)
	cmd := m.factories[commandName].Deserialize(data)
	return cmd, err
}

func (m MysqlStorage) Unset(kind string, name string, chatID int64) {
	panic("implement me")
}
