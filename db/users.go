package db

import (
	"fmt"
	"strings"
)

type User struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ProfilePic string `json:"profilePic"`
}

func GetUser(userId int64) (*User, error) {
	rows := Db.QueryRow("SELECT * FROM users WHERE id = ?;", userId)

	var (
		id         int64
		name       string
		email      string
		profilePic string
	)

	err := rows.Scan(&id, &name, &email, &profilePic)
	if err != nil {
		return nil, err
	}

	return &User{
		Id:         id,
		Name:       name,
		Email:      email,
		ProfilePic: profilePic,
	}, nil
}

func GetAllUsers() ([]User, error) {
	users := []User{}

	rows, err := Db.Query("SELECT id, name, email, profilePic FROM users;")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			id         int64
			name       string
			email      string
			profilePic string
		)

		err := rows.Scan(&id, &name, &email, &profilePic)
		if err != nil {
			return nil, err
		}

		users = append(users, User{
			Id:         id,
			Name:       name,
			Email:      email,
			ProfilePic: profilePic,
		})
	}

	return users, nil
}

func CreateUser(u *User) (int64, error) {
	res, err := Db.Exec("INSERT INTO users (name, email, profilePic) VALUES (?, ?, ?) RETURNING id", u.Name, u.Email, u.ProfilePic)
	if err != nil {
		return 0, err
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func UpdateUser(u *User) error {
	q := "UPDATE users SET "

	if strings.TrimSpace(u.Name) != "" {
		q += fmt.Sprintf("name = \"%s\",", u.Name)
	}

	if strings.TrimSpace(u.Email) != "" {
		q += fmt.Sprintf("email = \"%s\",", u.Email)
	}

	if strings.TrimSpace(u.ProfilePic) != "" {
		q += fmt.Sprintf("profilePic = \"%s\",", u.ProfilePic)
	}

	q = strings.TrimRight(q, ",")

	q += fmt.Sprintf(" WHERE id = %d;", u.Id)

	_, err := Db.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(userId int64) error {
	_, err := Db.Exec("DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		return err
	}

	return nil
}

func CheckUserExists(userId int64) (bool, error) {
	row := Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ? LIMIT 1);", userId)

	var check bool

	err := row.Scan(&check)
	if err != nil {
		return false, err
	}

	return check, nil
}
