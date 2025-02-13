package storage

// // пока что не используется
// // type Storage interface {
// // 	Add(data *model.URLEntry) error
// // 	Get(shortURL string) (*model.URLEntry, error)
// // 	// Close() error
// // }

// func LoadFromMemoryToFile(shortURL string, mem *memory.MemoryStorage, file *file.FileStorage) error {
// 	data, err := mem.Get(shortURL)
// 	if err != nil {
// 		return err
// 	}
// 	return file.Add(data)
// }

// func LoadAllFromFileToMemory(file *file.FileStorage, mem *memory.MemoryStorage) error {
// 	data, err := file.GetAll()
// 	if err != nil {
// 		return err
// 	}
// 	for _, val := range data {
// 		if err := mem.Add(val); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
