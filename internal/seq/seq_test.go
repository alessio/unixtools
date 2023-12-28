package seq_test

import (
	"fmt"
	"testing"

	"al.essio.dev/pkg/tools/internal/seq"
)

func Test_IntSequence(t *testing.T) {
	t.Parallel()
	type args struct {
		start int
		incr  uint
		end   int
		width uint
	}
	tests := []struct {
		name            string
		args            args
		wantItemsLen    int
		wantOutOfBounds bool
	}{
		{"5 to 10", args{5, 1, 10, 5}, 6, false},
		{"10 to 5", args{10, 1, 5, 5}, 6, false},
		{"0 to 100, out of bounds", args{0, 1, 100, 2}, 100, true},
		{"0 to 100, nil width", args{0, 1, 100, 0}, 101, false},
		{"0 to 100, nil width", args{0, 5, 100, 0}, 21, false},
		{"wrong args, out of bounds", args{10, 1, 20, 1}, 0, true},
		{"-5 to 5", args{-5, 1, 5, 2}, 11, false},
		{"5 to -5", args{5, 1, -5, 2}, 11, false},
		{"0", args{0, 1, 0, 1}, 1, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := seq.NewInt(tt.args.start, tt.args.incr, tt.args.end, tt.args.width)
			out := []string{}
			if s.WidthExceeded() {
				t.Fatal("width exceeded")
			}

			for i := range s.Items() {
				t.Logf("item = %s", i)
				out = append(out, i)
			}

			if tt.wantOutOfBounds != s.WidthExceeded() {
				t.Fatalf("wantOutOfBounds: want %v, got: %v", tt.wantOutOfBounds, s.WidthExceeded())
			}

			if tt.wantItemsLen != len(out) {
				t.Fatalf("wantItemsLen: want: %d, got: %d, sequence: %v", tt.wantItemsLen, len(out), out)
			}
		})
	}
}

func ExampleSequence_Items() {
	s := seq.NewInt(20, 5, 100, 3)
	for i := range s.Items() {
		fmt.Println(i)
	}
	// Output:
	// 020
	// 025
	// 030
	// 035
	// 040
	// 045
	// 050
	// 055
	// 060
	// 065
	// 070
	// 075
	// 080
	// 085
	// 090
	// 095
	// 100
}
