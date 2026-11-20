// Library catalogue: encoding/json with struct tags, time, maps and text/tabwriter.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type Author struct {
	Name string `json:"name"`
	Born int    `json:"born,omitempty"`
}

type Book struct {
	ID        int               `json:"id"`
	Title     string            `json:"title"`
	Authors   []Author          `json:"authors"`
	Tags      []string          `json:"tags,omitempty"`
	Rating    float64           `json:"rating"`
	Available bool              `json:"available"`
	Published string            `json:"published"`
	Extra     map[string]string `json:"extra,omitempty"`
	secret    string
}

type Catalogue struct {
	Name  string  `json:"name"`
	Books []*Book `json:"books"`
}

const raw = `{
  "name": "City Library",
  "books": [
    {"id": 1, "title": "The Go Programming Language", "authors": [{"name": "Alan Donovan", "born": 1970}, {"name": "Brian Kernighan", "born": 1942}],
     "tags": ["programming", "go"], "rating": 4.7, "available": true, "published": "2015-11-16T00:00:00Z"},
    {"id": 2, "title": "Dune", "authors": [{"name": "Frank Herbert", "born": 1920}], "tags": ["scifi"],
     "rating": 4.5, "available": false, "published": "1965-08-01T00:00:00Z", "extra": {"series": "Dune", "volume": "1"}},
    {"id": 3, "title": "Poems", "authors": [{"name": "Anonymous"}], "rating": 3.2, "available": true, "published": "1850-01-01T00:00:00Z"}
  ]
}`

func mustDate(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	var cat Catalogue
	if err := json.Unmarshal([]byte(raw), &cat); err != nil {
		fmt.Println("decode error:", err)
		return
	}
	fmt.Println(cat.Name, len(cat.Books), "books")

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTITLE\tAUTHORS\tYEAR\tRATING\tIN")
	for _, b := range cat.Books {
		var names []string
		for _, a := range b.Authors {
			names = append(names, a.Name)
		}
		in := "no"
		if b.Available {
			in = "yes"
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%.1f\t%s\n", b.ID, b.Title, strings.Join(names, " & "), mustDate(b.Published).Year(), b.Rating, in)
	}
	tw.Flush()

	sort.Slice(cat.Books, func(i, j int) bool { return cat.Books[i].Rating > cat.Books[j].Rating })
	fmt.Println("best:", cat.Books[0].Title)

	cat.Books[1].Available = true
	cat.Books[2].Extra = map[string]string{"note": "loaned out", "branch": "north"}
	cat.Books = append(cat.Books, &Book{ID: 4, Title: "New Arrival", Rating: 0, Published: "2024-02-29T12:30:00Z", secret: "x"})
	out, err := json.MarshalIndent(cat.Books[2:], "", "  ")
	fmt.Println(string(out), err)

	compact, _ := json.Marshal(map[string]any{"b": []int{1, 2}, "a": nil, "c": 1.5, "d": "x<y"})
	fmt.Println(string(compact))

	var generic map[string]any
	json.Unmarshal([]byte(`{"n": 3, "list": [1, "two", true, null], "obj": {"k": "v"}}`), &generic)
	keys := make([]string, 0)
	for k := range generic {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %v (%T)\n", k, generic[k], generic[k])
	}

	var bad Book
	fmt.Println(json.Unmarshal([]byte(`{"id": "seven"}`), &bad))
	fmt.Println(json.Unmarshal([]byte(`{"id": 7,`), &bad))

	first, last := mustDate(cat.Books[0].Published), mustDate(cat.Books[3].Published)
	d := first.Sub(last)
	fmt.Println(first.Format("Mon Jan 2 2006"), first.Weekday(), int(d.Hours()/24/365))
	t0 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	fmt.Println(t0.AddDate(0, 1, 0).Format("2006-01-02"), t0.Add(36*time.Hour).Format(time.RFC3339), time.Duration(5400)*time.Second)
}
