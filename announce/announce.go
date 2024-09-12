package announce

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chazari-x/hmtpk_parser/model"
	"github.com/sirupsen/logrus"
)

type Announce struct {
	log *logrus.Logger
	re  *regexp.Regexp
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
func (a *Announce) GetAnnounces(ctx context.Context, page int) ([]model.Announce, error) {
	doc, err := a.getDocument(ctx, page)
	if err != nil {
		return nil, err
	}

	return a.parseAnnounces(doc), nil
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
	announce.Html, err = s.Html()
	if err != nil {
		return
	}

	announce.Html = a.removeExtraSpaces(announce.Html)

	announce.Path, err = a.searchAnnouncePath(s)
	if err != nil {
		return
	}

	return
}

func (a *Announce) searchAnnouncePath(s *goquery.Selection) (string, error) {
	path, exists := s.Find("h3 > a").First().Attr("href")
	if !exists {
		return "", errors.New("path not found")
	}

	return strings.TrimSpace(path), nil
}

// Функция для удаления лишних пробелов между HTML-блоками
func (a *Announce) removeExtraSpaces(html string) string {
	cleanedHTML := a.re.ReplaceAllString(html, " ")
	return strings.TrimSpace(cleanedHTML)
}
