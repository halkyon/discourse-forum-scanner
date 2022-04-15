package main

import (
	"fmt"
	"net/url"
)

func validateNotEmpty(name, value string) error {
	if value == "" {
		return fmt.Errorf("flag %s is empty", name)
	}

	return nil
}

func validateURL(name, value string) error {
	if _, parseErr := url.ParseRequestURI(value); parseErr != nil {
		return fmt.Errorf("flag %s is invalid: %w", name, parseErr)
	}

	return nil
}
