package database

import (
	"fmt"
	"time"
)

// Topic representa um tópico em um fórum.
type Topic struct {
	ID        int
	ForumID   int
	UserID    int
	Username  string // Para exibição, obtido com um JOIN
	Title     string
	CreatedAt time.Time
}

// CreateTopic cria um novo tópico no banco de dados.
func CreateTopic(forumID, userID int, title string) error {
	stmt, err := DB.Prepare("INSERT INTO topics(forum_id, user_id, title) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(forumID, userID, title)
	return err
}

// GetTopicsByForumID retorna todos os tópicos de um determinado fórum, incluindo o nome do autor.
// DeleteTopic remove um tópico e todos os seus posts.
func DeleteTopic(id int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}

	// Deleta os posts associados ao tópico
	_, err = tx.Exec("DELETE FROM posts WHERE topic_id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao deletar posts do tópico: %w", err)
	}

	// Deleta o tópico
	_, err = tx.Exec("DELETE FROM topics WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao deletar o tópico: %w", err)
	}

	return tx.Commit()
}

func GetTopicsByForumID(forumID int) ([]*Topic, error) {
	rows, err := DB.Query(`
		SELECT t.id, t.forum_id, t.user_id, u.username, t.title, t.created_at
		FROM topics t
		JOIN users u ON t.user_id = u.id
		WHERE t.forum_id = ?
		ORDER BY t.created_at DESC
	`, forumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []*Topic
	for rows.Next() {
		topic := &Topic{}
		if err := rows.Scan(&topic.ID, &topic.ForumID, &topic.UserID, &topic.Username, &topic.Title, &topic.CreatedAt); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	return topics, nil
}
