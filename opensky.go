package opensky

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Origin of a state's position.
type PositionSource int

const (
	baseOpenSkyURL = "https://opensky-network.org/api"

	ADSB    PositionSource = 0
	ASTERIX PositionSource = 1
	MLAT    PositionSource = 2
	FLARM   PositionSource = 3
)

// Represents the state of a vehicle at a particular time.
//
// All pointer fields are nullable, therefore checks are required, before accessing those fields.
type State struct {
	ICAO24             string         `json:"icao24"`                  // ICAO24 address of the transmitter in hex string representation.
	CallSign           string         `json:"callsign,omitempty"`      // CallSign of the vehicle. Can be nil if no callsign has been received.
	OriginCountry      string         `json:"origin_country"`          // Inferred through the ICAO24 address.
	TimePosition       *UnixTime      `json:"time_position,omitempty"` // UnixTime of last position report. Can be nil if there was no position report received by OpenSky within 15s before.
	LastContact        UnixTime       `json:"last_contact"`            // UnixTime of last received message from this transponder.
	Longitude          *float64       `json:"longitude,omitempty"`     // In ellipsoidal coordinates (WGS-84) and degrees. Can be nil.
	Latitude           *float64       `json:"latitude,omitempty"`      // In ellipsoidal coordinates (WGS-84) and degrees. Can be nil.
	GeoAltitude        *float64       `json:"geo_altitude,omitempty"`  // Geometric altitude in meters. Can be nil.
	OnGround           bool           `json:"on_ground"`               // True if aircraft is on ground (sends ADS-B surface position reports).
	Velocity           *float64       `json:"velocity,omitempty"`      // Velocity over ground in m/s. Can be nil if information not present.
	Heading            *float64       `json:"heading,omitempty"`       // Heading in decimal degrees (0 is north). Can be nil if information not present.
	VerticalRate       *float64       `json:"vertical_rate,omitempty"` // In m/s, incline is positive, decline negative. Can be nil if information not present.
	Sensors            []int          `json:"sensors,omitempty"`       // Serial numbers of sensors which received messages from the vehicle within the validity period of this state vector. Can be nil if no filtering for sensor has been requested.
	BarometricAltitude *float64       `json:"baro_altitude,omitempty"` // Barometric altitude in meters. Can be nil.
	Squawk             string         `json:"squawk,omitempty"`        // Transponder code aka Squawk. Can be empty.
	Spi                bool           `json:"spi"`                     // Special purpose indicator.
	PositionSource     PositionSource `json:"position_source"`         // Origin of this state’s position.
}

// Represents a single flight of an aircraft.
type Flight struct {
	ICAO24                           string   `json:"icao24"`                           // ICAO24 address of the transmitter in hex string representation.
	FirstSeen                        UnixTime `json:"firstSeen"`                        // Estimated time of departure for the flight.
	EstDepartureAirport              string   `json:"estDepartureAirport,omitempty"`    // ICAO code of the estimated departure airport. Can be nil if the airport could not be identified.
	LastSeen                         UnixTime `json:"lastSeen"`                         // Estimated time of arrival for the flight.
	EstArrivalAirport                string   `json:"estArrivalAirport,omitempty"`      // ICAO code of the estimated arrival airport. Can be nil if the airport could not be identified.
	CallSign                         string   `json:"callsign,omitempty"`               // CallSign of the vehicle. Can be nil if no callsign has been received.
	EstDepartureAirportHorizDistance int      `json:"estDepartureAirportHorizDistance"` // Horizontal distance of the last received airborne position to the estimated departure airport in meters.
	EstDepartureAirportVertDistance  int      `json:"estDepartureAirportVertDistance"`  // Vertical distance of the last received airborne position to the estimated departure airport in meters.
	EstArrivalAirportHorizDistance   int      `json:"estArrivalAirportHorizDistance"`   // Horizontal distance of the last received airborne position to the estimated arrival airport in meters.
	EstArrivalAirportVertDistance    int      `json:"estArrivalAirportVertDistance"`    // Vertical distance of the last received airborne position to the estimated arrival airport in meters.
	DepartureAirportCandidatesCount  int      `json:"departureAirportCandidatesCount"`  // Number of other possible departure airports. These are airports in short distance to EstDepartureAirport.
	ArrivalAirportCandidatesCount    int      `json:"arrivalAirportCandidatesCount"`    // Number of other possible departure airports. These are airports in short distance to EstArrivalAirport.
}

// Bounding box of WGS84 coordinates.
type BoundingBox struct {
	LatMin float64 `json:"lamin"` // Lower bound for the latitude in decimal degrees.
	LonMin float64 `json:"lomin"` // Lower bound for the longitude in decimal degrees.
	LatMax float64 `json:"lamax"` // Upper bound for the latitude in decimal degrees.
	LonMax float64 `json:"lomax"` // Upper bound for the longitude in decimal degrees.
}

// An OpenSky API client.
// To instantiate a new client, use the NewClient function.
type Client struct {
	username   string
	password   string
	httpClient http.Client
}

// Unstructured raw response for state queries.
type unstructuredStateResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}

// The response for state vectors.
type GetStatesResponse struct {
	Time   time.Time `json:"time"`
	States []State   `json:"states"`
}

// Creates a new OpenSky client.
// Username and password fields are optional.
func NewClient(username string, password string) *Client {
	return &Client{
		username: username,
		password: password,
		httpClient: http.Client{
			Timeout: time.Minute * 5,
		},
	}
}

// Creates a new HTTP request, with the basic authentication header already set.
func (c *Client) newRequest(method string, apiURL string) (request *http.Request, err error) {
	request, err = http.NewRequest(method, apiURL, nil)
	if err != nil {
		return
	}
	if request != nil && c.username != "" && c.password != "" {
		request.SetBasicAuth(c.username, c.password)
	}
	return
}

// doHTTP is a utility method for performing an HTTP request and parsing the
// JSON response inside the passed responseObject.
//
// If the operation fails for any reason, an error is returned.
// If the HTTP request returns any status code other than 200, an error is returned.
func (c *Client) doHTTP(request *http.Request, responseObject interface{}) (err error) {
	var resp *http.Response
	resp, err = c.httpClient.Do(request)
	if err != nil {
		return
	}
	// Parse response
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%d: %v", resp.StatusCode, string(body))
		return
	}
	// Parse JSON
	err = json.Unmarshal(body, responseObject)
	if err != nil {
		return
	}
	return nil
}

// Retrieves any state vectors from OpenSky, at the specified timestamp and
// according to the additional optional filters.
//
// If time.Time{} is passed as a parameter, then the current time will be used instead.
//
// One or more ICAO24 addresses may be passed, to filter only for the specified aircraft.
// By default, the state vectors of all aircraft are returned.
//
// If a bounding box is passed, then only the specified area will be queried.
func (c *Client) GetStates(time time.Time, icao24 []string, bbox *BoundingBox) (response GetStatesResponse, err error) {
	request, err := c.newRequest("GET", fmt.Sprintf("%s/states/all", baseOpenSkyURL))
	if err != nil {
		return
	}
	q := request.URL.Query()
	//request := c.baseRequest.Clone().
	//	Get(fmt.Sprintf("%s/states/all", baseOpenSkyURL))
	// Add optional parameters
	if !time.IsZero() {
		q.Set("time", fmt.Sprintf("%v", time.Unix()))
	}
	if icao24 != nil && len(icao24) > 0 {
		q.Set("icao24", strings.Join(icao24, ","))
	}
	if bbox != nil {
		q.Set("lamin", fmt.Sprintf("%v", bbox.LatMin))
		q.Set("lomin", fmt.Sprintf("%v", bbox.LonMin))
		q.Set("lamax", fmt.Sprintf("%v", bbox.LatMax))
		q.Set("lomax", fmt.Sprintf("%v", bbox.LonMax))
	}
	request.URL.RawQuery = q.Encode()
	// Fetch response
	var rawResponse unstructuredStateResponse
	err = c.doHTTP(request, &rawResponse)
	if err != nil {
		return
	}
	return parseStatesResponse(rawResponse)
}

// Retrieves state vectors from OpenSky for your own sensors (without rate limitations),
// at the specified timestamp and according to the additional optional filters.
//
// If time.Time{} is passed as a parameter, then the current time will be used instead.
//
// One or more ICAO24 addresses may be passed, to filter only for the specified aircraft.
// By default, the state vectors of all aircraft are returned.
//
// You may retrieve the states of only a subset of your receivers, by passing the serial
// parameter. In this case, the API returns states of aircraft that are visible to at
// least one of the given receivers.
func (c *Client) GetOwnStates(time time.Time, icao24 []string, serials []int) (response GetStatesResponse, err error) {
	request, err := c.newRequest("GET", fmt.Sprintf("%s/states/own", baseOpenSkyURL))
	if err != nil {
		return
	}
	q := request.URL.Query()
	// Add optional parameters
	if !time.IsZero() {
		q.Set("time", fmt.Sprintf("%v", time.Unix()))
	}
	if icao24 != nil && len(icao24) > 0 {
		q.Set("icao24", strings.Join(icao24, ","))
	}
	if serials != nil && len(serials) > 0 {
		serialQuery := ""
		for i, s := range serials {
			if i > 0 {
				serialQuery += ","
			}
			serialQuery += fmt.Sprintf("%v", s)
		}
		q.Set("serials", serialQuery)
	}
	request.URL.RawQuery = q.Encode()
	// Fetch response
	var rawResponse unstructuredStateResponse
	err = c.doHTTP(request, &rawResponse)
	if err != nil {
		return
	}
	return parseStatesResponse(rawResponse)
}

// Retrieves all flight information within a certain time interval.
// Flights departed and arrived within the [begin, end] boundaries will be returned.
//
// If no flights were found for the given time period, a 404 error will be returned instead.
func (c *Client) GetFlights(begin time.Time, end time.Time) (flights []Flight, err error) {
	request, err := c.newRequest("GET", fmt.Sprintf("%s/flights/all", baseOpenSkyURL))
	if err != nil {
		return
	}
	q := request.URL.Query()
	// Add optional parameters
	if !begin.IsZero() {
		q.Set("begin", fmt.Sprintf("%v", begin.Unix()))
	}
	if !end.IsZero() {
		q.Set("end", fmt.Sprintf("%v", end.Unix()))
	}
	request.URL.RawQuery = q.Encode()
	// Fetch response
	err = c.doHTTP(request, &flights)
	return
}

// Retrieves flight information for a particular aircraft, identified by the icao24 address parameter,
// within a certain time interval.
// Flights departed and arrived within the [begin, end] boundaries will be returned.
//
// If no flights were found for the given time period, a 404 error will be returned instead.
func (c *Client) GetFlightsByAircraft(icao24 string, begin time.Time, end time.Time) (flights []Flight, err error) {
	request, err := c.newRequest("GET", fmt.Sprintf("%s/flights/aircraft", baseOpenSkyURL))
	if err != nil {
		return
	}
	q := request.URL.Query()
	// Add optional parameters
	if !begin.IsZero() {
		q.Set("begin", fmt.Sprintf("%v", begin.Unix()))
	}
	if !end.IsZero() {
		q.Set("end", fmt.Sprintf("%v", end.Unix()))
	}
	if icao24 != "" {
		q.Set("icao24", icao24)
	}
	request.URL.RawQuery = q.Encode()
	// Fetch response
	err = c.doHTTP(request, &flights)
	return
}

// Parse a single state array from an unstructured states response.
// The i parameter represents the index of the state element in the states response.
func parseState(s []interface{}, i int) (state State, err error) {
	if len(s) < 17 {
		err = fmt.Errorf("invalid state object at position %v: response contains %v values, expected 17", i, len(s))
		return
	}
	// icao24
	icao24, ok := s[0].(string)
	if !ok {
		err = fmt.Errorf("invalid icao24 value at position %d: %v", i, s[0])
		return
	}
	// callsign
	var callsign string
	if s[1] != nil {
		callsign, ok = s[1].(string)
		if !ok {
			err = fmt.Errorf("invalid callsign value at position %d: %v", i, s[1])
			return
		}
	}
	// origin_country
	originCountry, ok := s[2].(string)
	if !ok {
		err = fmt.Errorf("invalid origin_country value at position %d: %v", i, s[2])
		return
	}
	// time_position
	var rawTimePosition int64
	var timePosition *UnixTime
	if s[3] != nil {
		rawTimePosition, err = jsonNumberToInt(s[3])
		if err != nil {
			err = fmt.Errorf("invalid time_position value at position %d: %w", i, err)
			return
		}
		unixTime := newUnixTime(rawTimePosition)
		timePosition = &unixTime
	}
	// last_contact
	var lastContact int64
	lastContact, err = jsonNumberToInt(s[4])
	if err != nil {
		err = fmt.Errorf("invalid last_contact value at position %d: %w", i, err)
		return
	}
	// longitude
	var lon *float64
	if rawLon, ok := s[5].(float64); ok {
		lon = &rawLon
	}
	// latitude
	var lat *float64
	if rawLat, ok := s[6].(float64); ok {
		lat = &rawLat
	}
	// baro_altitude
	var baroAltitude *float64
	if rawBaroAltitude, ok := s[7].(float64); ok {
		baroAltitude = &rawBaroAltitude
	}
	// on_ground
	onGround, ok := s[8].(bool)
	if !ok {
		err = fmt.Errorf("invalid on_ground value at position %d: %v", i, s[8])
		return
	}
	// velocity
	var velocity *float64
	if rawVelocity, ok := s[9].(float64); ok {
		velocity = &rawVelocity
	}
	// true_track
	var trueTrack *float64
	if rawTrueTrack, ok := s[10].(float64); ok {
		trueTrack = &rawTrueTrack
	}
	// vertical_rate
	var verticalRate *float64
	if rawVerticalRate, ok := s[11].(float64); ok {
		verticalRate = &rawVerticalRate
	}
	// sensors
	var sensors []int
	if s[12] != nil {
		sensors, err = jsonNumberArrayToIntArray(s[12])
		if err != nil {
			err = fmt.Errorf("invalid sensors value at position %d: %w", i, err)
			return
		}
	}
	// geo_altitude
	var geoAltitude *float64
	if rawGeoAltitude, ok := s[13].(float64); ok {
		geoAltitude = &rawGeoAltitude
	}
	// squawk
	var squawk string
	if s[14] != nil {
		squawk, ok = s[14].(string)
		if !ok {
			err = fmt.Errorf("invalid squawk value at position %d: %v", i, s[14])
			return
		}
	}
	// spi
	spi, ok := s[15].(bool)
	if !ok {
		err = fmt.Errorf("invalid spi value at position %d: %v", i, s[15])
		return
	}
	// position_source
	var positionSource int64
	positionSource, err = jsonNumberToInt(s[16])
	if err != nil {
		err = fmt.Errorf("invalid position_source value at position %d: %w", i, err)
		return
	}
	// Set state values
	state = State{
		ICAO24:             icao24,
		CallSign:           callsign,
		OriginCountry:      originCountry,
		TimePosition:       timePosition,
		LastContact:        newUnixTime(lastContact),
		Longitude:          lon,
		Latitude:           lat,
		GeoAltitude:        geoAltitude,
		OnGround:           onGround,
		Velocity:           velocity,
		Heading:            trueTrack,
		VerticalRate:       verticalRate,
		Sensors:            sensors,
		BarometricAltitude: baroAltitude,
		Squawk:             squawk,
		Spi:                spi,
		PositionSource:     PositionSource(positionSource),
	}
	return
}

// Parses an unstructured state response.
func parseStatesResponse(rawResponse unstructuredStateResponse) (response GetStatesResponse, err error) {
	response.Time = time.Unix(rawResponse.Time, 0)
	// Parse state vectors
	for i, s := range rawResponse.States {
		var state State
		state, err = parseState(s, i)
		if err != nil {
			return
		}
		// Add state
		response.States = append(response.States, state)
	}
	return
}

// Helper function to convert a number received in a json object to an int64 type.
// Throws an error, if the number could not be parsed.
func jsonNumberToInt(val interface{}) (i int64, err error) {
	fVal, ok := val.(float64)
	if !ok {
		err = fmt.Errorf("couldn't parse %v as number", val)
		return
	}
	i = int64(fVal)
	return
}

// Helper function to convert a number array received in a json object to an []int type.
// Throws an error, if the value could not be parsed as a number array.
func jsonNumberArrayToIntArray(val interface{}) (a []int, err error) {
	aVal, ok := val.([]float64)
	if !ok {
		err = fmt.Errorf("couldn't parse %v as number array", val)
		return
	}
	for _, v := range aVal {
		a = append(a, int(v))
	}
	return
}
