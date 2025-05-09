package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func write(users []User, db *sql.DB) error {
	for _, user := range users {
		deleteUser(db, user.ID)
	}
	err := createTable(db)
	if err != nil {
		return err
	}
	for _, user := range users {
		_, exec := db.Exec("INSERT INTO users (id, name, phone, account, address, isOwner) VALUES (?, ?, ?, ?, ?, ?)", user.ID, user.Name, user.Phone, user.Account, user.Address, user.isOwner)
		if exec != nil {
			return exec

		}
	}
	return nil
}

func read(id int64, db *sql.DB) (User, error) {
	// В запросе нужно добавить WHERE для фильтрации по id
	row := db.QueryRow("SELECT id, name, phone, account, address, isOwner FROM chats WHERE id = ?", id)

	var (
		Id      int64
		Name    string
		Phone   int
		Account string
		Address string
		IsOwner bool
	)

	err := row.Scan(&Id, &Name, &Phone, &Account, &Address, &IsOwner)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("Клиент %d не найден", id)
		}
		return User{}, err
	}

	user := User{
		ID:      Id,
		Name:    Name,
		Phone:   Phone,
		Account: Account,
		Address: Address,
		isOwner: IsOwner,
	}

	return user, nil
}

func pickOverIds(db *sql.DB) ([]int64, error) {
	raw, err := db.Query("SELECT id FROM users")
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	var ids []int64

	for raw.Next() {
		var id int64

		err = raw.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func deleteUser(db *sql.DB, id int64) error {
	result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("Клиент %v не удален", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Эмм хз лол че за ошибка :| %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Клиент %d не найден", id)
	}

	return nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name STRING,
		phone int,
		account STRING,
		address STRING,
		isOwner BOOLEAN,
    	name TEXT NOT NULL,
    	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	return nil
}
