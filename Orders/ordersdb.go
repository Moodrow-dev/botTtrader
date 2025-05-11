package Orders

import (
	"botTtrader/Items"
	"botTtrader/Users"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type Order struct {
	ID       int           `json:"id"`
	Time     time.Time     `json:"time"`
	Customer *Users.User   `json:"customer_id"`
	Items    []*Items.Item `json:"items"`
	Track    string        `json:"track"`
	IsPaid   bool          `json:"is_paid"`
}

func InitDB(db *sql.DB) error {
	// Создаем таблицу заказов
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY,
		time TIMESTAMP NOT NULL,
		customer_id INTEGER NOT NULL,
		track TEXT,
		is_paid BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	// Создаем таблицу товаров в заказе
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS order_items (
		order_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		quantity INTEGER DEFAULT 1,
		PRIMARY KEY (order_id, item_id),
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES items(id)
	)`)
	return err
}

// Save сохраняет заказ в БД
func Save(order Order, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Откатываем если будет ошибка

	// Сохраняем основной заказ
	_, err = tx.Exec(
		`INSERT INTO orders (id, time, customer_id, track, is_paid) 
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			time = excluded.time,
			customer_id = excluded.customer_id,
			track = excluded.track,
			is_paid = excluded.is_paid`,
		order.ID,
		order.Time,
		order.Customer.ID,
		order.Track,
		order.IsPaid,
	)
	if err != nil {
		return err
	}

	// Удаляем старые товары заказа
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", order.ID)
	if err != nil {
		return err
	}

	// Добавляем новые товары
	for _, item := range order.Items {
		_, err = tx.Exec(
			"INSERT INTO order_items (order_id, item_id) VALUES (?, ?)",
			order.ID,
			item.ID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID возвращает заказ по ID
func GetByID(id int, db *sql.DB) (*Order, error) {
	// Начинаем транзакцию для согласованного состояния
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Получаем основной заказ
	var order Order
	var customerID int64

	err = tx.QueryRow(`
		SELECT id, time, customer_id, track, is_paid 
		FROM orders WHERE id = ?`, id).Scan(
		&order.ID,
		&order.Time,
		&customerID,
		&order.Track,
		&order.IsPaid,
	)
	if err != nil {
		return nil, err
	}

	// Получаем данные покупателя
	order.Customer, err = Users.GetByID(customerID, db)
	if err != nil {
		return nil, err
	}

	// Получаем товары заказа
	rows, err := tx.Query(`
		SELECT i.id, i.name, i.price 
		FROM order_items oi
		JOIN items i ON oi.item_id = i.id
		WHERE oi.order_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Items.Item
		err = rows.Scan(&item.ID, &item.Name, &item.Price)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &order, nil
}

// GetAllIDs возвращает все ID заказов
func GetAllIDs(db *sql.DB) ([]int, error) {
	rows, err := db.Query("SELECT id FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Printf("Ошибка чтения ID: %v", err)
			continue
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// Delete удаляет заказ по ID
func Delete(id int, db *sql.DB) error {
	res, err := db.Exec("DELETE FROM orders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("заказ %d не найден", id)
	}

	return nil
}
