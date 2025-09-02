package database

import (
	"fmt"
	"time"
)

// Post representa uma postagem em um tópico.
type Post struct {
	ID        int
	TopicID   int
	UserID    int
	Username  string // Para exibição, obtido com um JOIN
	Content   string
	CreatedAt time.Time
}

// CreatePost cria uma nova postagem em um tópico.
func CreatePost(topicID, userID int, content string) error {
	stmt, err := DB.Prepare("INSERT INTO posts(topic_id, user_id, content) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(topicID, userID, content)
	return err
}

// GetPostsByTopicID retorna todas as postagens de um determinado tópico, incluindo o nome do autor.
// DeletePost remove um post do banco de dados.
func DeletePost(id int) error {
	stmt, err := DB.Prepare("DELETE FROM posts WHERE id = ?")
	if err != nil {
		return fmt.Errorf("falha ao preparar statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("falha ao deletar post: %w", err)
	}

	return nil
}

func GetPostsByTopicID(topicID int) ([]*Post, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.topic_id, p.user_id, u.username, p.content, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.topic_id = ?
		ORDER BY p.created_at ASC
	`, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		if err := rows.Scan(&post.ID, &post.TopicID, &post.UserID, &post.Username, &post.Content, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}
