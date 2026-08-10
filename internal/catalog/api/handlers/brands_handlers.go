package handlers // Слой API Handlers (Контроллеры): отвечает за прием HTTP-запросов и отправку ответов клиенту

import (
	"net/http" // Стандартный пакет Go с константами статусов HTTP (http.StatusOK, http.StatusInternalServerError)

	"marketplace/internal/catalog/application/queries" // Импортируем слой бизнес-логики (Application Queries)

	"github.com/gin-gonic/gin" // Фреймворк Gin
)

// BrandsHandler — структура HTTP-контроллера для работы с брендами.
// Хранит в себе ссылку на объект бизнес-логики (queries.BrandsHandler).
type BrandsHandler struct {
	brands *queries.BrandsHandler
}

// NewBrandsHandler — конструктор для создания экземпляра BrandsHandler.
// Внедряет зависимость (Dependency Injection) слоя приложений.
func NewBrandsHandler(brands *queries.BrandsHandler) *BrandsHandler {
	return &BrandsHandler{brands: brands}
}

// Brands — метод-обработчик (HTTP Endpoint) для получения списка всех брендов.
// Вызывается автоматически веб-фреймворком Gin при запросе GET /api/v1/brands.
//
// Параметры:
//   - c (*gin.Context): контекст запроса Gin, содержит заголовки, параметры и методы для отправки ответа (c.JSON).
func (h *BrandsHandler) Brands(c *gin.Context) {
	// c.Request.Context() — берем контекст отмены/таймаута текущего HTTP-запроса.
	// h.brands.Handle(...) — вызов бизнес-логики (получение бренда из репозитория/БД).
	brands, err := h.brands.Handle(c.Request.Context())

	// Проверяем, произошла ли ошибка при выполнении бизнес-логики или запроса к БД
	if err != nil {
		// Ошибка 500 (Internal Server Error): возвращаем клиенту JSON с текстом ошибки {"error": "..."}
		// gin.H — это удобный псевдоним в Gin для map[string]any
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // Прерываем дальше выполнение метода
	}

	// Успешный ответ 200 (OK): отправляем клиенту массив брендов в формате JSON {"data": [...]}
	c.JSON(http.StatusOK, gin.H{"data": brands})
}
