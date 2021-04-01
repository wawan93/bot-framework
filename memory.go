package tgbot

type InMemoryStorage struct {
	handlers map[string]map[int64]CommonHandler
}

func (i InMemoryStorage) Set(kind string, name string, chatID int64, handler CommonHandler) {
	panic("implement me")
}

func (i InMemoryStorage) Get(kind string, name string, chatID int64) (CommonHandler, error) {

}

func (i InMemoryStorage) Unset(kind string, name string, chatID int64) {
	panic("implement me")
}
