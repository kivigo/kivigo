package models

// Encoder defines how values are serialized and deserialized when stored or retrieved from a backend in KiviGo.
//
// KiviGo uses encoders to convert Go values (structs, strings, etc.) into a byte slice for storage,
// and to decode byte slices back into Go values when reading from the backend.
// This allows you to use different formats (JSON, YAML, etc.) or implement your own encoding logic.
//
// Example: using the JSON encoder, a struct will be marshaled to JSON before being saved in the database.
type Encoder interface {
	// Encode serializes the given value into a byte slice.
	//
	// Example:
	//   data, err := encoder.Encode("hello world")
	//   if err != nil {
	//       log.Fatal(err)
	//   }
	//   fmt.Println("Encoded:", data)
	Encode(v any) ([]byte, error)

	// Decode deserializes the given byte slice into the provided destination.
	//
	// Example:
	//   var s string
	//   err := encoder.Decode([]byte(`"hello world"`), &s)
	//   if err != nil {
	//       log.Fatal(err)
	//   }
	//   fmt.Println("Decoded:", s)
	Decode(data []byte, v any) error
}
