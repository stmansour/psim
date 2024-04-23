// A simple time series plotter using Plotly and PapaParse

let datasetTimeSeries = [];  // Make sure this is declared at the top of your file
let dateColumn = "Date";  // Adjust this if your date column has a different header

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

// Normalize data function
function normalizeData(values) {
    let mean = ss.mean(values);
    let standardDeviation = ss.standardDeviation(values);
    return values.map(value => (value - mean) / standardDeviation);
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

    let xValues = filteredData.map(data => data[dateColumn]);
    let yValues1 = filteredData.map(data => data[metric1]);
    let yValues2 = filteredData.map(data => data[metric2]);

    let windowSize = 30; // Moving window size of 30 days
    let correlationValues = []; // Array to hold correlation values

    // Calculate moving window correlation
    for (let i = windowSize; i < xValues.length; i++) {
        let windowYValues1 = yValues1.slice(i - windowSize, i);
        let windowYValues2 = yValues2.slice(i - windowSize, i);

        // Ensure the window has enough data points
        if (windowYValues1.length >= windowSize && windowYValues2.length >= windowSize) {
            let correlation = ss.sampleCorrelation(windowYValues1, windowYValues2);
            correlationValues.push(correlation);
        } else {
            correlationValues.push(null); // Or handle insufficient data points
        }
    }

    let trace1 = {
        x: xValues,
        y: yValues1,
        mode: 'lines',
        name: metric1,
        yaxis: 'y1',
        line: { color: '#17BECF' }
    };

    let trace2 = {
        x: xValues,
        y: yValues2,
        mode: 'lines',
        name: metric2,
        yaxis: 'y2',
        line: { color: '#B22222' }
    };

    // Adjust xValues for correlation trace to match the windowed calculation
    let correlationXValues = xValues.slice(windowSize);

    let trace3 = {
        x: correlationXValues,
        y: correlationValues,
        mode: 'lines',
        name: 'Correlation',
        yaxis: 'y3',
        line: { color: '#DAA520' }
    };

    let layout = {
        title: 'Time Series Data with Correlation',
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
        },
        yaxis3: {
            title: 'Correlation',
            titlefont: { color: trace3.line.color },
            tickfont: { color: trace3.line.color },
            overlaying: 'y',
            side: 'right',
            position: 0.95
        }
    };

    Plotly.newPlot('plot', [trace1, trace2, trace3], layout);
}
