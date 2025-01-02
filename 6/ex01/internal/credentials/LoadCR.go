package credentials

import (
	"bufio"
	"os"
	"strings"
)

type Credentials struct {
	AdminUsername string
	AdminPassword string
	DBUsername    string
	DBPassword    string
	DBName        string
}

var creds Credentials

func LoadCredentials(filename string) (Credentials, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Credentials{}, err
	}
	defer file.Close()

	creds := Credentials{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "admin_username":
			creds.AdminUsername = value
		case "admin_password":
			creds.AdminPassword = value
		case "db_username":
			creds.DBUsername = value
		case "db_password":
			creds.DBPassword = value
		case "db_name":
			creds.DBName = value
		}
	}
	if err := scanner.Err(); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}
