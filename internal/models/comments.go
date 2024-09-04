package models

import (
	"database/sql"
	"errors"
	"time"
)

type CommentModelInterface interface {
	Insert(snippetID int, author string, content string) (int, error)
	GetBySnippetID(snippetID int) ([]*Comment, error)
	Get(id int) (*Comment, error)
	Update(id int, content string) error
	Upvote(commentID, userID int) (string, error)
	Downvote(commentID, userID int) (string, error)
	Delete(id int) error
}

// Comment representa um comentário no banco de dados.
type Comment struct {
	ID        int
	SnippetID int
	Author    string
	Content   string
	Created   time.Time
	Updated   time.Time
	Upvotes   int
}

// CommentModel encapsula uma pool de conexões sql.DB.
type CommentModel struct {
	DB *sql.DB
}

// Insert insere um novo comentário no banco de dados.
func (m *CommentModel) Insert(snippetID int, author string, content string) (int, error) {
	stmt := `INSERT INTO comments (snippet_id, content, author, created, updated, upvotes)
	         VALUES(?, ?, ?, UTC_TIMESTAMP(), UTC_TIMESTAMP(), 0)`

	result, err := m.DB.Exec(stmt, snippetID, author, content)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetBySnippetID retorna todos os comentários associados a um snippet específico.
func (m *CommentModel) GetBySnippetID(snippetID int) ([]*Comment, error) {
	stmt := `SELECT id, snippet_id, author, content, created, updated, upvotes 
	         FROM comments WHERE snippet_id = ? ORDER BY created ASC`

	rows, err := m.DB.Query(stmt, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*Comment{}

	for rows.Next() {
		c := &Comment{}
		err = rows.Scan(&c.ID, &c.SnippetID, &c.Author, &c.Content, &c.Created, &c.Updated, &c.Upvotes)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// Update atualiza o conteúdo de um comentário existente.
func (m *CommentModel) Update(id int, content string) error {
	stmt := `UPDATE comments SET content = ?, updated = UTC_TIMESTAMP() WHERE id = ?`

	_, err := m.DB.Exec(stmt, content, id)
	if err != nil {
		return err
	}

	return nil
}

// Upvote altera o número de votos de um comentário.
func (m *CommentModel) Upvote(commentID int, userID int) (string, error) {
	// Verifica o tipo de voto do usuário
	var voteType string
	err := m.DB.QueryRow(`SELECT vote_type FROM comment_votes WHERE comment_id = ? AND user_id = ?`, commentID, userID).Scan(&voteType)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	switch voteType {
	case "upvote":
		// Remove o upvote
		_, err = m.DB.Exec(`DELETE FROM comment_votes WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		if err != nil {
			return "", err
		}
		// Atualiza o número de upvotes
		_, err = m.DB.Exec(`UPDATE comments SET upvotes = upvotes - 1 WHERE id = ?`, commentID)
		if err != nil {
			return "", err
		}
		return "Vote removed!", nil
	case "downvote":
		// Atualiza o voto para upvote
		_, err = m.DB.Exec(`UPDATE comment_votes SET vote_type = 'upvote' WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		if err != nil {
			return "", err
		}
		// Atualiza o número de upvotes
		_, err = m.DB.Exec(`UPDATE comments SET upvotes = upvotes + 2 WHERE id = ?`, commentID)
		if err != nil {
			return "", err
		}
		return "Vote updated to upvote!", nil
	default:
		// Adiciona o upvote
		_, err = m.DB.Exec(`INSERT INTO comment_votes (comment_id, user_id, vote_type) VALUES (?, ?, 'upvote')`, commentID, userID)
		if err != nil {
			return "", err
		}
		// Atualiza o número de upvotes
		_, err = m.DB.Exec(`UPDATE comments SET upvotes = upvotes + 1 WHERE id = ?`, commentID)
		if err != nil {
			return "", err
		}
		return "Vote successfully registered!", nil
	}
}

// Downvote altera o número de votos de um comentário.
func (m *CommentModel) Downvote(commentID int, userID int) (string, error) {
	// Verifica o tipo de voto do usuário
	var voteType string
	err := m.DB.QueryRow(`SELECT vote_type FROM comment_votes WHERE comment_id = ? AND user_id = ?`, commentID, userID).Scan(&voteType)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	switch voteType {
	case "downvote":
		// Remove o downvote
		_, err = m.DB.Exec(`DELETE FROM comment_votes WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		if err != nil {
			return "", err
		}
		// Atualiza o número de upvotes
		_, err = m.DB.Exec(`UPDATE comments SET upvotes = upvotes + 1 WHERE id = ?`, commentID)
		if err != nil {
			return "", err
		}
		return "Vote removed!", nil
	case "upvote":
		// Atualiza o voto para downvote
		_, err = m.DB.Exec(`UPDATE comment_votes SET vote_type = 'downvote' WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		if err != nil {
			return "", err
		}
		// Atualiza o número de upvotes
		_, err = m.DB.Exec(`UPDATE comments SET upvotes = upvotes - 2 WHERE id = ?`, commentID)
		if err != nil {
			return "", err
		}
		return "Vote updated to downvote!", nil
	default:
		// Adiciona o downvote
		_, err = m.DB.Exec(`INSERT INTO comment_votes (comment_id, user_id, vote_type) VALUES (?, ?, 'downvote')`, commentID, userID)
		if err != nil {
			return "", err
		}
		// Atualiza o número de upvotes
		_, err = m.DB.Exec(`UPDATE comments SET upvotes = upvotes - 1 WHERE id = ?`, commentID)
		if err != nil {
			return "", err
		}
		return "Vote successfully registered!", nil
	}
}


// Delete remove um comentário do banco de dados.
func (m *CommentModel) Delete(id int) error {
	stmt := `DELETE FROM comments WHERE id = ?`

	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Get retorna um comentário específico pelo seu ID.
func (m *CommentModel) Get(id int) (*Comment, error) {
	stmt := `SELECT id, snippet_id, author, content, created, updated, upvotes 
	         FROM comments WHERE id = ?`

	c := &Comment{}

	err := m.DB.QueryRow(stmt, id).Scan(&c.ID, &c.SnippetID, &c.Author, &c.Content, &c.Created, &c.Updated, &c.Upvotes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return c, nil
}
