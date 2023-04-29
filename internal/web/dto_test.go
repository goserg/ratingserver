package web

import (
	"testing"

	"github.com/google/uuid"
)

func Test_createMatch_Validate(t *testing.T) {
	tests := []struct {
		name    string
		match   createMatch
		wantErr bool
	}{
		{
			name: "winA",
			match: createMatch{
				PlayerA: uuid.NameSpaceDNS,
				PlayerB: uuid.NameSpaceURL,
				Winner:  uuid.NameSpaceDNS,
			},
			wantErr: false,
		},
		{
			name: "winB",
			match: createMatch{
				PlayerA: uuid.NameSpaceDNS,
				PlayerB: uuid.NameSpaceURL,
				Winner:  uuid.NameSpaceURL,
			},
			wantErr: false,
		},
		{
			name: "draw",
			match: createMatch{
				PlayerA: uuid.NameSpaceDNS,
				PlayerB: uuid.NameSpaceURL,
				Winner:  uuid.Nil,
			},
			wantErr: false,
		},
		{
			name: "missing A",
			match: createMatch{
				PlayerA: uuid.Nil,
				PlayerB: uuid.NameSpaceURL,
				Winner:  uuid.NameSpaceURL,
			},
			wantErr: true,
		},
		{
			name: "missing B",
			match: createMatch{
				PlayerA: uuid.NameSpaceDNS,
				PlayerB: uuid.Nil,
				Winner:  uuid.NameSpaceDNS,
			},
			wantErr: true,
		},
		{
			name: "missing both",
			match: createMatch{
				PlayerA: uuid.Nil,
				PlayerB: uuid.Nil,
				Winner:  uuid.Nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.match.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
