package repository

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
)

type CatalogItemRepository interface { //это интерфейс для работы с товарами
	Items(ctx context.Context) ([]entities.CatalogItem, error) //это метод для получения всех товаров
}
