package crawler

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const CrawlInitialDelay = 0
const CrawlMaximumDelay = 5000 * time.Millisecond

type Crawler struct {
	config *Config

	userId     string
	queryId    string
	rhxGis     string
	sharedData sharedDataJsonType

	store      *ResourceStore
	wait       <-chan time.Time
	crawlDelay time.Duration
}

type ResourceStore struct {
	sync.Mutex
	resources []Resource
}

var pageChan = make(chan page, 1000)
var resourceChan = make(chan Resource, 10000)
var galleryPageChan = make(chan Resource, 1000)
var videoPageChan = make(chan Resource, 1000)

func FetchProfileImage(config *Config) (string, error) {
	crawler := &Crawler{
		config:     config,
		crawlDelay: CrawlInitialDelay,
	}

	if err := crawler.prepareConfig(); err != nil {
		return "", err
	}

	url := crawler.sharedData.EntryData.ProfilePage[0].GraphQL.User.ProfilePicUrl
	if url == "" {
		return "", fmt.Errorf("profile image missing")
	}

	return url, nil
}

func FetchResources(config *Config) ([]Resource, error) {
	crawler := &Crawler{
		config:     config,
		crawlDelay: CrawlInitialDelay,
		store:      &ResourceStore{},
	}

	if err := crawler.prepareConfig(); err != nil {
		return nil, err
	}

	if err := crawler.crawl(); err != nil {
		return nil, err
	}

	return crawler.store.resources, nil
}

func (c *Crawler) prepareConfig() error {
	profileUrl := "https://www.instagram.com/" + c.config.Username + "/"
	response, err := c.fetch(profileUrl)
	if err != nil {
		return err
	}

	jsonString, err := extractSharedDataJsonString(response)
	if err != nil {
		return err
	}

	c.queryId, err = c.extractQueryId(response)
	if err != nil {
		return fmt.Errorf("couldn't find queryId")
	}

	err = json.Unmarshal([]byte(jsonString), &c.sharedData)
	if err != nil {
		return errors.Wrapf(err, "invalid main page json \"%s\"", jsonString)
	}

	if c.sharedData.EntryData.ProfilePage[0].GraphQL.User.IsPrivate {
		return fmt.Errorf("\"%s\" is private account", c.config.Username)
	}

	c.userId = c.sharedData.EntryData.ProfilePage[0].GraphQL.User.Id
	if c.userId == "" {
		return fmt.Errorf("couldn't find userId")
	}

	c.rhxGis = c.sharedData.RhxGis
	if c.rhxGis == "" {
		return fmt.Errorf("couldn't find rhx-gis")
	}

	return nil
}

func (c *Crawler) crawl() error {
	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup root media
	c.handleMedia(ctx, c.sharedData.EntryData.ProfilePage[0].GraphQL.User.Media)

	for i := 0; i < c.config.MaxConnections; i++ {
		eg.Go(func() error {
			if err := c.workerWithContext(ctx); err != nil {
				cancel()
				return err
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (c *Crawler) setCrawlNextDelay() {
	delaySec := float64(c.crawlDelay) / float64(time.Second)
	requestCount := math.Round(math.Pow(10, delaySec/2.5))
	nextDelay := time.Duration(2.5*math.Log10(requestCount+1)*1000) * time.Millisecond
	c.crawlDelay = nextDelay

	// 最大遅延を5秒に抑える
	if nextDelay >= CrawlMaximumDelay {
		c.crawlDelay = CrawlMaximumDelay
	}

	// 最小遅延は0秒より下回らない
	if nextDelay < CrawlInitialDelay {
		c.crawlDelay = CrawlInitialDelay
	}
}

func (c *Crawler) fetch(url string) ([]byte, error) {
	return c.fetchWithHeaders(url, map[string]string{})
}

func (c *Crawler) fetchWithHeaders(url string, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// ヘッダを追加
	request.Header.Set("user-agent", c.config.UserAgent)
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	// GraphQLの場合はリクエストを遅延させる
	if isGraphqlRequest(request) && c.wait != nil {
		<-c.wait
		// 次のリクエストの遅延を計算
		c.setCrawlNextDelay()
		// リクエストの遅延を設定
		c.wait = time.After(c.crawlDelay)
	}

	response, err := fetchWithRequest(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Crawler) signatureFromParams(p string) string {
	hasher := md5.New()
	hasher.Write([]byte(c.rhxGis + ":" + p))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (c *Crawler) extractQueryId(response []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return "", err
	}

	var queryId string
	doc.Find("script").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		scriptUri, exists := s.Attr("src")
		if exists && strings.Contains(scriptUri, "/static/bundles/base/ProfilePageContainer.js") {
			scriptUrl := "https://www.instagram.com" + scriptUri
			response, err := c.fetch(scriptUrl)
			if err != nil {
				return false
			}

			regex := regexp.MustCompile(`queryId:"([^"]+)"`)
			result := regex.FindAllStringSubmatch(string(response), -1)
			queryId = result[2][1]
			return false
		}

		return true
	})

	if queryId == "" {
		return "", fmt.Errorf("couldn't find queryId")
	}

	return queryId, nil
}

func (c *Crawler) workerWithContext(ctx context.Context) error {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case resource := <-resourceChan:
			err := c.handleResource(ctx, resource)
			if err != nil {
				return err
			}
			continue
		case resource := <-galleryPageChan:
			err := c.handleGalleryPage(ctx, resource)
			if err != nil {
				return err
			}
			continue
		case resource := <-videoPageChan:
			err := c.handleVideoPage(ctx, resource)
			if err != nil {
				return err
			}
			continue
		case page := <-pageChan:
			err := c.handlePage(ctx, page)
			if err != nil {
				return err
			}
			continue
		default:
			break loop
		}
	}

	return nil
}

func (c *Crawler) handleMedia(ctx context.Context, m mediaJsonType) {
	hasNextPage := m.PageInfo.HasNextPage
	for _, element := range m.Edges {
		if element.Node.Timestamp <= c.config.After {
			hasNextPage = false
			continue
		}

		if !element.Node.IsVideo {
			if element.Node.Typename == "GraphImage" {
				resourceChan <- Resource{
					Url:       element.Node.DisplaySrc,
					Timestamp: element.Node.Timestamp,
					IsVideo:   false,
				}
			}
			if element.Node.Typename == "GraphSidecar" {
				galleryPageChan <- Resource{
					Url:       "https://www.instagram.com/p/" + element.Node.Code,
					Timestamp: element.Node.Timestamp,
					IsVideo:   false,
				}
			}
		} else {
			videoPageChan <- Resource{
				Url:       "https://www.instagram.com/p/" + element.Node.Code,
				Timestamp: element.Node.Timestamp,
				IsVideo:   true,
			}
		}
	}

	if hasNextPage {
		pageChan <- page{m.PageInfo.EndCursor}
	}
}

func (c *Crawler) handlePage(ctx context.Context, p page) error {
	params := "{\"id\":" + string(c.userId) + ",\"first\":" + "12" + ",\"after\":\"" + p.cursor + "\"}"
	queryUrl := "https://www.instagram.com/graphql/query/?query_hash=" + c.queryId + "&variables=" + url.QueryEscape(params)
	response, err := c.fetchWithHeaders(queryUrl, map[string]string{"x-instagram-gis": c.signatureFromParams(params)})
	if err != nil {
		return err
	}

	pageJson := pageJsonType{}
	if err = json.Unmarshal(response, &pageJson); err != nil {
		return errors.Wrapf(err, "invalid graphql json \"%s\"", string(response))
	}

	c.handleMedia(ctx, pageJson.Data.User.Media)

	return nil
}

func (c *Crawler) handleGalleryPage(ctx context.Context, r Resource) error {
	if r.Timestamp <= c.config.After {
		return nil
	}

	response, err := c.fetch(r.Url)
	if err != nil {
		return err
	}

	jsonString, err := extractSharedDataJsonString(response)
	if err != nil {
		return err
	}

	pageJson := galleryPageJsonType{}
	if err = json.Unmarshal([]byte(jsonString), &pageJson); err != nil {
		return errors.Wrapf(err, "invalid gallery page json \"%s\"", jsonString)
	}

	for _, element := range pageJson.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecarToChildren.Edges {
		if !element.Node.IsVideo {
			resourceChan <- Resource{
				Url:       element.Node.DisplaySrc,
				Timestamp: r.Timestamp,
				IsVideo:   false,
			}
		} else {
			resourceChan <- Resource{
				Url:       element.Node.VideoUrl,
				Timestamp: r.Timestamp,
				IsVideo:   true,
			}
		}
	}

	return nil
}

func (c *Crawler) handleVideoPage(ctx context.Context, r Resource) error {
	if r.Timestamp <= c.config.After {
		return nil
	}

	response, err := c.fetch(r.Url)
	if err != nil {
		return err
	}

	jsonString, err := extractSharedDataJsonString(response)
	if err != nil {
		return err
	}

	pageJson := videoPageJsonType{}
	if err = json.Unmarshal([]byte(jsonString), &pageJson); err != nil {
		return errors.Wrapf(err, "invalid video page json \"%s\"", jsonString)
	}

	resourceChan <- Resource{
		Url:       pageJson.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoUrl,
		Timestamp: r.Timestamp,
		IsVideo:   true,
	}

	return nil
}

func (c *Crawler) handleResource(ctx context.Context, r Resource) error {
	c.store.Lock()
	defer c.store.Unlock()

	c.store.resources = append(c.store.resources, r)
	return nil
}

func isGraphqlRequest(request *http.Request) bool {
	return strings.Contains(request.URL.Path, "graphql")
}

func extractSharedDataJsonString(response []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(response))
	if err != nil {
		return "", err
	}

	var jsonString string
	doc.Find("script").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		script := s.Text()
		if strings.HasPrefix(script, "window._sharedData") {
			jsonString = script[21 : len(script)-1]
			return false
		}

		return true
	})

	if jsonString == "" {
		return "", fmt.Errorf("couldn't find window._sharedData")
	}

	return jsonString, nil
}

type page struct {
	cursor string
}

type mediaJsonType struct {
	Count int `json:"count"`
	Edges []struct {
		Node struct {
			Typename   string `json:"__typename"`
			Id         string `json:"id"`
			IsVideo    bool   `json:"is_video"`
			Code       string `json:"shortcode"`
			Timestamp  int    `json:"taken_at_timestamp"`
			DisplaySrc string `json:"display_url"`
		} `json:"node"`
	} `json:"edges"`
	PageInfo struct {
		HasNextPage bool   `json:"has_next_page"`
		EndCursor   string `json:"end_cursor"`
	} `json:"page_info"`
}

type sharedDataJsonType struct {
	EntryData struct {
		ProfilePage []struct {
			GraphQL struct {
				User struct {
					Id            string        `json:"id"`
					Media         mediaJsonType `json:"edge_owner_to_timeline_media"`
					ProfilePicUrl string        `json:"profile_pic_url_hd"`
					IsPrivate     bool          `json:"is_private"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
	RhxGis string `json:"rhx_gis"`
}

type pageJsonType struct {
	Data struct {
		User struct {
			Media mediaJsonType `json:"edge_owner_to_timeline_media"`
		} `json:"user"`
	} `json:"data"`
}

type galleryPageJsonType struct {
	EntryData struct {
		PostPage []struct {
			Graphql struct {
				ShortcodeMedia struct {
					EdgeSidecarToChildren struct {
						Edges []struct {
							Node struct {
								Typename     string `json:"__typename"`
								Id           string `json:"id"`
								MediaPreview string `json:"media_preview"`
								IsVideo      bool   `json:"is_video"`
								DisplaySrc   string `json:"display_url"`
								VideoUrl     string `json:"video_url"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_sidecar_to_children"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
}

type videoPageJsonType struct {
	EntryData struct {
		PostPage []struct {
			Graphql struct {
				ShortcodeMedia struct {
					Typename string `json:"__typename"`
					Id       string `json:"id"`
					VideoUrl string `json:"video_url"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
}
