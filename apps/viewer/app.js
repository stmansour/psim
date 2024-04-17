// A simple time series plotter using Plotly and PapaParse

let datasetTimeSeries = [];  // Make sure this is declared at the top of your file
let dateColumn = "Date";  // Adjust this if your date column has a different header

// Load CSV from the specified directory
Papa.parse('data/platodb.csv', {
    download: true,
    header: true,
    dynamicTyping: true,
    skipEmptyLines: true,
    complete: function(results) {
        datasetTimeSeries = results.data;  // Ensure data is assigned to datasetTimeSeries
        if (datasetTimeSeries.length > 0) {
            initializeControls();
        }
    }
});

function initializeControls() {
    // Ensure we have data to work with
    if (datasetTimeSeries.length === 0) {
        console.error("No data available for initializing controls.");
        return;
    }

    // Get all column headers except the date column
    let metricOptions = Object.keys(datasetTimeSeries[0]).filter(key => key !== dateColumn);
    const metric1Select = document.getElementById('metric1-select');
    const metric2Select = document.getElementById('metric2-select');

    // Clear existing options
    metric1Select.innerHTML = '';
    metric2Select.innerHTML = '';

    // Append new options to both dropdowns
    metricOptions.forEach(option => {
        let opt1 = new Option(option, option);
        let opt2 = new Option(option, option);
        metric1Select.appendChild(opt1);
        metric2Select.appendChild(opt2);
    });

    // Set default date inputs
    let dateArray = datasetTimeSeries.map(data => new Date(data[dateColumn]));
    document.getElementById('start-date').valueAsDate = new Date(Math.min(...dateArray));
    document.getElementById('end-date').valueAsDate = new Date(Math.max(...dateArray));
}

function plotData() {
    let metric1 = document.getElementById('metric1-select').value;
    let metric2 = document.getElementById('metric2-select').value;
    let startDate = new Date(document.getElementById('start-date').value);
    let endDate = new Date(document.getElementById('end-date').value);

    let filteredData = datasetTimeSeries.filter(data => {
        let dataDate = new Date(data[dateColumn]);
        return (dataDate >= startDate && dataDate <= endDate);
    });

    let trace1 = {
        x: filteredData.map(data => data[dateColumn]),
        y: filteredData.map(data => data[metric1]),
        mode: 'lines',
        name: metric1,
        yaxis: 'y1',  // Assign to the first y-axis
        line: { color: '#17BECF' }  // Defaulting to a specific color
    };

    let trace2 = {
        x: filteredData.map(data => data[dateColumn]),
        y: filteredData.map(data => data[metric2]),
        mode: 'lines',
        name: metric2,
        yaxis: 'y2',  // Assign to the second y-axis
        line: { color: '#B22222' }  // Defaulting to a specific color
    };

    let layout = {
        title: 'Time Series Data',
        xaxis: { title: 'Date' },
        yaxis: {
            title: metric1,
            titlefont: { color: trace1.line.color },
            tickfont: { color: trace1.line.color },
            side: 'left'
        },
        yaxis2: {
            title: metric2,
            titlefont: { color: trace2.line.color },
            tickfont: { color: trace2.line.color },
            overlaying: 'y',
            side: 'right'
        }
    };

    Plotly.newPlot('plot', [trace1, trace2], layout);
}
