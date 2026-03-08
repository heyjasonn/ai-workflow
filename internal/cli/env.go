package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	lineNo := 0
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid .env line %d: missing '='", lineNo)
		}
		key := strings.TrimSpace(k)
		if key == "" {
			return fmt.Errorf("invalid .env line %d: empty key", lineNo)
		}

		val := strings.TrimSpace(v)
		if len(val) >= 2 {
			if (val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"') {
				val = val[1 : len(val)-1]
			}
		}

		if os.Getenv(key) == "" {
			if err := os.Setenv(key, val); err != nil {
				return err
			}
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	return nil
}
