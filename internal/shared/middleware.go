// Package shared содержит общие утилиты и middleware, используемые всеми сервисами.
package shared

import (
	"log"      // Логирование ошибок и паник на сервере
	"net/http" // HTTP-коды статуса

	"github.com/gin-gonic/gin" // HTTP-фреймворк Gin
)

// ErrorHandleMiddleware — middleware для централизованной обработки ошибок и паник.
//
// Выполняет две задачи:
//  1. Recovery (восстановление после паники): если где-то в хендлере произошёл panic,
//     middleware перехватывает его, логирует и возвращает клиенту 500 Internal Server Error.
//     Без этого паника роняет весь сервер.
//
//  2. Error propagation (передача ошибок): хендлеры могут вызвать c.Error(err) вместо
//     c.JSON(...), а middleware после c.Next() сам сформирует правильный JSON-ответ.
//     Это позволяет хендлерам не думать о формате ошибок.
//
// Использование в main.go:
//
//	r := gin.New()                           // gin.New() — без встроенных middleware
//	r.Use(shared.ErrorHandleMiddleware())    // подключаем наш middleware
func ErrorHandleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer выполнится ПОСЛЕ того, как хендлер завершится (или упадёт с паникой).
		// recover() перехватывает панику и не даёт ей обрушить сервер.
		defer func() {
			if r := recover(); r != nil {
				// Логируем панику на сервере для отладки
				log.Printf("[PANIC RECOVERED] %v", r)

				// Возвращаем клиенту стандартный ответ 500
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		// c.Next() передаёт управление следующему middleware или хендлеру.
		// Код ПОСЛЕ Next() выполнится уже после того, как хендлер вернул ответ.
		c.Next()

		// --- Обработка ошибок, записанных хендлером через c.Error(err) ---
		// c.Errors — список всех ошибок, накопленных в процессе обработки запроса.
		// Хендлер может вызвать c.Error(err) и не возвращать ответ сам —
		// middleware сделает это централизованно.
		if len(c.Errors) > 0 {
			// Берём последнюю (самую релевантную) ошибку из цепочки
			err := c.Errors.Last()

			// Логируем ошибку для мониторинга
			log.Printf("[ERROR] %s %s → %v", c.Request.Method, c.Request.URL.Path, err)

			// Если хендлер уже записал статус — не перетираем его
			if c.Writer.Written() {
				return
			}

			// Возвращаем единый формат ошибки клиенту
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
}
