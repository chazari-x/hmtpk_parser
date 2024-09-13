package announce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chazari-x/hmtpk_parser/model"
	"github.com/chazari-x/hmtpk_parser/storage"
	"github.com/chazari-x/hmtpk_parser/utils"
	"github.com/sirupsen/logrus"
)

type Announce struct {
	log *logrus.Logger
	re  *regexp.Regexp
	r   *storage.Redis
}

func NewAnnounce(logger *logrus.Logger) *Announce {
	return &Announce{
		log: logger,
		re:  regexp.MustCompile(`\s+`),
	}
}

const (
	href = "https://hmtpk.ru/ru/press-center/announce"
)

// GetAnnounces получает блок с объявлениями с сайта hmtpk.ru и возвращает его как html строку
func (a *Announce) GetAnnounces(ctx context.Context, page int) (announces model.Announces, err error) {
	if utils.RedisIsNil(a.r) {
		if redisData, err := a.r.Get(fmt.Sprintf("announce?page=%d", page)); err == nil && redisData != "" {
			if json.Unmarshal([]byte(redisData), &announces) == nil {
				a.log.Trace("announces получены из redis")
				return announces, nil
			}
		}
	}

	doc, err := a.getDocument(ctx, page)
	if err != nil {
		return
	}

	announces.Announces = a.parseAnnounces(doc)

	announces.LastPage, err = a.searchLastPage(doc)
	if err != nil {
		return
	}

	if utils.RedisIsNil(a.r) {
		if marshal, err := json.Marshal(announces); err == nil {
			if err = a.r.Set(fmt.Sprintf("announce?page=%d", page), string(marshal), 60); err != nil {
				a.log.Error(err)
			} else {
				a.log.Trace("announces сохранены в redis")
			}
		}
	}

	return
}

// getDocument получает html страницу с сайта hmtpk.ru
func (a *Announce) getDocument(ctx context.Context, page int) (*goquery.Document, error) {
	href := fmt.Sprintf("%s?PAGEN_1=%d", href, page)
	request, err := http.NewRequestWithContext(ctx, "POST", href, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Ошибка: %s", resp.Status))
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func (a *Announce) parseAnnounces(doc *goquery.Document) []model.Announce {
	announcesBlock := doc.Find("section.sf-pagewrap-area.overflow-hidden.d-flex.flex-col.justify-content-start > div > section > main > section > div > div.row").First()

	announces := make([]model.Announce, 0, 10)
	announcesBlock.Find("div.iblock-list-item-text.p-3").Each(func(i int, s *goquery.Selection) {
		announce, err := a.parseAnnounce(s)
		if err != nil {
			a.log.Error(err)
			return
		}

		announces = append(announces, announce)
	})

	return announces
}

func (a *Announce) parseAnnounce(s *goquery.Selection) (announce model.Announce, err error) {
	announce.Date, err = a.searchDate(s)
	if err != nil {
		return
	}

	announce.Path, announce.Title, err = a.searchAnnounceTitleAndPath(s)
	if err != nil {
		return
	}

	announce.Body, err = a.searchBody(s)
	if err != nil {
		return
	}

	return
}

func (a *Announce) searchAnnounceTitleAndPath(s *goquery.Selection) (string, string, error) {
	element := s.Find("h3 > a").First()

	path, exists := element.Attr("href")
	if !exists {
		return "", "", errors.New("path not found")
	}

	title := strings.ReplaceAll(element.Text(), "\n", " ")

	return strings.TrimSpace(path), strings.TrimSpace(title), nil
}

func (a *Announce) searchBody(s *goquery.Selection) (string, error) {
	body, err := s.Find("div.c-text-secondary").Html()
	if err != nil {
		return "", err
	}

	if body == "" {
		return "", errors.New("body not found")
	}

	return a.removeExtraSpaces(body), nil
}

func (a *Announce) searchDate(s *goquery.Selection) (string, error) {
	date := s.Find("p.c-text-secondary").First().Text()
	if date == "" {
		return "", errors.New("date not found")
	}

	return strings.TrimSpace(date), nil
}

// Функция для удаления лишних пробелов между HTML-блоками
func (a *Announce) removeExtraSpaces(html string) string {
	cleanedHTML := a.re.ReplaceAllString(html, " ")
	return strings.TrimSpace(cleanedHTML)
}

func (a *Announce) searchLastPage(doc *goquery.Document) (int, error) {
	elements := doc.Find("main div.sf-viewbox.position-relative > div:last-child > *")

	if elements.Length() == 0 {
		return 0, errors.New("elements not found")
	}

	lastElement := elements.Last()

	if lastElement.Is("span") {
		page, err := strconv.Atoi(lastElement.Text())
		if err != nil {
			return 0, err
		}

		return page, nil
	}

	page, err := strconv.Atoi(lastElement.Prev().Text())
	if err != nil {
		return 0, err
	}

	return page, nil
}
