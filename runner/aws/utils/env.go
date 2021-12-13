package utils

import (
	"os"
)

func GetEnvVariable(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}
