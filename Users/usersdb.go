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
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Phone        string        `json:"phone"`
	Account      string        `json:"account"`
	Address      string        `json:"address"`
	Discount     int           `json:"discount"`
	ShoppingCart []*Items.Item `json:"shopping_cart"`
	IsOwner      bool          `json:"isOwner"`
	CreatedAt    time.Time     `json:"createdAt"`
}

// NewUser creates a new user with safe defaults.
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
		ShoppingCart: []*Items.Item{},
		IsOwner:      isOwner,
		CreatedAt:    time.Now(), // Set explicitly for consistency
	}
}

// InitDB initializes the users and shopping_carts tables.
func InitDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		phone TEXT, -- Relaxed NOT NULL and UNIQUE constraints
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
		quantity INTEGER DEFAULT 1,
		PRIMARY KEY (user_id, item_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (item_id) REFERENCES items(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create shopping_carts table: %v", err)
	}
	return nil
}

// Save persists a user and their shopping cart to the database.
func Save(user *User, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for user %d: %v", user.ID, err)
	}

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
		tx.Rollback()
		return fmt.Errorf("failed to save user %d: %v", user.ID, err)
	}

	_, err = tx.Exec("DELETE FROM shopping_carts WHERE user_id = ?", user.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete shopping cart for user %d: %v", user.ID, err)
	}

	for _, item := range user.ShoppingCart {
		if item == nil || item.ID == 0 {
			tx.Rollback()
			return fmt.Errorf("invalid item in shopping cart for user %d", user.ID)
		}
		_, err = tx.Exec(
			"INSERT INTO shopping_carts (user_id, item_id, quantity) VALUES (?, ?, ?)",
			user.ID,
			item.ID,
			1, // Explicitly set quantity
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save cart item %d for user %d: %v", item.ID, user.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction for user %d: %v", user.ID, err)
	}

	log.Printf("Successfully saved user %d with %d cart items", user.ID, len(user.ShoppingCart))
	return nil
}

// GetByID retrieves a user by ID, including their shopping cart.
func GetByID(id int64, db *sql.DB) (*User, error) {
	var user User
	err := db.QueryRow(
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
			return nil, fmt.Errorf("user %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user %d: %v", id, err)
	}

	// Fetch shopping cart items
	rows, err := db.Query(`
		SELECT i.id, i.name, i.price, i.type, i.description, i.photo_id
		FROM shopping_carts sc
		JOIN items i ON sc.item_id = i.id
		WHERE sc.user_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping cart for user %d: %v", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var item Items.Item
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Type, &item.Description, &item.PhotoId)
		if err != nil {
			log.Printf("Failed to scan cart item for user %d: %v", id, err)
			continue
		}
		user.ShoppingCart = append(user.ShoppingCart, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shopping cart for user %d: %v", id, err)
	}

	return &user, nil
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

// GetOwner retrieves the owner user.
func GetOwner(db *sql.DB) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, name, phone, account, address, discount, is_owner, created_at 
		FROM users WHERE is_owner = ?`, true,
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
			return User{}, fmt.Errorf("owner not found")
		}
		return User{}, fmt.Errorf("failed to get owner: %v", err)
	}

	return user, nil
}
