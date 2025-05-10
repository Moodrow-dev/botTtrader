package Users

import (
	"botTtrader/Items"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type User struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	Phone        string       `json:"phone"`
	Account      string       `json:"account"`
	Address      string       `json:"address"`
	ShoppingCart []Items.Item `json:"shopping_cart"`
	IsOwner      bool         `json:"isOwner"`
	CreatedAt    time.Time    `json:"createdAt"`
}

// InitDB инициализирует таблицу пользователей
func InitDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		phone TEXT NOT NULL UNIQUE,
		account TEXT,
		address TEXT,
		is_owner BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	// Создаем таблицу товаров в корзине
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS order_items (
		order_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		quantity INTEGER DEFAULT 1,
		PRIMARY KEY (order_id, item_id),
		FOREIGN KEY (order_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES items(id)
	)`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы users: %v", err)
	}
	return nil
}

// Save сохраняет пользователя в БД
func Save(user User, db *sql.DB) error {
	_, err := tx.Exec(
		`INSERT INTO users (id, name, phone, account, address, is_owner) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			phone = excluded.phone,
			account = excluded.account,
			address = excluded.address,
			is_owner = excluded.is_owner`,
		user.ID,
		user.Name,
		user.Phone,
		user.Account,
		user.Address,
		user.IsOwner,
	)

	return err
}

// GetByID возвращает пользователя по ID
func GetByID(id int64, db *sql.DB) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, name, phone, account, address, is_owner, created_at 
		FROM users WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&user.Account,
		&user.Address,
		&user.IsOwner,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("пользователь %d не найден", id)
		}
		return User{}, fmt.Errorf("ошибка получения пользователя: %v", err)
	}

	return user, nil
}

// GetAllIDs возвращает все ID пользователей
func GetAllIDs(db *sql.DB) ([]int64, error) {
	rows, err := db.Query("SELECT id FROM users")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ID пользователей: %v", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("Ошибка чтения ID: %v", err)
			continue
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после чтения строк: %v", err)
	}

	return ids, nil
}

// Delete удаляет пользователя по ID
func Delete(id int64, db *sql.DB) error {
	res, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления пользователя: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("пользователь %d не найден", id)
	}

	return nil
}

// GetAll возвращает всех пользователей
func GetAll(db *sql.DB) ([]User, error) {
	rows, err := db.Query(
		`SELECT id, name, phone, account, address, is_owner, created_at 
		FROM users`,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователей: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Phone,
			&user.Account,
			&user.Address,
			&user.IsOwner,
			&user.CreatedAt,
		); err != nil {
			log.Printf("Ошибка чтения пользователя: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после чтения строк: %v", err)
	}

	return users, nil
}

func GetOwner(db *sql.DB) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, name, phone, account, address, is_owner, created_at 
		FROM users WHERE is_owner = ?`, true,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&user.Account,
		&user.Address,
		&user.IsOwner,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("владелец не найден")
		}
		return User{}, fmt.Errorf("ошибка получения владельца: %v", err)
	}

	return user, nil
}
