package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	flagImgPathPrefix = flag.String("imgprefix", "/images/", "prefix of image path")
	flagImgDir        = flag.String("imgdir", "images", "image dir")
	flagPostDir       = flag.String("postdir", "posts", "posts dir")
	flagTmplFile      = flag.String("template", "", "template file")

	imgRegexp = regexp.MustCompile(`https://qiita-image-store\.s3\.amazonaws\.com/.+\.png`)
)

type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

type Item struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	RenderedBody string    `json:"rendered_body"`
	CreatedAt    time.Time `json:"created_at"`
	Tags         []*Tag    `json:"tags"`
	Private      bool      `json:"private"`
}

func (item *Item) AllTags() string {
	tags := make([]string, len(item.Tags))
	for i := range item.Tags {
		tags[i] = strconv.Quote(item.Tags[i].Name)
	}
	return strings.Join(tags, ",")
}

func (item *Item) Date() string {
	return item.CreatedAt.Format("2006-01-02")
}

func (item *Item) ImageToLocal(dir string) error {
	var (
		rerr  error
		count int
	)
	body := imgRegexp.ReplaceAllStringFunc(item.Body, func(s string) string {
		if rerr != nil {
			return s
		}

		count++
		ext := path.Ext(s)
		fname := fmt.Sprintf("qiita-%s-%d%s", item.ID, count, ext)
		f, err := os.Create(filepath.Join(dir, fname))
		if err != nil {
			rerr = err
			return s
		}

		resp, err := http.Get(s)
		if err != nil {
			rerr = err
			return s
		}
		defer resp.Body.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			rerr = err
			return s
		}

		if err := f.Close(); err != nil {
			rerr = err
			return s
		}

		return path.Join(*flagImgPathPrefix, fname)
	})

	if rerr != nil {
		return rerr
	}

	item.Body = body

	return nil
}

func main() {
	flag.Parse()

	if *flagTmplFile != "" {
		var err error
		tmpl, err = template.ParseFiles(*flagTmplFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}

	for i := 1; ; i++ {
		hasNext, err := download100(i)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if !hasNext {
			break
		}
	}
}

func download100(page int) (hasNext bool, rerr error) {

	url := fmt.Sprintf("https://qiita.com/api/v2/authenticated_user/items?page=%d&per_page=20", page)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}

	var items []*Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return false, err
	}

	imgdir := filepath.Join(*flagPostDir, *flagImgDir)
	if err := os.MkdirAll(imgdir, 0777); err != nil {
		return false, err
	}

	for i := range items {
		if items[i].Private {
			continue
		}

		if err := items[i].ImageToLocal(imgdir); err != nil {
			return false, err
		}

		fname := fmt.Sprintf("%s-qiita-%s.ja.md", items[i].Date(), items[i].ID)
		fmt.Print(items[i].Title, "....")
		f, err := os.Create(filepath.Join(*flagPostDir, fname))
		if err != nil {
			return false, err
		}

		if err := tmpl.Execute(f, items[i]); err != nil {
			return false, err
		}

		if err := f.Close(); err != nil {
			return false, err
		}

		fmt.Println("done")
	}

	total, err := strconv.Atoi(resp.Header.Get("Total-Count"))
	if err != nil {
		return false, err
	}

	return page < total, nil
}

func do(req *http.Request) (*http.Response, error) {
	token := fmt.Sprintf("Bearer %s", os.Getenv("QIITA"))
	req.Header.Set("Authorization", token)
	return http.DefaultClient.Do(req)
}
