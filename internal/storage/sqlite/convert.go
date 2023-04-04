package sqlite

import (
	"ratingserver/gen/model"
	"ratingserver/internal/domain"

	"github.com/google/uuid"
)

func convertPlayersToDomain(players []model.Players) []domain.Player {
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
		})
	}
	return converted
}

func convertPlayerFromDomain(player domain.Player) model.Players {
	return model.Players{
		ID:        player.ID.String(),
		Name:      player.Name,
		CreatedAt: player.RegisteredAt,
	}
}

func convertMatchesToDomain(matches []model.Matches) []domain.Match {
	converted := make([]domain.Match, 0, len(matches))
	for _, match := range matches {
		idA, err := uuid.Parse(match.PlayerA)
		if err != nil {
			return nil
		}
		playerA := domain.Player{ID: idA}
		idB, err := uuid.Parse(match.PlayerB)
		if err != nil {
			return nil
		}
		var winner domain.Player
		playerB := domain.Player{ID: idB}
		if match.Winner != nil && *match.Winner != uuid.Nil.String() {
			if *match.Winner == playerA.ID.String() {
				winner = playerA
			} else {
				winner = playerB
			}
		}
		converted = append(converted, domain.Match{
			ID:      int(match.ID),
			PlayerA: playerA,
			PlayerB: playerB,
			Winner:  winner,
			Date:    match.CreatedAt,
		})
	}
	return converted
}
