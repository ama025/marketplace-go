package entities

import "time"

// Discount — доменная модель скидки на товар.
// Описывает скидку в процентах, привязанную к конкретному товару каталога.
type Discount struct {
	ID        string     // UUID скидки
	ItemID    string     // UUID товара из каталога
	Percent   float64    // Процент скидки (0–100)
	Active    bool       // Активна ли скидка прямо сейчас
	StartsAt  *time.Time // Начало действия (nil = сразу)
	EndsAt    *time.Time // Конец действия (nil = бессрочно)
	CreatedAt time.Time  // Когда создана
	UpdatedAt time.Time  // Когда обновлена
}
