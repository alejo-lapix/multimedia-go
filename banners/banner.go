package banners

import "github.com/alejo-lapix/multimedia-go/persistence"

type Banner struct {
	Background  string                     `json:"background"`
	Multimedia  persistence.MultimediaItem `json:"multimedia"`
	HtmlContent string                     `json:"htmlContent"`
}
