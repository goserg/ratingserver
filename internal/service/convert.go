package service

import (
	"ratingserver/gen/model"
	"ratingserver/internal/domain"

	"github.com/google/uuid"
)

func convertPlayers(players []model.Players) []domain.Player {
	converted := make([]domain.Player, 0, len(players))
	for _, player := range players {
		id, err := uuid.Parse(player.ID)
		if err != nil {
			return nil
		}
		converted = append(converted, domain.Player{
			ID:           id,
			Name:         player.Name,
			RegisteredAt: player.CreatedAt,
			EloRating:    777,
		})
	}
	return converted
}
