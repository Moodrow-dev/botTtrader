package Users

import (
	"botTtrader/Items"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mymmrac/telego"
)

type User struct {
	ID           int64               `json:"id"`
	Name         string              `json:"name"`
	Phone        string              `json:"phone"`
	Account      string              `json:"account"`
	Address      string              `json:"address"`
	Discount     int                 `json:"discount"`
	ShoppingCart map[*Items.Item]int `json:"shopping_cart"` // Key: item, Value: quantity
	IsOwner      bool                `json:"isOwner"`
	CreatedAt    time.Time           `json:"createdAt"`
}

func Drop(db *sql.DB) {
	_, err := db.Exec("DROP TABLE users")
	if err != nil {
		log.Fatal(err)
	}
}

// NewUser создает нового пользователя с безопасными значениями по умолчанию
func NewUser(user *telego.User, isOwner bool) *User {
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if name == "" {
		name = "Unknown"
	}
	return &User{
		ID:           user.ID,
		Name:         name,
		Phone:        "",
		Account:      user.Username,
		Address:      "",
		Discount:     0,
		ShoppingCart: make(map[*Items.Item]int),
		IsOwner:      isOwner,
		CreatedAt:    time.Now(),
	}
}

// InitDB инициализирует таблицы пользователей и корзин
func InitDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		phone TEXT,
		account TEXT,
		address TEXT,
		discount INTEGER DEFAULT 0,
		is_owner BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS shopping_carts (
		user_id INTEGER NOT NULL,
		item_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 1,
		PRIMARY KEY (user_id, item_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
	)`)
	if err != nil {
		return fmt.Errorf("failed to create shopping_carts table: %v", err)
	}

	return nil
}

// Save сохраняет пользователя и его корзину
func Save(user *User, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Сохраняем основную информацию о пользователе
	_, err = tx.Exec(
		`INSERT INTO users (id, name, phone, account, address, discount, is_owner) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			phone = excluded.phone,
			account = excluded.account,
			address = excluded.address,
			discount = excluded.discount,
			is_owner = excluded.is_owner`,
		user.ID,
		user.Name,
		user.Phone,
		user.Account,
		user.Address,
		user.Discount,
		user.IsOwner,
	)
	if err != nil {
		return fmt.Errorf("failed to save user: %v", err)
	}

	// Очищаем старую корзину
	_, err = tx.Exec("DELETE FROM shopping_carts WHERE user_id = ?", user.ID)
	if err != nil {
		return fmt.Errorf("failed to clear shopping cart: %v", err)
	}

	// Сохраняем товары в корзине
	for item, quantity := range user.ShoppingCart {
		_, err = tx.Exec(
			"INSERT INTO shopping_carts (user_id, item_id, quantity) VALUES (?, ?, ?)",
			user.ID,
			item.ID,
			quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to save cart item: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetByID получает пользователя по ID с корзиной
func GetByID(id int64, db *sql.DB) (*User, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Получаем основную информацию о пользователе
	var user User
	err = tx.QueryRow(
		`SELECT id, name, phone, account, address, discount, is_owner, created_at 
		FROM users WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Phone,
		&user.Account,
		&user.Address,
		&user.Discount,
		&user.IsOwner,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// Инициализируем корзину
	user.ShoppingCart = make(map[*Items.Item]int)

	// Получаем товары в корзине
	rows, err := tx.Query(`
		SELECT i.id, i.name, i.price, i.type, i.description, i.photo_id, sc.quantity
		FROM shopping_carts sc
		JOIN items i ON sc.item_id = i.id
		WHERE sc.user_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping cart: %v", err)
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
			log.Printf("Failed to scan cart item: %v", err)
			continue
		}
		user.ShoppingCart[&item] = quantity
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading cart items: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &user, nil
}

// Вспомогательные методы для работы с корзиной
func (u *User) AddToCart(item *Items.Item, quantity int) {
	if u.ShoppingCart == nil {
		u.ShoppingCart = make(map[*Items.Item]int)
	}
	u.ShoppingCart[item] += quantity
}

func (u *User) RemoveFromCart(item Items.Item) {
	delete(u.ShoppingCart, &item)
}

func (u *User) ClearCart() {
	u.ShoppingCart = make(map[*Items.Item]int)
}

func (u *User) CartTotal() float64 {
	total := 0.0
	for item, quantity := range u.ShoppingCart {
		total += item.Price * float64(quantity)
	}
	return total
}

// GetAllIDs returns all user IDs.
func GetAllIDs(db *sql.DB) ([]int64, error) {
	rows, err := db.Query("SELECT id FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to get user IDs: %v", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("Failed to scan user ID: %v", err)
			continue
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user IDs: %v", err)
	}

	return ids, nil
}

// Delete removes a user by ID.
func Delete(id int64, db *sql.DB) error {
	res, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user %d: %v", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion for user %d: %v", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user %d not found", id)
	}

	log.Printf("Successfully deleted user %d", id)
	return nil
}

// GetAll returns all users (without shopping carts).
func GetAll(db *sql.DB) ([]User, error) {
	rows, err := db.Query(
		`SELECT id, name, phone, account, address, discount, is_owner, created_at 
		FROM users`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
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
			&user.Discount,
			&user.IsOwner,
			&user.CreatedAt,
		); err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}

	return users, nil
}

// GetOwnerID retrieves the owner user.
func GetOwnerID(db *sql.DB) (int64, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, is_owner FROM users WHERE is_owner = ?`, true,
	).Scan(
		&user.ID,
		&user.IsOwner,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("owner not found")
		}
		return 0, fmt.Errorf("failed to get owner: %v", err)
	}

	return user.ID, nil
}
