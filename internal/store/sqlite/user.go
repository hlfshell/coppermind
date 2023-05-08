package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	internal_users "github.com/hlfshell/coppermind/internal/users"
	"github.com/hlfshell/coppermind/pkg/users"
	"github.com/wissance/stringFormatter"
)

const userSelectAllColumns = `id, name, created_at, updated_at, password, reset_token, reset_token_attempts, reset_token_generated_at`
const userSelectColumns = `id, name, created_at, updated_at`
const userSelectAuthColumns = `id, password, reset_token, reset_token_attempts, reset_token_generated_at`

func (store *SqliteStore) CreateUser(user *users.User, password string) error {
	// We need to hash the password before writing it
	auth := internal_users.UserAuth{}
	err := auth.SetPassword(password)
	if err != nil {
		return err
	}

	query := `INSERT INTO {0} ({1}) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	query = stringFormatter.Format(query, USERS_TABLE, userSelectAllColumns)

	_, err = store.db.Exec(
		query,
		user.ID,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
		auth.Password,
		auth.ResetToken,
		auth.ResetTokenAttempts,
		auth.ResetTokenGeneratedAt,
	)
	return err
}

func (store *SqliteStore) GetUser(id string) (*users.User, error) {
	query := `SELECT {0} FROM {1} WHERE id = ?`

	query = stringFormatter.Format(query, userSelectColumns, USERS_TABLE)

	rows, err := store.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	users, err := store.sqlToUsers(rows)
	if err != nil {
		return nil, err
	} else if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

func (store *SqliteStore) GetUserAuth(id string) (*internal_users.UserAuth, error) {
	query := `SELECT {0} FROM {1} WHERE id = ?`

	query = stringFormatter.Format(query, userSelectAuthColumns, USERS_TABLE)

	rows, err := store.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	users, err := store.sqlToUserAuths(rows)
	if err != nil {
		return nil, nil
	} else if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

func (store *SqliteStore) SaveUserAuth(auth *internal_users.UserAuth) error {
	query := `UPDATE {0}
		SET
			password = ?,
			reset_token = ?,
			reset_token_attempts = ?,
			reset_token_generated_at = ?
		WHERE id = ?`

	query = stringFormatter.Format(query, USERS_TABLE)

	_, err := store.db.Exec(
		query,
		auth.Password,
		auth.ResetToken,
		auth.ResetTokenAttempts,
		auth.ResetTokenGeneratedAt,
		auth.ID,
	)
	return err
}

func (store *SqliteStore) GenerateUserPasswordResetToken(id string) (string, error) {
	auth, err := store.GetUserAuth(id)
	if err != nil {
		return "", err
	}
	if auth == nil {
		return "", fmt.Errorf("user doesn't exist")
	}

	auth.GenerateResetToken()

	err = store.SaveUserAuth(auth)
	if err != nil {
		return "", err
	}

	return auth.ResetToken, nil
}

func (store *SqliteStore) ResetPassword(id string, token string, password string) error {
	auth, err := store.GetUserAuth(id)
	if err != nil {
		return err
	}
	if auth == nil {
		return fmt.Errorf("user doesn't exist")
	}

	if auth.ResetToken != token {
		// Increment our attempts
		auth.ResetTokenAttempts += 1
		err = store.SaveUserAuth(auth)
		if err != nil {
			return err
		}
		return fmt.Errorf("invalid token")
	}

	auth.ResetToken = ""
	auth.ResetTokenAttempts = 0
	auth.ResetTokenGeneratedAt = time.Time{}

	err = auth.SetPassword(password)
	if err != nil {
		return err
	}

	return store.SaveUserAuth(auth)
}

func (store *SqliteStore) DeleteUser(id string) error {
	query := `DELETE FROM {0} WHERE id = ?`

	query = stringFormatter.Format(query, USERS_TABLE)

	_, err := store.db.Exec(query, id)
	return err
}

func (store *SqliteStore) sqlToUsers(rows *sql.Rows) ([]*users.User, error) {
	defer rows.Close()
	var foundUsers []*users.User
	for rows.Next() {
		user := &users.User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		foundUsers = append(foundUsers, user)
	}
	return foundUsers, nil
}

func (store *SqliteStore) sqlToUserAuths(rows *sql.Rows) ([]*internal_users.UserAuth, error) {
	defer rows.Close()
	var foundUsers []*internal_users.UserAuth
	for rows.Next() {
		user := &internal_users.UserAuth{}
		err := rows.Scan(
			&user.ID,
			&user.Password,
			&user.ResetToken,
			&user.ResetTokenAttempts,
			&user.ResetTokenGeneratedAt,
		)
		if err != nil {
			return nil, err
		}
		foundUsers = append(foundUsers, user)
	}
	return foundUsers, nil
}
