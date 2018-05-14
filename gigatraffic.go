package gigatraffic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

type Quota struct {
	Start       time.Time
	Limit, Used uint64
}

func (q *Quota) End() time.Time {
	return q.Start.AddDate(0, 1, 0)
}

func (q *Quota) Duration() time.Duration {
	return q.End().Sub(q.Start)
}

func (q *Quota) Elapsed() time.Duration {
	return time.Now().Sub(q.Start)
}

func (q *Quota) Remaining() time.Duration {
	return q.End().Sub(time.Now())
}

func (q *Quota) PercentElapsed() float64 {
	return 100 * q.Elapsed().Seconds() / q.Duration().Seconds()
}

func (q *Quota) PercentRemaining() float64 {
	return 100 * q.Remaining().Seconds() / q.Duration().Seconds()
}

func (q *Quota) Available() uint64 {
	return q.Limit - q.Used
}

func (q *Quota) PercentUsed() float64 {
	return 100 * float64(q.Used) / float64(q.Limit)
}

func (q *Quota) PercentAvailable() float64 {
	return 100 * float64(q.Available()) / float64(q.Limit)
}

func (q *Quota) BytesPerSecond() float64 {
	return float64(q.Used) / float64(q.Elapsed().Seconds())
}

func (q *Quota) UsageOver(d time.Duration) uint64 {
	return uint64(q.BytesPerSecond() * d.Seconds())
}

func (q *Quota) SufficientFor(bytes uint64) time.Duration {
	d, err := time.ParseDuration(fmt.Sprintf("%fs", 1.0/q.BytesPerSecond()*float64(bytes)))
	if err != nil {
		return 0
	}
	return d
}

func (q *Quota) Leftover() int64 {
	return int64(q.Available()) - int64(q.UsageOver(q.Remaining()))
}

func (q *Quota) Depletion() time.Time {
	return time.Now().Add(q.SufficientFor(q.Available()))
}

func (q *Quota) WillSuffice() bool {
	return q.Leftover() >= 0
}

func (q *Quota) String() string {
	leftover := q.Leftover()
	var leftoverText string
	if leftover > 0 {
		leftoverText = fmt.Sprintf("%s to be left at the end of the interval", humanize.Bytes(uint64(leftover)))
	} else {
		leftoverText = fmt.Sprintf("%s to be missing until the end of the interval", humanize.Bytes(uint64(-1*leftover)))
	}
	return fmt.Sprintf("%s of %s (%.0f%%) used after %.0f%% of the time (%.0f bytes/s), estimating %s, depleted %s",
		humanize.Bytes(q.Used),
		humanize.Bytes(q.Limit),
		q.PercentUsed(),
		q.PercentElapsed(),
		q.BytesPerSecond(),
		leftoverText,
		q.Depletion().Format("2006-01-02 15:04:05"),
	)
}

func Retrieve() (Quota, error) {
	q := Quota{}
	doc, err := download("http://quickcheck.vodafone.de/")
	if err != nil {
		return q, err
	}
	err = fill(&q, doc)
	if err != nil {
		return q, err
	}
	return q, nil
}

func download(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving quota page")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("got HTTP %d, expected 200", res.StatusCode))
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "parsing HTML")
	}
	return doc, nil
}

func fill(quota *Quota, doc *goquery.Document) error {
	start, err := time.Parse("seit dem 02.01.2006",
		doc.Find("table").Eq(1).Find("tr").Eq(0).Find("td[align='right'] > span.hd1").Text())
	if err != nil {
		return errors.Wrap(err, "finding and parsing date")
	}
	quota.Start = start

	var eachErr error
	doc.Find("table").Eq(2).Find("tr").Eq(3).Find("td").Each(func(i int, el *goquery.Selection) {
		bytes, eachErr := humanize.ParseBytes(el.Text())
		if eachErr != nil {
			return
		}
		switch i {
		case 0:
			quota.Limit = bytes
		case 1:
			quota.Used = bytes
		}
	})
	if eachErr != nil {
		return eachErr
	}
	return nil
}
