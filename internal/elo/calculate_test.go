package elo

import "testing"

func TestCalculate(t *testing.T) {
	type args struct {
		Ra int
		Rb int
		K  int
		Sa Points
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "same rating draw",
			args: args{
				Ra: 1000,
				Rb: 1000,
				K:  40,
				Sa: Draw,
			},
			want: 1000,
		},
		{
			name: "same rating win",
			args: args{
				Ra: 1000,
				Rb: 1000,
				K:  40,
				Sa: Win,
			},
			want: 1020,
		},
		{
			name: "same rating lose",
			args: args{
				Ra: 1000,
				Rb: 1000,
				K:  40,
				Sa: Lose,
			},
			want: 980,
		},
		{
			name: "top rating draw",
			args: args{
				Ra: 1100,
				Rb: 1000,
				K:  40,
				Sa: Draw,
			},
			want: 1094,
		},
		{
			name: "top rating win",
			args: args{
				Ra: 1100,
				Rb: 1000,
				K:  40,
				Sa: Win,
			},
			want: 1114,
		},
		{
			name: "top rating lose",
			args: args{
				Ra: 1100,
				Rb: 1000,
				K:  40,
				Sa: Lose,
			},
			want: 1074,
		},
		{
			name: "bottom rating draw",
			args: args{
				Ra: 1000,
				Rb: 1100,
				K:  40,
				Sa: Draw,
			},
			want: 1006,
		},
		{
			name: "bottom rating win",
			args: args{
				Ra: 1000,
				Rb: 1100,
				K:  40,
				Sa: Win,
			},
			want: 1026,
		},
		{
			name: "bottom rating lose",
			args: args{
				Ra: 1000,
				Rb: 1100,
				K:  40,
				Sa: Lose,
			},
			want: 986,
		},
		{
			name: "close rating draw",
			args: args{
				Ra: 944,
				Rb: 938,
				K:  40,
				Sa: Draw,
			},
			want: 944,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Calculate(tt.args.Ra, tt.args.Rb, tt.args.K, tt.args.Sa); got != tt.want {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
