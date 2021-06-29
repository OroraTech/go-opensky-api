package opensky

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newFloat(f float64) *float64 {
	return &f
}

func TestUnmarshalUnixTime(t *testing.T) {
	type wrapper struct {
		Time UnixTime `json:"time"`
	}
	type testCase struct {
		jsonString    string
		expectedTime  UnixTime
		expectedError bool
	}
	cases := []testCase{
		{`{"time":1624891429}`, UnixTime{Time: time.Unix(1624891429, 0)}, false},
		{`{"time":"string"}`, UnixTime{Time: time.Time{}}, true},
		{`{"time":null}`, UnixTime{}, false},
		{`{"time":{}}`, UnixTime{}, true},
	}
	for _, c := range cases {
		b := []byte(c.jsonString)
		var w wrapper
		err := json.Unmarshal(b, &w)
		if c.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, c.expectedTime, w.Time)
		}
	}
}

func TestJsonNumberToInt(t *testing.T) {
	type testCase struct {
		value         interface{}
		expectedValue int64
		expectedError bool
	}
	cases := []testCase{
		{42.0, 42, false},
		{-1.0, -1, false},
		{0.0, 0, false},
		{2.99, 2, false},
		{"foo", 0, true},
		{true, 0, true},
		{[]float64{1, 3, 5}, 0, true},
	}
	for _, c := range cases {
		i, err := jsonNumberToInt(c.value)
		assert.Equal(t, c.expectedValue, i)
		if c.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestJsonNumberArrayToIntArray(t *testing.T) {
	type testCase struct {
		value         interface{}
		expectedValue []int
		expectedError bool
	}
	cases := []testCase{
		{[]float64{42.0, 33.0, 12.95, -2.3}, []int{42, 33, 12, -2}, false},
		{[]float64{1, 2, 100, -100}, []int{1, 2, 100, -100}, false},
		{1.0, nil, true},
		{[]int{1, 2, 100, -100}, nil, true},
		{"foo", nil, true},
		{true, nil, true},
	}
	for _, c := range cases {
		i, err := jsonNumberArrayToIntArray(c.value)
		assert.Equal(t, c.expectedValue, i)
		if c.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestParseState(t *testing.T) {
	type testCase struct {
		raw            []interface{}
		expectedResult State
		expectedError  bool
	}
	cases := []testCase{
		{
			// All optional values are filled -> OK
			[]interface{}{
				"ae1fa7",
				"TALON71 ",
				"United States",
				float64(1624891429),
				float64(1624891429),
				-116.2121,
				43.5431,
				914.4,
				false,
				17.95,
				117.3,
				-1.3,
				[]float64{1000, 1042},
				952.5,
				"0753",
				false,
				float64(0),
			},
			State{
				ICAO24:             "ae1fa7",
				CallSign:           "TALON71 ",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          newFloat(-116.2121),
				Latitude:           newFloat(43.5431),
				BarometricAltitude: newFloat(914.4),
				OnGround:           false,
				Velocity:           newFloat(17.95),
				Heading:            newFloat(117.3),
				VerticalRate:       newFloat(-1.3),
				Sensors:            []int{1000, 1042},
				GeoAltitude:        newFloat(952.5),
				Squawk:             "0753",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// All optional values are nil -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// icao24 is invalid -> Error
			[]interface{}{
				666,
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// callsign is invalid -> Error
			[]interface{}{
				"a50c7c",
				666,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// origin_country is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				666,
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// time_position is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				"invalid_time",
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// last_contact is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				"invalid_time",
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// longitude is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				"invalid_long",
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// latitude is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				"invalid_lat",
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// baro_altitude is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				"invalid_baro_altitude",
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// on_ground is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				666,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// velocity is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				"invalid_velocity",
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// heading is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				"invalid_heading",
				nil,
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// vertical_rate is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				"invalid_vertical_rate",
				nil,
				nil,
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// sensors is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				[]string{"invalid", "sensors"},
				nil,
				nil,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// geo_altitude is invalid -> ignored -> OK
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				"invalid_geo_altitude",
				nil,
				false,
				float64(0),
			},
			State{
				ICAO24:             "a50c7c",
				CallSign:           "",
				OriginCountry:      "United States",
				TimePosition:       newUnixTimeP(1624891429),
				LastContact:        newUnixTime(1624891429),
				Longitude:          nil,
				Latitude:           nil,
				BarometricAltitude: nil,
				OnGround:           false,
				Velocity:           nil,
				Heading:            nil,
				VerticalRate:       nil,
				Sensors:            nil,
				GeoAltitude:        nil,
				Squawk:             "",
				Spi:                false,
				PositionSource:     ADSB,
			},
			false,
		},
		{
			// squawk is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				666,
				false,
				float64(0),
			},
			State{},
			true,
		},
		{
			// spi is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				666,
				float64(0),
			},
			State{},
			true,
		},
		{
			// position_source is invalid -> Error
			[]interface{}{
				"a50c7c",
				nil,
				"United States",
				float64(1624891429),
				float64(1624891429),
				nil,
				nil,
				nil,
				false,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				666,
			},
			State{},
			true,
		},
	}
	for i, c := range cases {
		state, err := parseState(c.raw, i)
		assert.Equal(t, c.expectedResult, state)
		if c.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestParseStatesResponse(t *testing.T) {
	type testCase struct {
		raw            unstructuredStateResponse
		expectedResult GetStatesResponse
		expectedError  bool
	}
	cases := []testCase{
		{
			// All cases are valid
			unstructuredStateResponse{Time: 1624958210, States: [][]interface{}{
				{
					"ae1fa7",
					"TALON71 ",
					"United States",
					float64(1624891429),
					float64(1624891429),
					-116.2121,
					43.5431,
					914.4,
					false,
					17.95,
					117.3,
					-1.3,
					[]float64{1000, 1042},
					952.5,
					"0753",
					false,
					float64(0),
				},
				{
					"a50c7c",
					nil,
					"United States",
					float64(1624891429),
					float64(1624891429),
					nil,
					nil,
					nil,
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					false,
					float64(0),
				},
			}},
			GetStatesResponse{Time: time.Unix(1624958210, 0), States: []State{
				{
					ICAO24:             "ae1fa7",
					CallSign:           "TALON71 ",
					OriginCountry:      "United States",
					TimePosition:       newUnixTimeP(1624891429),
					LastContact:        newUnixTime(1624891429),
					Longitude:          newFloat(-116.2121),
					Latitude:           newFloat(43.5431),
					BarometricAltitude: newFloat(914.4),
					OnGround:           false,
					Velocity:           newFloat(17.95),
					Heading:            newFloat(117.3),
					VerticalRate:       newFloat(-1.3),
					Sensors:            []int{1000, 1042},
					GeoAltitude:        newFloat(952.5),
					Squawk:             "0753",
					Spi:                false,
					PositionSource:     ADSB,
				},
				{
					ICAO24:             "a50c7c",
					CallSign:           "",
					OriginCountry:      "United States",
					TimePosition:       newUnixTimeP(1624891429),
					LastContact:        newUnixTime(1624891429),
					Longitude:          nil,
					Latitude:           nil,
					BarometricAltitude: nil,
					OnGround:           false,
					Velocity:           nil,
					Heading:            nil,
					VerticalRate:       nil,
					Sensors:            nil,
					GeoAltitude:        nil,
					Squawk:             "",
					Spi:                false,
					PositionSource:     ADSB,
				},
			}},
			false,
		},
		{
			// Invalid field causes error -> no states result
			unstructuredStateResponse{Time: 1624958210, States: [][]interface{}{
				{
					"a50c7c",
					nil,
					"United States",
					float64(1624891429),
					float64(1624891429),
					nil,
					nil,
					nil,
					false,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					"invalid_spi",
					float64(0),
				},
			}},
			GetStatesResponse{Time: time.Unix(1624958210, 0), States: nil},
			true,
		},
		{
			// Empty states -> OK
			unstructuredStateResponse{Time: 1624958210, States: [][]interface{}{}},
			GetStatesResponse{Time: time.Unix(1624958210, 0), States: nil},
			false,
		},
		{
			// Empty result -> OK
			unstructuredStateResponse{Time: 0, States: nil},
			GetStatesResponse{Time: time.Unix(0, 0), States: nil},
			false,
		},
	}
	// Run tests
	for _, c := range cases {
		result, err := parseStatesResponse(c.raw)
		assert.Equal(t, c.expectedResult, result)
		if c.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func BenchmarkParseStatesResponse(b *testing.B) {
	rawResponse := unstructuredStateResponse{Time: 1624958210, States: [][]interface{}{
		{
			"ae1fa7",
			"TALON71 ",
			"United States",
			float64(1624891429),
			float64(1624891429),
			-116.2121,
			43.5431,
			914.4,
			false,
			17.95,
			117.3,
			-1.3,
			[]float64{1000, 1042},
			952.5,
			"0753",
			false,
			float64(0),
		},
		{
			"a50c7c",
			nil,
			"United States",
			float64(1624891429),
			float64(1624891429),
			nil,
			nil,
			nil,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			false,
			float64(0),
		},
	}}
	_, _ = parseStatesResponse(rawResponse)
}

func TestApi(t *testing.T) {
	client := NewClient("", "")
	// Test several API invocations
	_, err := client.GetStates(time.Time{}, []string{"ae1fa7", "a50c7c"}, nil)
	assert.NoError(t, err)
	_, err = client.GetFlightsByAircraft("a50c7c", time.Now().Add(-24*5*time.Hour), time.Now())
	assert.NoError(t, err)
}
