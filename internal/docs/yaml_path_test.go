package docs_test

import (
	"testing"

	"github.com/Jeffail/benthos/v3/lib/config"
	"github.com/Jeffail/gabs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	_ "github.com/Jeffail/benthos/v3/public/components/all"
)

func TestSetYAMLPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		path        string
		value       string
		output      string
		errContains string
	}{
		{
			name: "set input",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
    topics: [ "baz" ]

output:
  nats:
    urls: [ nats://127.0.0.1:4222 ]
    subject: benthos_messages
    max_in_flight: 1
`,
			path: "/input",
			value: `
bloblang:
  mapping: 'root = {"foo":"bar"}'`,
			output: `
input:
  bloblang:
    mapping: 'root = {"foo":"bar"}'
output:
  nats:
    urls: [ nats://127.0.0.1:4222 ]
    subject: benthos_messages
    max_in_flight: 1
`,
		},
		{
			name: "set input addresses total",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
    topics: [ "baz" ]
`,
			path:  "/input/kafka/addresses",
			value: `"foobar"`,
			output: `
input:
  kafka:
    addresses: [ "foobar" ]
    topics: [ "baz" ]
`,
		},
		{
			name:  "set value to object",
			input: `input: "hello world"`,
			path:  "/input/kafka/addresses",
			value: `"foobar"`,
			output: `
input:
  kafka:
    addresses: ["foobar"]
`,
		},
		{
			name: "set array index",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
    topics: [ "baz" ]
`,
			path:  "/input/kafka/addresses/0",
			value: `"baz"`,
			output: `
input:
  kafka:
    addresses: [ "baz", "bar" ]
    topics: [ "baz" ]
`,
		},
		{
			name: "set array index child",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
    topics: [ "baz" ]
  processors:
    - compress:
        algorithm: gzip
`,
			path:  "/input/processors/0/compress/algorithm",
			value: `"baz"`,
			output: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
    topics: [ "baz" ]
  processors:
    - compress:
        algorithm: baz
`,
		},
		{
			name: "set array append",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
    topics: [ "baz" ]
`,
			path:  "/input/kafka/addresses/-",
			value: `"baz"`,
			output: `
input:
  kafka:
    addresses: [ "foo", "bar", "baz" ]
    topics: [ "baz" ]
`,
		},
		{
			name: "set array NaN",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
`,
			path:        "/input/kafka/addresses/nope",
			value:       `"baz"`,
			errContains: "input.kafka.addresses.nope: failed to parse path segment as array index",
		},
		{
			name: "set array big index",
			input: `
input:
  kafka:
    addresses: [ "foo", "bar" ]
`,
			path:        "/input/kafka/addresses/2",
			value:       `"baz"`,
			errContains: "input.kafka.addresses.2: target index greater than",
		},
		{
			name: "set nested array big index",
			input: `
input:
  kafka:
    addresses: [ [ "foo", "bar" ] ]
`,
			path:        "/input/kafka/addresses/0/2",
			value:       `"baz"`,
			errContains: "input.kafka.addresses.0.2: field not recognised",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var input, value yaml.Node

			require.NoError(t, yaml.Unmarshal([]byte(test.input), &input))
			require.NoError(t, yaml.Unmarshal([]byte(test.value), &value))

			path, err := gabs.JSONPointerToSlice(test.path)
			require.NoError(t, err)

			err = config.Spec().SetYAMLPath(nil, &input, &value, path...)
			if len(test.errContains) > 0 {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				require.NoError(t, err)

				var iinput, ioutput interface{}
				require.NoError(t, input.Decode(&iinput))
				require.NoError(t, yaml.Unmarshal([]byte(test.output), &ioutput))
				assert.Equal(t, ioutput, iinput)
			}
		})
	}
}
