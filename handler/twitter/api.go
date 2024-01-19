package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"gSmudgeAPI/cache"
	"gSmudgeAPI/handler"
	"gSmudgeAPI/utils"
	"log"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func TwitterIndexer(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	if len(url) == 0 {
		errorMessage := "No URL specified"
		ctx.Error(errorMessage, fasthttp.StatusMethodNotAllowed)
		return
	}

	TweetID := (regexp.MustCompile((`.*(?:twitter|x).com/.+status/([A-Za-z0-9]+)`))).FindStringSubmatch(url)[1]
	csrfToken := strings.ReplaceAll((uuid.New()).String(), "-", "")
	Headers := map[string]string{
		"Authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"Cookie":                    fmt.Sprintf("auth_token=ee4ebd1070835b90a9b8016d1e6c6130ccc89637; ct0=%v; ", csrfToken),
		"x-twitter-active-user":     "yes",
		"x-twitter-auth-type":       "OAuth2Session",
		"x-twitter-client-language": "en",
		"x-csrf-token":              csrfToken,
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	}
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

	Query := map[string]string{
		"variables":    string(variablesJson),
		"features":     string(featuresJson),
		"fieldToggles": string(fieldTogglesJson),
	}

	body := utils.GetHTTPRes("https://twitter.com/i/api/graphql/NmCeCgkVlsRGS1cAwqtgmw/TweetDetail", utils.RequestParams{Query: Query, Headers: Headers}).Body()

	var tweetData TweetData
	err := json.Unmarshal(body, &tweetData)
	if err != nil {
		log.Println(err)
	}

	var tweetResult interface{}
	indexedMedia := &handler.IndexedMedia{}

	for _, entry := range tweetData.Data.ThreadedConversationWithInjectionsV2.Instructions[0].Entries {
		if entry.EntryID == fmt.Sprintf("tweet-%v", TweetID) {
			if entry.Content.ItemContent.TweetResults.Result.Typename == "TweetWithVisibilityResults" {
				tweetResult = entry.Content.ItemContent.TweetResults.Result.Tweet.Legacy
			} else {
				tweetResult = entry.Content.ItemContent.TweetResults.Result.Legacy
			}
			break
		}
	}

	var caption string
	if tweet, ok := tweetResult.(Legacy); ok {
		caption = tweet.FullText
	}

	for _, media := range tweetResult.(Legacy).ExtendedEntities.Media {
		var videoType string
		if slices.Contains([]string{"animated_gif", "video"}, media.Type) {
			videoType = "video"
		}
		if videoType != "video" {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Height: media.OriginalInfo.Height,
				Width:  media.OriginalInfo.Width,
				Source: media.MediaURLHTTPS,
				Video:  false,
			})
		} else {
			indexedMedia.Medias = append(indexedMedia.Medias, handler.Medias{
				Height: media.OriginalInfo.Height,
				Width:  media.OriginalInfo.Width,
				Source: media.VideoInfo.Variants[0].URL,
				Video:  true,
			})
		}
	}
	ixt := handler.IndexedMedia{
		URL:     url,
		Medias:  indexedMedia.Medias,
		Caption: caption}

	jsonResponse, _ := json.Marshal(ixt)
	err = cache.GetRedisClient().Set(context.Background(), TweetID, jsonResponse, 24*time.Hour*60).Err()
	if err != nil {
		log.Println("Error setting cache:", err)
	}
	ctx.Response.Header.Add("Content-Type", "application/json")
	json.NewEncoder(ctx).Encode(ixt)
}
