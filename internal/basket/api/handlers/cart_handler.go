// Package handlers — слой API (Handlers) корзины покупок.
// Принимает HTTP-запросы, валидирует данные и делегирует работу слою Application.
package handlers

import (
	"net/http" // Стандартные HTTP-коды статуса
	"time"

	"marketplace/internal/basket/domain"              // Доменные модели корзины
	"marketplace/internal/basket/domain/repositories" // Интерфейс репозитория
	"marketplace/internal/shared/messaging"           // RabbitMQ publisher

	"github.com/gin-gonic/gin" // HTTP-фреймворк Gin
)

// saveCartRequest — структура тела POST /cart запроса.
// Описывает корзину, которую клиент хочет сохранить.
type saveCartRequest struct {
	AccountName string                `json:"accountName" binding:"required"` // Имя аккаунта владельца корзины (обязательно)
	Items       []saveCartItemRequest `json:"items" binding:"required"`       // Список позиций корзины (обязательно)
}

// saveCartItemRequest — одна позиция корзины в теле запроса.
type saveCartItemRequest struct {
	ItemId    string  `json:"itemId" binding:"required"`           // UUID товара из каталога (обязательно)
	Quantity  int     `json:"quantity" binding:"required,min=1"`   // Количество — минимум 1 (обязательно)
	UnitPrice float64 `json:"unitPrice" binding:"required,min=0"`  // Цена за единицу — не может быть отрицательной
	ItemTitle *string `json:"itemTitle"`                           // Название товара (необязательно)
	ItemNote  *string `json:"itemNote"`                            // Заметка покупателя (необязательно)
}

// CartHandler — HTTP-обработчик для работы с корзиной покупок.
// Зависит от абстрактного интерфейса репозитория (Dependency Injection),
// а не от конкретной реализации PostgreSQL.
type CartHandler struct {
	repo      repositories.ShoppingCartRepository // Абстрактный репозиторий корзины
	publisher *messaging.Publisher                // RabbitMQ publisher (nil если не настроен)
}

// NewCartHandler — конструктор обработчика корзины.
// Принимает интерфейс репозитория, что позволяет легко подменить реализацию (тесты, in-memory и т.д.).
func NewCartHandler(repo repositories.ShoppingCartRepository) *CartHandler {
	return &CartHandler{repo: repo}
}

// NewCartHandlerWithPublisher — конструктор с RabbitMQ publisher для публикации событий заказа.
func NewCartHandlerWithPublisher(repo repositories.ShoppingCartRepository, publisher *messaging.Publisher) *CartHandler {
	return &CartHandler{repo: repo, publisher: publisher}
}

// SaveCart обрабатывает POST /cart — сохранение корзины покупок.
//
// Тело запроса (JSON):
//
//	{
//	  "accountName": "john_doe",
//	  "items": [
//	    { "itemId": "uuid", "quantity": 2, "unitPrice": 99.90, "itemTitle": "Кроссовки", "itemNote": "Размер 42" }
//	  ]
//	}
//
// Ответы:
//   - 204 No Content  — корзина успешно сохранена
//   - 400 Bad Request — невалидное тело запроса или некорректный UUID товара
//   - 500 Internal Server Error — ошибка базы данных
func (h *CartHandler) SaveCart(c *gin.Context) {
	// Читаем и валидируем тело запроса.
	// ShouldBindJSON проверяет теги binding:"required" и возвращает ошибку, если поле отсутствует.
	var req saveCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Конвертируем DTO запроса в доменную модель ShoppingCart
	cart := domain.ShoppingCart{
		AccountName: req.AccountName,
		Items:       make([]domain.ShoppingCartItem, 0, len(req.Items)),
	}

	for _, reqItem := range req.Items {
		// Парсим UUID товара — проверяем корректность формата
		itemID, err := parseUUID(reqItem.ItemId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid itemId: " + reqItem.ItemId,
			})
			return
		}

		// Добавляем позицию в доменную модель корзины
		cart.Items = append(cart.Items, domain.ShoppingCartItem{
			ItemId:    itemID,
			Quantity:  reqItem.Quantity,
			UnitPrice: reqItem.UnitPrice,
			ItemTitle: reqItem.ItemTitle,
			ItemNote:  reqItem.ItemNote,
		})
	}

	// Сохраняем корзину через репозиторий (Create or Update — логика внутри репозитория)
	if err := h.repo.Save(c.Request.Context(), &cart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 204 No Content — стандартный ответ для успешных операций сохранения без возврата тела
	c.Status(http.StatusNoContent)
}

// GetCart обрабатывает GET /cart/:accountName — получение корзины покупок по имени аккаунта.
//
// Параметры URL:
//   - accountName (string) — уникальное имя аккаунта владельца корзины
//
// Ответы:
//   - 200 OK            — корзина найдена, возвращается в теле ответа (JSON)
//   - 404 Not Found     — корзина с таким accountName не существует
//   - 500 Internal Server Error — ошибка базы данных
func (h *CartHandler) GetCart(c *gin.Context) {
	// Извлекаем accountName из параметра URL (:accountName)
	accountName := c.Param("accountName")

	// Запрашиваем корзину из репозитория
	cart, err := h.repo.Get(c.Request.Context(), accountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Если репозиторий вернул nil — корзина не найдена (новый пользователь)
	if cart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart not found for account: " + accountName})
		return
	}

	// Возвращаем корзину вместе с подсчитанной итоговой стоимостью
	c.JSON(http.StatusOK, gin.H{
		"accountName": cart.AccountName,
		"items":       cart.Items,
		"totalPrice":  cart.TotalPrice(), // Вычисляется в доменной модели: sum(quantity * unitPrice)
	})
}

// DeleteCart обрабатывает DELETE /cart/:accountName — удаление корзины покупок.
//
// Параметры URL:
//   - accountName (string) — уникальное имя аккаунта владельца корзины
//
// Ответы:
//   - 204 No Content     — корзина успешно удалена
//   - 500 Internal Server Error — ошибка хранилища
func (h *CartHandler) DeleteCart(c *gin.Context) {
	accountName := c.Param("accountName")

	if err := h.repo.Delete(c.Request.Context(), accountName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 204 No Content — успешное удаление без тела ответа
	c.Status(http.StatusNoContent)
}

// CheckoutCart обрабатывает POST /cart/:accountName/checkout — оформление заказа.
//
// Что происходит:
//  1. Достаём корзину из Redis
//  2. Публикуем событие OrderConfirmed в RabbitMQ → Checkout создаёт заказ
//  3. Удаляем корзину (она уже оформлена)
//
// Ответы:
//   - 202 Accepted       — событие принято, заказ создаётся асинхронно
//   - 404 Not Found      — корзина пуста или не существует
//   - 500 Internal Server Error — ошибка публикации или удаления
func (h *CartHandler) CheckoutCart(c *gin.Context) {
	accountName := c.Param("accountName")

	// 1. Достаём корзину из Redis
	cart, err := h.repo.Get(c.Request.Context(), accountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cart == nil || len(cart.Items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart is empty or not found"})
		return
	}

	// 2. Публикуем событие в RabbitMQ (если publisher настроен)
	if h.publisher != nil {
		event := messaging.OrderConfirmedEvent{
			AccountName: accountName,
			ConfirmedAt: time.Now(),
			Items:       make([]messaging.OrderConfirmedItem, 0, len(cart.Items)),
		}

		for _, item := range cart.Items {
			title := ""
			if item.ItemTitle != nil {
				title = *item.ItemTitle
			}
			event.Items = append(event.Items, messaging.OrderConfirmedItem{
				ItemID:    item.ItemId.String(),
				ItemTitle: title,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Discount:  0, // TODO: запросить у promotion-сервиса
			})
		}

		if err := h.publisher.PublishOrderConfirmed(c.Request.Context(), event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish order: " + err.Error()})
			return
		}
	}

	// 3. Удаляем корзину — она уже оформлена
	if err := h.repo.Delete(c.Request.Context(), accountName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 202 Accepted — заказ принят и создаётся асинхронно через Checkout
	c.JSON(http.StatusAccepted, gin.H{
		"message":     "order accepted, processing asynchronously",
		"accountName": accountName,
	})
}
