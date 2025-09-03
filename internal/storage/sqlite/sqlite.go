package sqlite

import (
	"database/sql"
	"fmt"
	"money-tracker/internal/model"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage/sqlite/New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := createUsersTable(db); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := createTransactionsTable(db); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(email string, password string) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	var isExistEmail bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&isExistEmail)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if isExistEmail {
		return 0, fmt.Errorf("%s: the email already exist", op)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := s.db.Prepare("INSERT INTO users (email, password_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(email, string(hashedPassword))

	if err != nil {
        return 0, fmt.Errorf("%s: %w", op, err)
    }

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	const op = "storage.sqlite.GetUserByEmail"

	var user model.User
	err := s.db.QueryRow(`
        SELECT id, email, password_hash, created_at 
        FROM users WHERE email = ?
    `, email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUserByID(id int64) (*model.User, error) {
	const op = "storage.sqlite.GetUserByID"

	var user model.User
	err := s.db.QueryRow(`
		SELECT id, email, password, created_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err) 
	}

	return &user, nil
}

func (s *Storage) SaveTransaction(tr model.Trasaction) (int64, error) {
	const op = "storage.sqlite.SaveTransaction"

	var isExitsUser bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", tr.UserID).Scan(&isExitsUser)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if !isExitsUser {
		return 0, fmt.Errorf("%s: the user not found", op)
	}

	stmt, err := s.db.Prepare(`INSERT INTO transactions (user_id, amount, category, description, date)
		VALUES (?, ?, ?, ?, ?)`)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(tr.UserID, tr.Amount, tr.Category, tr.Description, tr.Date)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func createUsersTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func createTransactionsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			category TEXT NOT NULL,
			description TEXT DEFAULT '',
			date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	return err
}