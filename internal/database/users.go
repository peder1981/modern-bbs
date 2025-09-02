package database

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User representa um usuário no sistema.
type User struct {
	ID        int64
	Username  string
	Role      string
	CreatedAt time.Time
}

// HashPassword gera um hash bcrypt para uma senha.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compara uma senha com um hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateUser cria um novo usuário no banco de dados.
func CreateUser(username, password string) (*User, error) {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar hash da senha: %w", err)
	}

	query := `INSERT INTO users (username, password_hash) VALUES (?, ?)`
	res, err := DB.Exec(query, username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar usuário: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter o ID do usuário: %w", err)
	}

	return &User{ID: id, Username: username}, nil
}

// GetUserByUsername busca um usuário pelo nome de usuário.
// GetAllUsers busca todos os usuários do sistema.
func GetAllUsers() ([]User, error) {
	query := `SELECT id, username, role, created_at FROM users ORDER BY username ASC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar usuários: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("falha ao escanear usuário: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func GetUserByUsername(username string) (*User, string, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
	row := DB.QueryRow(query, username)

	var user User
	var passwordHash string

	err := row.Scan(&user.ID, &user.Username, &passwordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", nil // Usuário não encontrado
		}
		return nil, "", fmt.Errorf("falha ao buscar usuário: %w", err)
	}

	return &user, passwordHash, nil
}

// SetUserRole atualiza o papel de um usuário no banco de dados.
// UpdateUserPassword verifica a senha atual e atualiza para a nova senha.
func UpdateUserPassword(username, currentPassword, newPassword string) error {
	// 1. Buscar o usuário e o hash da senha atual.
	user, passwordHash, err := GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("falha ao buscar usuário: %w", err)
	}
	if user == nil {
		return fmt.Errorf("usuário '%s' não encontrado", username)
	}

	// 2. Verificar se a senha atual está correta.
	if !CheckPasswordHash(currentPassword, passwordHash) {
		return fmt.Errorf("senha atual incorreta")
	}

	// 3. Gerar o hash para a nova senha.
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("falha ao gerar hash da nova senha: %w", err)
	}

	// 4. Atualizar a senha no banco de dados.
	stmt, err := DB.Prepare("UPDATE users SET password_hash = ? WHERE username = ?")
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(newPasswordHash, username)
	if err != nil {
		return fmt.Errorf("falha ao atualizar a senha: %w", err)
	}

	return nil
}

// DeleteUser remove um usuário do banco de dados.
func DeleteUser(username string) error {
	// Futuramente, pode ser necessário lidar com o conteúdo do usuário (posts, tópicos).
	// Por enquanto, a restrição FOREIGN KEY deve prevenir a deleção se houver conteúdo associado.
	stmt, err := DB.Prepare("DELETE FROM users WHERE username = ?")
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(username)
	if err != nil {
		return fmt.Errorf("falha ao deletar usuário: %w", err)
	}

	return nil
}

// AdminResetPassword define uma nova senha para um usuário sem verificar a senha antiga.
func AdminResetPassword(username, newPassword string) error {
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("falha ao gerar hash da nova senha: %w", err)
	}

	stmt, err := DB.Prepare("UPDATE users SET password_hash = ? WHERE username = ?")
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(newPasswordHash, username)
	if err != nil {
		return fmt.Errorf("falha ao atualizar a senha: %w", err)
	}

	return nil
}

func SetUserRole(username, role string) error {
	// Validação simples do papel
	switch role {
	case "user", "moderator", "admin":
		// O papel é válido, continuar
	default:
		return fmt.Errorf("papel inválido: %s", role)
	}

	stmt, err := DB.Prepare("UPDATE users SET role = ? WHERE username = ?")
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(role, username)
	if err != nil {
		return fmt.Errorf("falha ao executar statement: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("usuário '%s' não encontrado", username)
	}

	return nil
}
