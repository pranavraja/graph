package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	Time  int64
	Count int64
}

func resample(times []int64, d time.Duration, multiply int64, cumulative bool) []sample {
	if len(times) == 0 {
		return nil
	}
	var samples []sample
	current := sample{
		Time: times[0],
	}
	for _, t := range times {
		if next := current.Time + int64(d/time.Millisecond); t < next {
			current.Count += multiply
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

func timestamps(f io.Reader, rfc bool) ([]int64, error) {
	scanner := bufio.NewScanner(f)
	i := 0
	var times []int64
	for scanner.Scan() {
		i++
		if rfc {
			next, err := time.Parse(time.RFC3339, scanner.Text())
			if err != nil {
				return nil, fmt.Errorf("error on line %d: %s", i, err)
			}
			times = append(times, next.UnixNano()/int64(time.Millisecond))
			continue
		}
		next, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error on line %d: %s", i, err)
		}
		times = append(times, next*1000)
	}
	return times, scanner.Err()
}

func percentile95(values []sample) int64 {
	v := make([]sample, len(values))
	copy(v, values)
	sort.Slice(v, func(i, j int) bool {
		return v[i].Count > v[j].Count
	})
	return v[len(v)/20].Count
}

func sortInt64s(ts []int64) {
	sort.Slice(ts, func(i, j int) bool {
		return ts[i] < ts[j]
	})
}

func main() {
	var (
		rfc      bool
		multiply int64
	)
	flag.BoolVar(&rfc, "rfc", false, "Parse dates as RFC3339 first")
	flag.Int64Var(&multiply, "multiply", 1, "multiple y axis by a certain factor")
	flag.Parse()
	filenames := flag.Args()
	var data []byte
	if stdin, err := os.Stdin.Stat(); err == nil && stdin.Mode()&os.ModeCharDevice == 0 {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("couldn't read stdin: %s", err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, _ := time.ParseDuration(r.FormValue("sample"))
		switch len(filenames) {
		case 0:
			times, err := timestamps(bytes.NewReader(data), rfc)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sortInt64s(times)
			if len(times) == 0 {
				http.Error(w, "no data to display", 500)
				return
			}
			cumulative := r.FormValue("cumulative") != ""
			if d == 0 {
				distance := time.Duration(times[len(times)-1]-times[0]) * time.Millisecond
				if distance < time.Hour {
					d = time.Second
				} else if distance < 168*time.Hour {
					d = 5 * time.Minute
				} else {
					d = 24 * time.Hour
				}
			}
			samples := resample(times, d, multiply, cumulative)
			var graph struct {
				Title string
				Y     string
				Data  []sample
			}
			graph.Title = r.FormValue("title")
			graph.Y = "count"
			graph.Data = samples
			singleTimeSeries.Execute(w, graph)
		case 1:
			first, err := os.Open(filenames[0])
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			times, err := timestamps(first, rfc)
			first.Close()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sortInt64s(times)
			if len(times) == 0 {
				http.Error(w, "no data to display", 500)
				return
			}
			cumulative := r.FormValue("cumulative") != ""
			if d == 0 {
				distance := time.Duration(times[len(times)-1]-times[0]) * time.Millisecond
				if distance < time.Hour {
					d = time.Second
				} else if distance < 168*time.Hour {
					d = 5 * time.Minute
				} else {
					d = 24 * time.Hour
				}
			}
			samples := resample(times, d, multiply, cumulative)
			var graph struct {
				Title string
				Y     string
				Data  []sample
			}
			graph.Title = r.FormValue("title")
			graph.Y = strings.TrimSuffix(filenames[0], filepath.Ext(filenames[0]))
			graph.Data = samples
			singleTimeSeries.Execute(w, graph)
		case 2:
			first, err := os.Open(filenames[0])
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			times1, err := timestamps(first, rfc)
			first.Close()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sortInt64s(times1)
			if len(times1) == 0 {
				http.Error(w, "no data to display", 500)
				return
			}
			second, err := os.Open(filenames[1])
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			times2, err := timestamps(second, rfc)
			second.Close()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sortInt64s(times2)
			if len(times2) == 0 {
				http.Error(w, "no data to display", 500)
				return
			}
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

			samples1 := resample(times1, d, multiply, cumulative)
			samples2 := resample(times2, d, multiply, cumulative)
			var graph struct {
				Title   string
				Y1      string
				Y2      string
				Data1   []sample
				Data2   []sample
				TwoAxes bool
			}
			const thresholdFor2Axes = 5
			if percentile95(samples1)/percentile95(samples2) > thresholdFor2Axes || percentile95(samples2)/percentile95(samples1) > thresholdFor2Axes {
				graph.TwoAxes = true
			}
			graph.Title = r.FormValue("title")
			graph.Y1 = strings.TrimSuffix(filenames[0], filepath.Ext(filenames[0]))
			graph.Y2 = strings.TrimSuffix(filenames[1], filepath.Ext(filenames[1]))
			graph.Data1 = samples1
			graph.Data2 = samples2
			doubleTimeSeries.Execute(w, graph)
		}
	}))
	defer server.Close()
	exec.Command("open", server.URL).Run()
	select {}
}
