package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"runtime"

	"github.com/joho/godotenv"
)

func get_os() string {
	return runtime.GOOS
}

func IsLinux() bool {
	return get_os() == "linux"
}
func IsWindows() bool {
	return get_os() == "windows"
}
func IsMac() bool {
	return get_os() == "darwin"
}
func GetEnvVariable(key string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}
	return os.Getenv(key), nil
}

func ReadCsvFile(fileName string) ([][]string, error) {

	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", fileName, err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ',' // Asegurarse de que el separador sea coma
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file %s: %w", fileName, err)
	}

	return rows, nil
}
