package bluez

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/ugorji/go/codec"
)

// Resolver holds an encoder and decoder.
type Resolver struct {
	check bool

	encoder *codec.Encoder
	decoder *codec.Decoder
	data    []byte

	lock sync.Mutex
}

var resolver Resolver

// DecodeVariantMap decodes a map of variants into the provided data.
func DecodeVariantMap(
	variants map[string]dbus.Variant, data interface{},
	checkProps ...string,
) error {
	resolver.lock.Lock()
	defer resolver.lock.Unlock()

	if !resolver.check {
		resolver.encoder = codec.NewEncoderBytes(&resolver.data, &codec.SimpleHandle{})
		resolver.decoder = codec.NewDecoderBytes(resolver.data, &codec.SimpleHandle{})

		resolver.check = true
	}

	props := make(map[string]interface{}, len(variants))
	for key, value := range variants {
		for _, prop := range checkProps {
			if prop == key && value.Signature().Empty() {
				return fmt.Errorf("No signature found for property '%s'", prop)
			}
		}

		val := value.Value()
		if v, ok := val.(dbus.ObjectPath); ok {
			val = string(v)
		}

		props[key] = val
	}

	resolver.encoder.ResetBytes(&resolver.data)
	if err := resolver.encoder.Encode(props); err != nil {
		return err
	}

	resolver.decoder.ResetBytes(resolver.data)
	return resolver.decoder.Decode(data)
}
