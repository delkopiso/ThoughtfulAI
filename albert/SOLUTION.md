# Solution

## Stack choice
I went with Go as I'm fastest with it because I have pre-existing Go webapp templates I've been using for side projects in recent months.

## How to run
- `make dev`: pre-populates `.env` file.
- Set value of `ALBERT_API_KEY` in `.env` file to a valid API key.
- `make build`: builds the require docker images.
- `make up`: starts up all the service components.
- `make open-app`: opens the web browser into the app's URL.

## Service Components
- `web` is a simple http server that renders the user-facing app.
- `load-prices` is a worker process that run every `ALBERT_LOAD_PRICES_FREQUENCY` (default of 5 seconds). It checks the database for any securities without a price or with a price that has not been updated in `ALBERT_PRICE_MAX_AGE` (default of 5 seconds), and calls the `casestudy/stock/prices` endpoint to retrieve the latest price for those securities.
- `load-securities` is a worker process that fetches the list of available securities from the `casestudy/stock/tickers` endpoint and writes them to the securities database table. It runs every `ALBERT_LOAD_STOCK_FREQUENCY` (default of 24 hours).

## Design Choices
- The simplest way to build this watchlist would be to bundle the behavior of the workers with the webapp, but that approach will prevent the ability to scale the webapp in order to adjust for any increased load from the users. Since the easiest way to scale to meet increased demand (e.g. more concurrent users) would be to run multiple instances of the webapp, we don't want to have each webapp instance running its own cron for populating the tickers or the prices.
- We use 2 separate worker processes because the data they process differ in use and context. The ticker data is much less likely to go stale than the price data, so we can have the worker that loads and processes tickers run more infrequently than the worker that loads and processes prices.
