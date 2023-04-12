package service

import (
	"ratingserver/internal/domain"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func Test_calculateMatches(t *testing.T) {
	player1 := uuid.New()
	player2 := uuid.New()
	player3 := uuid.New()
	tests := []struct {
		name    string
		matches []domain.Match
		want    []domain.Match
	}{
		{
			name: "maybe working",
			matches: []domain.Match{
				{
					PlayerA: domain.Player{
						ID: player1,
					},
					PlayerB: domain.Player{
						ID: player2,
					},
					Winner: domain.Player{
						ID: player2,
					},
				},
				{
					PlayerA: domain.Player{
						ID: player1,
					},
					PlayerB: domain.Player{
						ID: player3,
					},
					Winner: domain.Player{
						ID: player3,
					},
				},
				{
					PlayerA: domain.Player{
						ID: player2,
					},
					PlayerB: domain.Player{
						ID: player3,
					},
					Winner: domain.Player{
						ID: player2,
					},
				},
			},
			want: []domain.Match{
				{
					PlayerA: domain.Player{
						ID:           player1,
						EloRating:    980,
						GamesPlayed:  1,
						RatingChange: -20,
					},
					PlayerB: domain.Player{
						ID:           player2,
						EloRating:    1020,
						GamesPlayed:  1,
						RatingChange: 20,
					},
					Winner: domain.Player{
						ID: player2,
					},
				},
				{
					PlayerA: domain.Player{
						ID:           player1,
						EloRating:    961,
						GamesPlayed:  2,
						RatingChange: -19,
					},
					PlayerB: domain.Player{
						ID:           player3,
						EloRating:    1019,
						GamesPlayed:  1,
						RatingChange: 19,
					},
					Winner: domain.Player{
						ID: player3,
					},
				},
				{
					PlayerA: domain.Player{
						ID:           player2,
						EloRating:    1040,
						GamesPlayed:  2,
						RatingChange: 20,
					},
					PlayerB: domain.Player{
						ID:           player3,
						EloRating:    999,
						GamesPlayed:  2,
						RatingChange: -20,
					},
					Winner: domain.Player{
						ID: player2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateMatches(tt.matches); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
