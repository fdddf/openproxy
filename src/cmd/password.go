package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

func newPasswordCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "password <plaintext>",
		Short: "Generate a bcrypt hash for the provided password",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, err := generatePasswordHash(args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), hash)
			return nil
		},
	}
}

func generatePasswordHash(plaintext string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate bcrypt hash: %w", err)
	}

	return string(hash), nil
}
