package repository

import (

)

type CatalogItemRepository interface { //это интерфейс для работы с товарами
	Items(ctx context.Context)([]entities.CatalogItem, error) //это метод для получения всех товаров
}