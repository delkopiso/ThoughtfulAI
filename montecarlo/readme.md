# Monte Carlo Take Home Test

## Running instructions
- Install the latest version of [Docker](https://docs.docker.com/desktop/).
- Pull down service images:
    ```shell
    docker compose pull
    ```
- Build application images:
    ```shell
    docker compose build
    ```
- Run application containers:
    ```shell
    docker compose up -d
    ```
- Make API requests:
    ```shell
    $ curl -iL -X GET http://localhost:8081/metrics
    [
      "btcusd",
      "btceur",
      "ethusd",
      "etheur",
    ...snip...
    ]
    ```

    ```shell
    $ curl -iL -X GET http://localhost:8081/metrics/btcusd
    {
      "pricePoints": [
        {
          "timestamp": "2022-01-24T03:57:29.961+00:00",
          "amount": 0.0026086
        },
        {
          "timestamp": "2022-01-24T03:57:59.912+00:00",
          "amount": 0.0026086
        },
        {
          "timestamp": "2022-01-24T03:58:31.116+00:00",
          "amount": 0.0026097
        },
      ...snip...
      ],
      "rank": "46/413"
    }
    ```

## TODOs
- Add unit tests for business logic
- Add integration tests to validate interaction with external services e.g. ensuring the sql queries don't become invalid.

## Architecture
The solution is implemented with a service oriented architecture consisting of 2 services (_api_ and _originator_).
The services communicate asynchronously via RabbitMQ.

_originator_ is a periodic job that runs every minute.
On each run it fetches the latest crypto market pricing data from the configured exchange.
The metrics are then published via RabbitMQ.

_api_ is the web service that the user can use to interact with the API.
It is backed by a PostgresQL data store that holds the _prices_ data model.
It consumes pricing metric data from RabbitMQ and saves it to the database.

## Potential enhancements
- Allow users to query metrics for arbitrary windows of time instead of a fixed _past 24 hours_.
- Provide users with summarized data per requested period of time e.g. candlestick data (ohlc).

## Scalability
### What would you change if you needed to track many metrics?
Separate out the logic for consuming metric data from the _api_ service and into a new service.
This new service could be horizontally scaled to support however many metrics we need to track.
Each metric will be published via its own routing key to RabbitMQ.
With this setup, adding more metrics to track would not impact the rate at which we ingest existing metrics.

### What if you needed to sample them more frequently?
We could segment the origination of the metrics across a few services.
This will reduce the amount of work the originator has to do on each sourcing cycle.
Each service could be equipped with their own API key to separate their rate limit concerns.
This lets us offer different frequencies for different groups of metrics.

### What if you had many users accessing your dashboard to view metrics?
We could implement a CQRS style pattern in a couple of ways.
- Replicate the datastore and have the _api_ read from the followers. The leader would then focus on writing the data into the datastore.
- We could split the data layer to optimize for reads. For example, index the data into a fast aggregator like Elasticsearch.

## Testing
- Craft integration test suites that exercise both services in concert.
- Implement a scheme to guard against compatibility drift of the data representation of the RabbitMQ message.
- Performance testing or exploration to understand at what limits the system either starts to breakdown or drop below our desired levels.

## Feature request
Since we already have a message consumer that's writing records into the _prices_ table, we can add another similar consumer called _monitor_.

_monitor_ will receive the same message and instead of inserting into the _prices_ table, it will use the timestamp of the price metric to query the database for the moving average of our instrument in the previous hour.
It can then compare the price value of the instrument it consumed with the computed average.
If that average is 3x larger or more, we publish a new message containing information about the instrument and the nature of the anomaly.
