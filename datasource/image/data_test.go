package image

import (
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/require"
)

func TestFilterImages(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		images        []godo.Image
		expectedImage godo.Image
		expectedError string
	}{
		{
			name:   "by name - single match",
			config: &Config{Name: "test-image"},
			images: []godo.Image{
				{ID: 1, Name: "test-image"},
				{ID: 2, Name: "test-image-01"},
			},
			expectedImage: godo.Image{ID: 1, Name: "test-image"},
		},
		{
			name:   "by name - multiple matches",
			config: &Config{Name: "test-image"},
			images: []godo.Image{
				{ID: 1, Name: "test-image"},
				{ID: 2, Name: "test-image"},
			},
			expectedError: "More than one matching image found:",
		},
		{
			name:   "by name - multiple matches - latest",
			config: &Config{Name: "test-image", Latest: true},
			images: []godo.Image{
				{ID: 1, Name: "test-image", Created: "2022-08-08T21:31:54Z"},
				{ID: 2, Name: "test-image", Created: "2022-08-10T21:31:54Z"},
			},
			expectedImage: godo.Image{ID: 2, Name: "test-image", Created: "2022-08-10T21:31:54Z"},
		},
		{
			name:   "by name - multiple matches - region filter",
			config: &Config{Name: "test-image", Region: "nyc3", Latest: true},
			images: []godo.Image{
				{ID: 1, Name: "test-image", Created: "2022-08-08T21:31:54Z", Regions: []string{"nyc3"}},
				{ID: 2, Name: "test-image", Created: "2022-08-10T21:31:54Z", Regions: []string{"nyc2"}},
			},
			expectedImage: godo.Image{ID: 1, Name: "test-image", Created: "2022-08-08T21:31:54Z", Regions: []string{"nyc3"}},
		},
		{
			name:   "by name - no matches",
			config: &Config{Name: "test-image"},
			images: []godo.Image{
				{ID: 1, Name: "test-image-01", Created: "2022-08-08T21:31:54Z", Regions: []string{"nyc3"}},
				{ID: 2, Name: "test-image-02", Created: "2022-08-10T21:31:54Z", Regions: []string{"nyc2"}},
			},
			expectedError: "No image matching found",
		},

		{
			name:   "regex - single match",
			config: &Config{NameRegex: "test-image-.*"},
			images: []godo.Image{
				{ID: 1, Name: "test-image"},
				{ID: 2, Name: "test-image-01"},
			},
			expectedImage: godo.Image{ID: 2, Name: "test-image-01"},
		},
		{
			name:   "regex - multiple matches",
			config: &Config{NameRegex: "test-image-.*"},
			images: []godo.Image{
				{ID: 1, Name: "test-image-01"},
				{ID: 2, Name: "test-image-02"},
			},
			expectedError: "More than one matching image found:",
		},
		{
			name:   "regex - multiple matches - latest",
			config: &Config{NameRegex: "test-image-.*", Latest: true},
			images: []godo.Image{
				{ID: 1, Name: "test-image-01", Created: "2022-08-08T21:31:54Z"},
				{ID: 2, Name: "test-image-02", Created: "2022-08-10T21:31:54Z"},
			},
			expectedImage: godo.Image{ID: 2, Name: "test-image-02", Created: "2022-08-10T21:31:54Z"},
		},
		{
			name:   "regex - multiple matches - region filter",
			config: &Config{NameRegex: "test-image-.*", Region: "nyc3", Latest: true},
			images: []godo.Image{
				{ID: 1, Name: "test-image-01", Created: "2022-08-08T21:31:54Z", Regions: []string{"nyc3"}},
				{ID: 2, Name: "test-image-02", Created: "2022-08-10T21:31:54Z", Regions: []string{"nyc2"}},
			},
			expectedImage: godo.Image{ID: 1, Name: "test-image-01", Created: "2022-08-08T21:31:54Z", Regions: []string{"nyc3"}},
		},
		{
			name:   "regex - no matches",
			config: &Config{NameRegex: "test-image-.*"},
			images: []godo.Image{
				{ID: 1, Name: "test-image01", Created: "2022-08-08T21:31:54Z", Regions: []string{"nyc3"}},
				{ID: 2, Name: "test-image02", Created: "2022-08-10T21:31:54Z", Regions: []string{"nyc2"}},
			},
			expectedError: "No image matching found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := filterImages(tt.config, tt.images)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedImage, out)
			} else {
				require.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
