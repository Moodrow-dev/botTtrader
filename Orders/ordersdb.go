package Orders

import (
	"botTtrader/Items"
	"botTtrader/Users"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Order struct {
	ID         int                 `json:"id"`
	Time       time.Time           `json:"time"`
	Customer   *Users.User         `json:"customer"`
	Items      map[*Items.Item]int `json:"items"` // Товары и их количество
	OrderValue float64             `json:"order_value"`
	Track      string              `json:"track"`
	IsPaid     bool                `json:"is_paid"`
	CreatedAt  time.Time           `json:"created_at"`
}

func Drop(db *sql.DB) {
	_, err := db.Exec("DROP TABLE orders")
	if err != nil {
		log.Fatal(err)
	}
}

func NewOrder(ID int, Customer *Users.User, Items map[*Items.Item]int) *Order {
	order := &Order{ID: ID, Customer: Customer, Items: Items, OrderValue: 0, CreatedAt: time.Now(), Track: "", IsPaid: false}
	order.OrderValue = order.CalculateOrderValue()
	return order
}

// InitDB инициализирует таблицы для заказов
func InitDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY,
		time TIMESTAMP NOT NULL,
		customer_id INTEGER NOT NULL,
		order_value FLOAT NOT NULL,
		track TEXT,
		is_paid BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (customer_id) REFERENCES users(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS order_items (
		order_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		PRIMARY KEY (order_id, item_id),
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES items(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create order_items table: %v", err)
	}

	return nil
}

// Save сохраняет заказ в БД
func Save(order *Order, db *sql.DB) error {
	if order == nil {
		return fmt.Errorf("nil order provided")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Сохраняем основной заказ
	_, err = tx.Exec(
		`INSERT INTO orders (id, time, customer_id, order_value, track, is_paid) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			time = excluded.time,
			customer_id = excluded.customer_id,
			order_value = excluded.order_value,
			track = excluded.track,
			is_paid = excluded.is_paid`,
		order.ID,
		order.Time,
		order.Customer.ID,
		order.OrderValue,
		order.Track,
		order.IsPaid,
	)
	if err != nil {
		return fmt.Errorf("failed to save order: %v", err)
	}

	// Удаляем старые товары заказа
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", order.ID)
	if err != nil {
		return fmt.Errorf("failed to clear order items: %v", err)
	}

	// Добавляем новые товары
	for item, quantity := range order.Items {
		if item == nil {
			continue
		}
		_, err = tx.Exec(
			"INSERT INTO order_items (order_id, item_id, quantity) VALUES (?, ?, ?)",
			order.ID,
			item.ID,
			quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to save order item: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Order %d saved successfully", order.ID)
	return nil
}

// GetByID возвращает заказ по ID
func GetByID(id int, db *sql.DB) (*Order, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Получаем основной заказ
	var order Order
	var customerID int64

	err = tx.QueryRow(`
		SELECT id, time, customer_id, order_value, track, is_paid, created_at 
		FROM orders WHERE id = ?`, id).Scan(
		&order.ID,
		&order.Time,
		&customerID,
		&order.OrderValue,
		&order.Track,
		&order.IsPaid,
		&order.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order %d not found", id)
		}
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	// Получаем данные покупателя
	order.Customer, err = Users.GetByID(customerID, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %v", err)
	}

	// Инициализируем map для товаров
	order.Items = make(map[*Items.Item]int)

	// Получаем товары заказа
	rows, err := tx.Query(`
		SELECT i.id, i.name, i.price, i.type, i.description, i.photo_id, oi.quantity
		FROM order_items oi
		JOIN items i ON oi.item_id = i.id
		WHERE oi.order_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item Items.Item
		var quantity int
		err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.Type,
			&item.Description,
			&item.PhotoId,
			&quantity,
		)
		if err != nil {
			log.Printf("Failed to scan order item: %v", err)
			continue
		}
		// Создаем новый указатель на item
		itemPtr := &item
		order.Items[itemPtr] = quantity
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading order items: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &order, nil
}

// GetAllIDs возвращает все ID заказов
func GetAllIDs(db *sql.DB) ([]int, error) {
	rows, err := db.Query("SELECT id FROM orders ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get order IDs: %v", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Printf("Failed to scan order ID: %v", err)
			continue
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading order IDs: %v", err)
	}

	return ids, nil
}

// Delete удаляет заказ по ID
func Delete(id int, db *sql.DB) error {
	res, err := db.Exec("DELETE FROM orders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order %d not found", id)
	}

	log.Printf("Order %d deleted successfully", id)
	return nil
}

// CalculateOrderValue вычисляет стоимость заказа
func (o *Order) CalculateOrderValue() float64 {
	total := 0.0
	for item, quantity := range o.Items {
		if item != nil {
			total += item.Price * float64(quantity)
		}
	}
	return total
}

// AddItem добавляет товар в заказ
func (o *Order) AddItem(item *Items.Item, quantity int) {
	if o.Items == nil {
		o.Items = make(map[*Items.Item]int)
	}
	if item != nil {
		o.Items[item] += quantity
	}
}

// RemoveItem удаляет товар из заказа
func (o *Order) RemoveItem(item *Items.Item) {
	if o.Items != nil {
		delete(o.Items, item)
	}
}

func GetOrdersOfCustomer(customerID int64, db *sql.DB) ([]*Order, error) {
	userOrders := []*Order{}
	allOrdersIDs, err := GetAllIDs(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders %v", err)
	}
	for _, id := range allOrdersIDs {
		order, err := GetByID(id, db)
		if err != nil {
			return nil, fmt.Errorf("failed to get order: %v", err)
		}
		if order.Customer.ID == customerID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders, nil
}
