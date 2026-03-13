package pocketmcp

import (
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

const defaultPocketBaseURL = "http://localhost:8090"

type pocketBaseCredentials struct {
	URL      string
	Email    string
	Password string
}

func promptPocketBaseCredentials(cmd *cobra.Command, url string, email string, password string) (pocketBaseCredentials, error) {
	_ = cmd

	creds := pocketBaseCredentials{
		URL:      strings.TrimSpace(url),
		Email:    strings.TrimSpace(email),
		Password: strings.TrimSpace(password),
	}

	askOpts := []survey.AskOpt{
		survey.WithStdio(os.Stdin, os.Stdout, os.Stderr),
	}

	if creds.URL == "" {
		if err := survey.AskOne(
			&survey.Input{
				Message: "PocketBase URL:",
				Default: firstNonEmpty(os.Getenv("POCKETBASE_URL"), defaultPocketBaseURL),
			},
			&creds.URL,
			append(askOpts, survey.WithValidator(survey.Required))...,
		); err != nil {
			return pocketBaseCredentials{}, err
		}
	}

	if creds.Email == "" {
		if err := survey.AskOne(
			&survey.Input{
				Message: "PocketBase user/email:",
				Default: strings.TrimSpace(os.Getenv("POCKETBASE_EMAIL")),
			},
			&creds.Email,
			append(askOpts, survey.WithValidator(survey.Required))...,
		); err != nil {
			return pocketBaseCredentials{}, err
		}
	}

	if creds.Password == "" {
		if err := survey.AskOne(
			&survey.Password{
				Message: "PocketBase password:",
			},
			&creds.Password,
			append(askOpts, survey.WithValidator(survey.Required))...,
		); err != nil {
			return pocketBaseCredentials{}, err
		}
	}

	return creds, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
