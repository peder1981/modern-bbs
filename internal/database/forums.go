package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Forum representa um fórum no banco de dados.
type Forum struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

// CreateForum cria um novo fórum no banco de dados.
func CreateForum(name, description string) (*Forum, error) {
	stmt, err := DB.Prepare("INSERT INTO forums(name, description) VALUES(?, ?)")
	if err != nil {
		return nil, fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(name, description)
	if err != nil {
		return nil, fmt.Errorf("falha ao executar statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter último ID inserido: %w", err)
	}

	return &Forum{
		ID:          id,
		Name:        name,
		Description: description,
	}, nil
}

// GetAllForums retorna todos os fóruns do banco de dados.
// UpdateForum atualiza o nome e a descrição de um fórum existente.
func UpdateForum(id int64, name, description string) error {
	stmt, err := DB.Prepare("UPDATE forums SET name = ?, description = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, description, id)
	if err != nil {
		return fmt.Errorf("falha ao atualizar fórum: %w", err)
	}

	return nil
}

// DeleteForum remove um fórum e, em cascata, seus tópicos e posts.
func DeleteForum(id int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}

	// Deleta os posts associados aos tópicos do fórum
	_, err = tx.Exec(`DELETE FROM posts WHERE topic_id IN (SELECT id FROM topics WHERE forum_id = ?)`, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao deletar posts do fórum: %w", err)
	}

	// Deleta os tópicos do fórum
	_, err = tx.Exec("DELETE FROM topics WHERE forum_id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao deletar tópicos do fórum: %w", err)
	}

	// Deleta o fórum
	_, err = tx.Exec("DELETE FROM forums WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao deletar o fórum: %w", err)
	}

	return tx.Commit()
}

func GetAllForums() ([]Forum, error) {
	rows, err := DB.Query("SELECT id, name, description, created_at FROM forums ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar fóruns: %w", err)
	}
	defer rows.Close()

	var forums []Forum
	for rows.Next() {
		var forum Forum
		// O scan para a descrição pode ser nulo, então precisamos tratar isso.
		var description sql.NullString
		if err := rows.Scan(&forum.ID, &forum.Name, &description, &forum.CreatedAt); err != nil {
			return nil, fmt.Errorf("falha ao escanear linha do fórum: %w", err)
		}
		if description.Valid {
			forum.Description = description.String
		}
		forums = append(forums, forum)
	}

	return forums, nil
}
