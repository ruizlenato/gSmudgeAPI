package twitter

type TweetData struct {
	Data struct {
		ThreadedConversationWithInjectionsV2 struct {
			Instructions []Instruction `json:"instructions"`
		} `json:"threaded_conversation_with_injections_v2"`
	} `json:"data"`
}

type Instruction struct {
	Entries []Entry `json:"entries,omitempty"`
}

type Entry struct {
	EntryID string  `json:"entryId"`
	Content Content `json:"content"`
}

type Content struct {
	ItemContent ItemContent `json:"itemContent"`
}

type ItemContent struct {
	TweetResults TweetResults `json:"tweet_results"`
}

type TweetResults struct {
	Result Result `json:"result"`
}

type Result struct {
	Typename string `json:"__typename"`
	Tweet    Tweet  `json:"tweet"`
	Legacy   Legacy `json:"legacy"`
}
type Tweet struct {
	Legacy Legacy `json:"legacy"`
}

type Legacy struct {
	FullText         string           `json:"full_text"`
	ExtendedEntities ExtendedEntities `json:"extended_entities"`
}

type ExtendedEntities struct {
	Media []Media `json:"media"`
}

type Media struct {
	DisplayURL           string `json:"display_url"`
	ExpandedURL          string `json:"expanded_url"`
	Indices              []int  `json:"indices"`
	MediaURLHTTPS        string `json:"media_url_https"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	ExtMediaAvailability struct {
		Status string `json:"status"`
	} `json:"ext_media_availability"`
	Sizes        Sizes        `json:"sizes"`
	OriginalInfo OriginalInfo `json:"original_info"`
	VideoInfo    VideoInfo    `json:"video_info"`
}

type Sizes struct {
	Large Size `json:"large"`
	Thumb Size `json:"thumb"`
}

type Size struct {
	H      int    `json:"h"`
	W      int    `json:"w"`
	Resize string `json:"resize"`
}

type OriginalInfo struct {
	Height     int   `json:"height"`
	Width      int   `json:"width"`
	FocusRects []any `json:"focus_rects"`
}

type VideoInfo struct {
	AspectRatio    []int     `json:"aspect_ratio"`
	DurationMillis int       `json:"duration_millis"`
	Variants       []Variant `json:"variants"`
}

type Variant struct {
	Bitrate     int    `json:"bitrate,omitempty"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}
