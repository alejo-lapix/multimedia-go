package options

import "github.com/alejo-lapix/multimedia-go/persistence"

type PageOption struct {
	// Name is the option identifier
	Name      string                      `json:"name"`
	Terms     string                      `json:"terms"`
	Wallpaper *persistence.MultimediaItem `json:"wallpaper"`
}

type PageOptionRepository interface {
	Store(option *PageOption) error
	FindByName(name string) (*PageOption, error)
}
