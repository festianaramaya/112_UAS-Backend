package model

type Permission struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Resource    string `db:"resource"`
	Action      string `db:"action"`
	Description string `db:"description"`
}
