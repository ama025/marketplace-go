package shared

import "fmt"

// StatusNotFound — ошибка: ресурс не найден (аналог HTTP 404).
// Используется в репозиториях, когда запись не существует в БД.
//
// Пример:
//
//	return nil, shared.StatusNotFound{Resource: "CatalogItem", Key: id}
type StatusNotFound struct {
	Resource string // Название ресурса (например, "CatalogItem", "ShoppingCart")
	Key      any    // Ключ, по которому искали (UUID, строка и т.д.)
}

func (e StatusNotFound) Error() string {
	return fmt.Sprintf("%s with key '%v' not found", e.Resource, e.Key)
}

// StatusBadRequest — ошибка: некорректные входные данные (аналог HTTP 400).
// Используется при невалидных параметрах запроса.
//
// Пример:
//
//	return shared.StatusBadRequest{Message: "pageSize must be greater than 0"}
type StatusBadRequest struct {
	Message string // Описание проблемы
}

func (e StatusBadRequest) Error() string {
	return fmt.Sprintf("bad request: %s", e.Message)
}

// StatusConflict — ошибка: конфликт данных (аналог HTTP 409).
// Используется когда запись уже существует или нарушено уникальное ограничение.
//
// Пример:
//
//	return shared.StatusConflict{Resource: "Brand", Message: "title already exists"}
type StatusConflict struct {
	Resource string
	Message  string
}

func (e StatusConflict) Error() string {
	return fmt.Sprintf("conflict on %s: %s", e.Resource, e.Message)
}

// StatusUnauthorized — ошибка: не авторизован (аналог HTTP 401).
// Используется когда запрос выполнен без действительного токена.
type StatusUnauthorized struct {
	Message string
}

func (e StatusUnauthorized) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("unauthorized: %s", e.Message)
	}
	return "unauthorized"
}

// StatusForbidden — ошибка: доступ запрещён (аналог HTTP 403).
// Используется когда пользователь авторизован, но не имеет прав на операцию.
type StatusForbidden struct {
	Message string
}

func (e StatusForbidden) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("forbidden: %s", e.Message)
	}
	return "forbidden"
}

// IsNotFound — проверяет, является ли ошибка типом StatusNotFound.
// Удобно использовать в хендлерах для выбора HTTP-кода ответа.
//
// Пример:
//
//	if shared.IsNotFound(err) {
//	    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//	    return
//	}
func IsNotFound(err error) bool {
	_, ok := err.(StatusNotFound)
	return ok
}

// IsBadRequest — проверяет, является ли ошибка типом StatusBadRequest.
func IsBadRequest(err error) bool {
	_, ok := err.(StatusBadRequest)
	return ok
}

// IsConflict — проверяет, является ли ошибка типом StatusConflict.
func IsConflict(err error) bool {
	_, ok := err.(StatusConflict)
	return ok
}