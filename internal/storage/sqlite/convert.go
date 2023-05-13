package sqlite

import (
	"github.com/goserg/ratingserver/gen/model"
	"github.com/goserg/ratingserver/internal/domain"

	"github.com/google/uuid"
)

func convertPlayersToDomain(players []model.Players) []domain.Player {
	converted := make([]domain.Player, 0, len(players))
	for _, player := range players {
		p, err := convertPlayerToDomain(player)
		if err != nil {
			return nil
		}
		converted = append(converted, p)
	}
	return converted
}

func convertPlayerToDomain(player model.Players) (domain.Player, error) {
	id, err := uuid.Parse(player.ID)
	if err != nil {
		return domain.Player{}, err
	}
	return domain.Player{
		ID:           id,
		Name:         player.Name,
		RegisteredAt: player.CreatedAt,
	}, nil
}

func convertPlayerFromDomain(player domain.Player) model.Players {
	return model.Players{
		ID:        player.ID.String(),
		Name:      player.Name,
		CreatedAt: player.RegisteredAt,
	}
}

func convertMatchesToDomain(matches []model.Matches) ([]domain.Match, error) {
	converted := make([]domain.Match, 0, len(matches))
	for _, match := range matches {
		m, err := convertMatchToDomain(match)
		if err != nil {
			return nil, err
		}
		converted = append(converted, m)
	}
	return converted, nil
}

func convertMatchToDomain(match model.Matches) (domain.Match, error) {
	idA, err := uuid.Parse(match.PlayerA)
	if err != nil {
		return domain.Match{}, err
	}
	playerA := domain.Player{ID: idA}
	idB, err := uuid.Parse(match.PlayerB)
	if err != nil {
		return domain.Match{}, err
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
	return domain.Match{
		ID:      int(match.ID),
		PlayerA: playerA,
		PlayerB: playerB,
		Winner:  winner,
		Date:    match.CreatedAt,
	}, nil
}
