package commands

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"

	"github.com/google/uuid"
)

type CreateCatalogItemCommand struct {
	Title            string             `json:"title" validate:"required"`
	ShortDescription string             `json:"short_description"`
	FullDescription  string             `json:"full_description"`
	ImageURL         string             `json:"image_url"`
	BrandID          *entities.Brand    `json:"brand_id"`
	CategoryID       *entities.Category `json:"category_id"`
	Price            float64            `json:"price"`
}

type CreateCatalogItemHandler struct {
	repo repositories.CatalogItemRepository
}

func NewCreateCatalogItemHandler(repo repositories.CatalogItemRepository) *CreateCatalogItemHandler {
	return &CreateCatalogItemHandler{repo: repo}
}

func (h *CreateCatalogItemHandler) Handle(ctx context.Context, command CreateCatalogItemCommand) (uuid.UUID, error) {
	item := entities.CatalogItem{
		BaseEntity: entities.BaseEntity{
			Id:    uuid.New(),
			Title: command.Title,
		},
		ShortDescription: command.ShortDescription,
		FullDescription:  command.FullDescription,
		ImageURL:         command.ImageURL,
		Brand:            command.BrandID,
		Category:         command.CategoryID,
		Price:            command.Price,
	}

	created, err := h.repo.Create(ctx, &item)
	if err != nil {
		return uuid.Nil, err
	}

	return created.Id, nil

}
