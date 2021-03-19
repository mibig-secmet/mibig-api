package postgres

import (
	"database/sql"
	"errors"
	"time"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/utils"
)

type TokenModel struct {
	DB *sql.DB
}

func (t *TokenModel) Generate(email string) (*models.Token, error) {
	randomString, err := utils.GenerateUid(32)
	if err != nil {
		return nil, err
	}

	token := models.Token{
		Token:   randomString,
		Expires: time.Now().Add(time.Hour * 24),
	}

	tx, err := t.DB.Begin()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow("SELECT user_id FROM users WHERE email = ?", email)
	err = row.Scan(&token.UserId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO tokens VALUES(?, ?, ?)", token.Token, token.UserId, token.Expires.Unix())
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (t *TokenModel) Validate(userId int, token string) (bool, error) {
	var dbToken models.Token
	var unixTime int64
	row := t.DB.QueryRow("SELECT token_id, user_id, expires FROM tokens WHERE token_id = ? AND user_id = ?", token, userId)

	err := row.Scan(&dbToken.Token, &dbToken.UserId, &unixTime)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return false, nil
		}
		return false, err
	}
	dbToken.Expires = time.Unix(unixTime, 0)

	return time.Now().Before(dbToken.Expires), nil
}

func (t *TokenModel) Remove(token string) error {
	_, err := t.DB.Exec(`DELETE FROM tokens WHERE token_id = ?`, token)
	if err != nil {
		return err
	}

	return nil
}

func (t *TokenModel) Expire() error {
	_, err := t.DB.Exec(`DELETE FROM tokens WHERE expires < strftime('%s', 'now')`)
	if err != nil {
		return err
	}

	return nil
}
