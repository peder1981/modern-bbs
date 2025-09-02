package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Driver do SQLite
)

var DB *sql.DB

// InitDB inicializa o banco de dados SQLite e cria as tabelas necessárias.
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir o banco de dados: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	if err := createTables(); err != nil {
		return err
	}

	return seedDatabase()
}

// createTables cria as tabelas do banco de dados se elas não existirem.
// seedDatabase popula o banco de dados com dados iniciais (usuários, fóruns, etc.)
func seedDatabase() error {
	// Criar usuário admin se não existir
	user, _, err := GetUserByUsername("admin")
	if err != nil {
		return fmt.Errorf("falha ao verificar usuário admin: %w", err)
	}
	if user == nil {
		if _, err := CreateUser("admin", "adminpass"); err != nil {
			return fmt.Errorf("falha ao criar usuário admin: %w", err)
		}
		if err := SetUserRole("admin", "admin"); err != nil {
			return fmt.Errorf("falha ao definir papel do admin: %w", err)
		}
	}

	// Criar usuário moderator se não existir
	user, _, err = GetUserByUsername("mod")
	if err != nil {
		return fmt.Errorf("falha ao verificar usuário mod: %w", err)
	}
	if user == nil {
		if _, err := CreateUser("mod", "modpass"); err != nil {
			return fmt.Errorf("falha ao criar usuário mod: %w", err)
		}
		if err := SetUserRole("mod", "moderator"); err != nil {
			return fmt.Errorf("falha ao definir papel do mod: %w", err)
		}
	}

	// Criar usuário comum se não existir
	user, _, err = GetUserByUsername("user")
	if err != nil {
		return fmt.Errorf("falha ao verificar usuário user: %w", err)
	}
	if user == nil {
		if _, err := CreateUser("user", "userpass"); err != nil {
			return fmt.Errorf("falha ao criar usuário user: %w", err)
		}
	}

	return nil
}

func createTables() error {
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user', -- 'user', 'moderator', 'admin'
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS forums (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		forum_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(forum_id) REFERENCES forums(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(topic_id) REFERENCES topics(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`

	_, err := DB.Exec(createTablesSQL)
	if err != nil {
		return fmt.Errorf("falha ao criar as tabelas: %w", err)
	}

	return nil
}
