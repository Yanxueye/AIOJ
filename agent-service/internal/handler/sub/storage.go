package handler

import (
	"encoding/json"
	"os"
	"sync"
)

//存储结构

type UserData struct {
	UserID  string       `json:"user_id"`
	Records []UserRecord `json:"records"`
}

type UserStorage struct {
	Users []UserData `json:"users"`
}

//存储操作

var storageMutex sync.Mutex

const RecordsFile = "records.json"

// LoadRecords 加载用户的所有做题记录
func LoadRecords(userID string) ([]UserRecord, error) {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	data, err := os.ReadFile(RecordsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []UserRecord{}, nil
		}
		return nil, err
	}

	var storage UserStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, err
	}

	for _, u := range storage.Users {
		if u.UserID == userID {
			return u.Records, nil
		}
	}
	return []UserRecord{}, nil
}

// SaveRecord 保存一条做题记录
func SaveRecord(userID, questionID string) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	var storage UserStorage
	data, err := os.ReadFile(RecordsFile)
	if err == nil {
		json.Unmarshal(data, &storage)
	}

	newRecord := UserRecord{
		QuestionID: questionID,
	}

	found := false
	for i := range storage.Users {
		if storage.Users[i].UserID == userID {
			storage.Users[i].Records = append(storage.Users[i].Records, newRecord)
			found = true
			break
		}
	}

	if !found {
		storage.Users = append(storage.Users, UserData{
			UserID:  userID,
			Records: []UserRecord{newRecord},
		})
	}

	jsonData, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(RecordsFile, jsonData, 0644)
}
