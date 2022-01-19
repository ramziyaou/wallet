package domain

import "time"

type Claims struct {
	IIN       string
	Username  string
	createdAt string
	Iat       time.Duration
	Admin     *bool
}

func (c *Claims) Valid() error {
	return nil
}
