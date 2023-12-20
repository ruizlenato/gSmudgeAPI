package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func rgraphql(TweetID string) []byte {
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/NmCeCgkVlsRGS1cAwqtgmw/TweetDetail", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	csrfToken := strings.ReplaceAll((uuid.New()).String(), "-", "")
	req.Header.Add("Cookie", fmt.Sprintf("auth_token=ee4ebd1070835b90a9b8016d1e6c6130ccc89637; ct0=%v; ", csrfToken))
	req.Header.Add("x-twitter-active-user", "yes")
	req.Header.Add("x-twitter-auth-type", "OAuth2Session")
	req.Header.Add("x-twitter-client-language", "en")
	req.Header.Add("x-csrf-token", csrfToken)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0")
	query := req.URL.Query()
	variables := map[string]interface{}{
		"focalTweetId":                           TweetID,
		"referrer":                               "messages",
		"includePromotedContent":                 true,
		"withCommunity":                          true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withBirdwatchNotes":                     true,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	}
	features := map[string]interface{}{
		"rweb_lists_timeline_redesign_enabled":                                    true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              false,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}
	fieldtoggles := map[string]interface{}{
		"withAuxiliaryUserLabels":     false,
		"withArticleRichContentState": false,
	}

	variablesJson, _ := json.Marshal(variables)
	featuresJson, _ := json.Marshal(features)
	fieldTogglesJson, _ := json.Marshal(fieldtoggles)
	query.Add("variables", string(variablesJson))
	query.Add("features", string(featuresJson))
	query.Add("fieldToggles", string(fieldTogglesJson))
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func TwitterIndexer(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		response := "No URL specified"
		http.Error(w, response, http.StatusMethodNotAllowed)
		return
	}

	TweetID := (regexp.MustCompile((`.*(twitter|x).com/.+status/([A-Za-z0-9]+)`))).FindStringSubmatch(url)[2]
	s := gjson.ParseBytes(rgraphql(TweetID)).String()
	indexedMedia := &handler.IndexedMedia{}
	var caption string
	results := gjson.Get(s, fmt.Sprintf(`data.threaded_conversation_with_injections_v2.instructions.0.entries.#(entryId="tweet-%v").content.itemContent.tweet_results.result`, string(TweetID)))
	if results.Get("__typename").String() == "TweetWithVisibilityResults" {
		results = results.Get("tweet")
	}
	caption = results.Get("legacy.full_text").String()

	medias := results.Get("legacy.extended_entities.media")
	for _, media := range medias.Array() {
		var videoType string
		for _, value := range []string{"animated_gif", "video"} {
			if strings.Contains(media.Get("type").String(), value) {
				videoType = "video"
			}
		}

		if videoType != "video" {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Width:  int(media.Get("original_info.width").Int()),
				Height: int(media.Get("original_info.height").Int()),
				Source: media.Get("media_url_https").String(),
				Video:  false,
			})
		} else {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Width:  int(media.Get("original_info.width").Int()),
				Height: int(media.Get("original_info.height").Int()),
				Source: media.Get("video_info.variants.0.url").String(),
				Video:  true,
			})
		}
	}

	ixt := handler.IndexedMedia{
		URL:     url,
		Medias:  indexedMedia.Medias,
		Caption: caption}

	jsonResponse, _ := json.Marshal(ixt)
	err := cache.GetRedisClient().Set(context.Background(), r.RequestURI, jsonResponse, 24*time.Hour*60).Err()
	if err != nil {
		log.Println("Error setting cache:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ixt)

}
