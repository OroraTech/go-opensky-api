# OpenSky Network API

This repository contains a community API client implementation in Golang for the [OpenSky Network](https://opensky-network.org/).
It is used to retrieve live and historical details about aircraft positioning and flight information.

The library is based on the [REST API docs](https://opensky-network.org/apidoc/rest.html), and takes inspiration from the [official client](https://github.com/openskynetwork/opensky-api).

## Installation

```
go get github.com/ororatech/go-opensky-api
```

The library relies on the stdlib only, so no further dependencies are required.

## User Account

The client does not strictly require an account to use the OpenSky API. Username and password are, therefore, optional!

Refer to the [limitations](https://opensky-network.org/apidoc/rest.html#limitations), to see why/when a user account would be preferred.

## Usage

Create your API client:
```go
client := opensky.NewClient("myusername", "mypassword")
```

### Get States

```go
// Retrieve all states.
// This query may take a long time! Filtering via function parameters is recommended.
response, err := client.GetStates(time.Time{}, nil, nil)
if err != nil {
    // Something went wrong, check the error
}
fmt.Printf("received %d aircraft state objects", len(response.States))
for _, state := range response.States {
    // Check the contents of each received state
}
```

### Get Flights

```go
flights, err := client.GetFlights(time.Now().Add(-2*time.Hour), time.Now())
if err != nil {
    // Something went wrong, check the error
}
fmt.Printf("received %d flight objects", len(flights))
for _, flight := range flights {
	// Check the contents of each received flight
}
```
