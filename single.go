package main

import (
	"html/template"
)

var singleTimeSeries = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
  <head>
    <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
    <script type="text/javascript">
      google.charts.load('current', {'packages':['corechart']});
      google.charts.setOnLoadCallback(drawChart);

      function drawChart() {
        var data = google.visualization.arrayToDataTable([
          ['Time', '{{ .Y }}'],
        {{ range .Data }}
            [new Date({{.Time}}), {{.Count}}],
        {{ end }}
        ]);

        var options = {
          title: '{{ .Title }}',
          hAxis: { title: 'Time', titleTextStyle: {color: '#333'} },
          explorer: {
            actions: ['dragToZoom', 'rightClickToReset'],
            axis: 'horizontal',
            keepInBounds: true,
            maxZoomIn: 4.0
          },
        };

        var chart = new google.visualization.AreaChart(document.getElementById('chart_div'));
        chart.draw(data, options);
      }
    </script>
  </head>
  <body>
    <div id="chart_div" style="width: 100%; height: 500px;"></div>
    <section>
    <form>
        <label for="sample">Resample</label>
        <input type="text" placeholder="5m" name="sample" id="sample">
	<button onClick="sample.value='5m'; this.parentNode.submit()" type="button">5m</button>
	<button onClick="sample.value='1h'; this.parentNode.submit()" type="button">1h</button>
	<button onClick="sample.value='24h'; this.parentNode.submit()" type="button">1d</button>
	<button onClick="sample.value='168h'; this.parentNode.submit()" type="button">1w</button>
    </form>
    </section>
  </body>
</html>`))
