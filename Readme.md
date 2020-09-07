# Counter

Tracks sum, count, and average for any arbitrary metrics that you send to it.

## Installation

Download the binary for your platform from the bin directory to the location where you want it
to store data, then run it as an unpriviledged user. It will run on port 8080 by default.

## API endpoints

### GET /metric

List metrics that are currently being tracked

### GET /metric/:name

Get a JSON representation of all samples currently being tracked for this metric

### GET /metric/:name/:operation.png

Get a PNG representation of this metric, where operation is "sum", "count", or "avg"

### POST /metric

Send metrics to be tracked. Any unrecognized metrics will be created automatically.
The API format is as follows:

    {
        "metric.name": {
            "count": 100,
            "value": 86753.09
        }
    }

Multiple metrics can be sent in a single request, and you can aggregate upstream by including a
count in your request. Count is optional, and will default to 1.

## Purging data

If you no longer wish to track a metric, stop the server, delete its datafile in the data directory,
then restart the server.