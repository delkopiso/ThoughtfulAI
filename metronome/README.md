# Running

Install the Go dependencies:

```
go get .
```

Run the server:

```
go run main.go
```

This will start the local server on port 8081.

# Debugging Interview Prompt

The `/usage` endpoint in this repo given a customer_id and start/end timestamps, returns per-hour counts of events for the customer in the specified time range. However, it seems to return incorrect values for some cases.

Your job is to edit the [`main.go`](./main.go) file, which currently produces [this output](./actual.json), so that it produces this [output](./expected.json) when sent the following curl command:

```bash
curl -X POST http://localhost:8081/usage \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "b4f9279a0196e40632e947dd1a88e857",
    "start_timestamp": "2021-03-02T12:06:00Z",
    "end_timestamp": "2021-03-02T18:06:00Z"
  }'
```
