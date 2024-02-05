package instagram

type InstagramData struct {
	ShortcodeMedia    ShortcodeMedia `json:"shortcode_media"`
	XDTShortcodeMedia ShortcodeMedia `json:"xdt_shortcode_media"`
}

type ShortcodeMedia struct {
	Typename              string                `json:"__typename"`
	ID                    string                `json:"id"`
	Shortcode             string                `json:"shortcode"`
	Dimensions            Dimensions            `json:"dimensions"`
	DisplayResources      []DisplayResources    `json:"display_resources"`
	IsVideo               bool                  `json:"is_video"`
	Title                 string                `json:"title"`
	VideoURL              string                `json:"video_url"`
	EdgeMediaToCaption    EdgeMediaToCaption    `json:"edge_media_to_caption"`
	EdgeSidecarToChildren EdgeSidecarToChildren `json:"edge_sidecar_to_children"`
}

type Dimensions struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type DisplayResources struct {
	ConfigWidth  int    `json:"config_width"`
	ConfigHeight int    `json:"config_height"`
	Src          string `json:"src"`
}

type EdgeMediaToCaption struct {
	Edges []struct {
		Node struct {
			Text string `json:"text"`
		} `json:"node"`
	} `json:"edges"`
}

type EdgeSidecarToChildren struct {
	Edges []struct {
		Node struct {
			Typename         string             `json:"__typename"`
			ID               string             `json:"id"`
			Shortcode        string             `json:"shortcode"`
			CommenterCount   int                `json:"commenter_count"`
			Dimensions       Dimensions         `json:"dimensions"`
			DisplayResources []DisplayResources `json:"display_resources"`
			IsVideo          bool               `json:"is_video"`
		} `json:"node"`
	} `json:"edges"`
}
