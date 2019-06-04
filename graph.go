package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type sample struct {
	Time  int
	Count int
}

func resample(times []int, d time.Duration, cumulative bool) []sample {
	if len(times) == 0 {
		return nil
	}
	var samples []sample
	current := sample{
		Time: times[0],
	}
	for _, t := range times {
		if next := current.Time + int(d/time.Millisecond); t < next {
			current.Count++
			continue
		}
		samples = append(samples, current)
		current.Time = t
		if !cumulative {
			current.Count = 0
		}
	}
	return samples
}

func timestamps(filename string) ([]int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	i := 0
	var times []int
	for scanner.Scan() {
		i++
		next, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("error on line %d: %s", i, err)
		}
		times = append(times, next*1000)
	}
	return times, scanner.Err()
}

func main() {
	filenames := os.Args[1:]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, _ := time.ParseDuration(r.FormValue("sample"))
		switch len(filenames) {
		case 1:
			first := filenames[0]
			times, err := timestamps(first)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sort.Ints(times)
			cumulative := r.FormValue("cumulative") != ""
			samples := resample(times, d, cumulative)
			var graph struct {
				Title string
				Y     string
				Data  []sample
			}
			graph.Title = r.FormValue("title")
			graph.Y = strings.TrimSuffix(first, filepath.Ext(first))
			graph.Data = samples
			singleTimeSeries.Execute(w, graph)
		case 2:
			first := filenames[0]
			second := filenames[1]
			times1, err := timestamps(first)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sort.Ints(times1)
			times2, err := timestamps(second)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sort.Ints(times2)
			cumulative := r.FormValue("cumulative") != ""
			if d == 0 {
				distance := time.Duration(times1[len(times1)-1]-times1[0]) * time.Millisecond
				if distance < time.Hour {
					d = time.Second
				} else if distance < 168*time.Hour {
					d = 5 * time.Minute
				} else {
					d = 24 * time.Hour
				}
			}
			samples1 := resample(times1, d, cumulative)
			samples2 := resample(times2, d, cumulative)
			var graph struct {
				Title string
				Y1    string
				Y2    string
				Data1 []sample
				Data2 []sample
			}
			graph.Title = r.FormValue("title")
			graph.Y1 = strings.TrimSuffix(first, filepath.Ext(first))
			graph.Y2 = strings.TrimSuffix(second, filepath.Ext(second))
			graph.Data1 = samples1
			graph.Data2 = samples2
			doubleTimeSeries.Execute(w, graph)
		}
	}))
	defer server.Close()
	exec.Command("open", server.URL).Run()
	select {}
}
