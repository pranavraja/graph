Produces interactive charts of time series data, using the free Google Chart
SDK (see
https://developers.google.com/chart/interactive/docs/gallery/areachart)

# Background

Sometimes you just have a bunch of timestamps, e.g. from the results of a
database query, or extracted from a logfile.

You would like to resample these timestamps with a reasonable interval and
visualise the trend over time. `graph` is a simple tool that helps do this.

# Example

Suppose you have a file `series1.txt`, with unix timestamps, one per line, and
not necessarily sorted, e.g.

```
1588946340
1466776740
1585313940
1525183140
1455800340
1487681940
1455800340
1455800340
1489064340
```

You can run `graph series1.txt`, which will:

- aggregate the data by a reasonable time interval
- open a browser to show you an interactive chart with the trend
- give you some buttons + an input field to change the sampling interval

If instead, your data is in the form of rfc3339 timestamps, like this:

```
2019-06-06T23:00:17.602349Z
2019-06-06T23:00:18.408788Z
2019-06-06T23:00:18.625850Z
2019-06-06T23:00:18.691918Z
2019-06-06T23:00:18.639435Z
2019-06-06T23:00:18.646476Z
2019-06-06T23:00:18.725604Z
2019-06-06T23:00:18.732247Z
```

...then just add `-rfc` as a flag, e.g. `graph -rfc series1.txt`

# Prerequisites

- Go (https://golang.org/doc/install)

# Installation

    go get github.com/pranavraja/graph

# Usage

    graph [-rfc] series1.txt [series2.txt]

Opens a browser showing an interactive chart of your time series data,
aggregated by a reasonable sampling interval.

There are buttons to change the sampling interval as you see fit to better
visualise the trend.

If you provide multiple series, and the range of data is different enough, it
will automatically plot the first series markers on the left y axis and the
second on the right y axis.

# TODO

- Linux support for opening the browser
