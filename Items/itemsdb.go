package Items

import (
	"database/sql"
	"fmt"
	_ "gi
	"log"
	"time"
)

type Item struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"createdAt"`
}

// InitDB инициализирует таблицу товаров
func InitDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL CHECK(price >= 0),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы items: %v", err)
	}
	return nil
}

// Save сохраняет товар в БД
func Save(item Item, db *sql.DB) error {
	_, err := db.Exec(
		`INSERT INTO items (id, type, name, description, price) 
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			name = excluded.name,
			description = excluded.description,
			price = excluded.price`,
		item.ID,
		item.Type,
		item.Name,
		item.Description,
		item.Price,
	)
	return err
}

// GetByID возвращает товар по ID
func GetByID(id int, db *sql.DB) (Item, error) {
	var item Item
	err := db.QueryRow(
		`SELECT id, type, name, description, price, created_at 
		FROM items WHERE id = ?`,
		id,
	).Scan(
		&item.ID,
		&item.Type,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, fmt.Errorf("товар %d не найден", id)
		}
		return Item{}, fmt.Errorf("ошибка получения товара: %v", err)
	}

	return item, nil
}

// GetAll возвращает все товары
func GetAll(db *sql.DB) ([]Item, error) {
	rows, err := db.Query(
		`SELECT id, type, name, description, price, created_at 
		FROM items ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения товаров: %v", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.CreatedAt,
		); err != nil {
			log.Printf("Ошибка чтения товара: %v", err)
			continue
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после чтения строк: %v", err)
	}

	return items, nil
}

// Delete удаляет товар по ID
func Delete(id int, db *sql.DB) error {
	res, err := db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления товара: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("товар %d не найден", id)
	}

	return nil
}

// GetByType возвращает товары определенного типа
func GetByType(itemType string, db *sql.DB) ([]Item, error) {
	rows, err := db.Query(
		`SELECT id, type, name, description, price, created_at 
		FROM items WHERE type = ? ORDER BY name`,
		itemType,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения товаров по типу: %v", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.CreatedAt,
		); err != nil {
			log.Printf("Ошибка чтения товара: %v", err)
			continue
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после чтения строк: %v", err)
	}

	return items, nil
}
