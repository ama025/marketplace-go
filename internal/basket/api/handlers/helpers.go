package handlers

import "github.com/google/uuid"

// parseUUID парсит строку в uuid.UUID.
// Возвращает ошибку, если строка не является валидным UUID.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
