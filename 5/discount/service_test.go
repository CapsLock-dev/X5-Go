package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePercent(t *testing.T) {
	tests := []struct {
		name string
		discountPercent int
		wantError error 
	} {
		{"0% discount", 0, nil},
		{"10% discount", 10, nil},
		{"100% discount", 100, nil},
		{"negative discount", -1, ErrInvalidPercent},
		{"101% discount", 101, ErrInvalidPercent},
	}

	for _,tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePercent(tt.discountPercent)	
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name string
		price int
		discountPercent int
		wantValue int
		wantError error 
	} {
		{"regular discount", 1000, 15, 850, nil},
		{"0% discount", 1000, 0, 1000, nil},
		{"100% discount", 1000, 100, 0, nil},
		{"negative price", -1, 15, 0, ErrNegativePrice},
		{"101% discount", 1000, 101, 0, ErrInvalidPercent},
	}

	for _,tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Apply(tt.price, tt.discountPercent)	
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, got, tt.wantValue)
			}
		})
	}
}

func TestFinalPrice(t *testing.T) {
	tests := []struct {
		name string
		items []Item
		wantValue int
		wantError error 
	} {
		{"empty list", []Item{}, 0, nil},
		{"correct list", []Item{{1000, 10}, {1000, 100}, {1000, 15}}, 1750, nil},
		{"incorrect price list", []Item{{1000, 10}, {1000, 100}, {-1000, 15}}, 0, ErrNegativePrice},
		{"incorrect discount list", []Item{{1000, 10}, {1000, 101}, {1000, 15}}, 0, ErrInvalidPercent},
	}

	for _,tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := FinalPrice(tt.items)	
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, got, tt.wantValue)
			}
		})
	}
}
