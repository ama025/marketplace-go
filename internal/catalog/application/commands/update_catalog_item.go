package commands

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"

	"github.com/google/uuid"
)

type UpdateCatalogItemCommand struct {
	Id               uuid.UUID          `json:"id" validate:"required"`
	Title            string             `json:"title" validate:"required"`
	ShortDescription string             `json:"short_description"`
	FullDescription  string             `json:"full_description"`
	ImageURL         string             `json:"image_url"`
	BrandID          *entities.Brand    `json:"brand_id"`
	CategoryID       *entities.Category `json:"category_id"`
	Price            float64            `json:"price"`
}

type UpdateCatalogItemHandler struct {
	repo repositories.CatalogItemRepository
}

func NewUpdateCatalogItemHandler(repo repositories.CatalogItemRepository) *UpdateCatalogItemHandler {
	return &UpdateCatalogItemHandler{repo: repo}
}

func (h *UpdateCatalogItemHandler) Handle(ctx context.Context, command UpdateCatalogItemCommand) (bool, error) {

	existing, err := h.repo.Item(ctx, command.Id)
	if err != nil {
		return false, err
	}

	if existing == nil {
		return false, nil
	}

	item := entities.CatalogItem{
		BaseEntity: entities.BaseEntity{
			Id:    command.Id,
			Title: command.Title,
		},
		ShortDescription: command.ShortDescription,
		FullDescription:  command.FullDescription,
		ImageURL:         command.ImageURL,
		Brand:            command.BrandID,
		Category:         command.CategoryID,
		Price:            command.Price,
	}

	return h.repo.Update(ctx, &item)

}
