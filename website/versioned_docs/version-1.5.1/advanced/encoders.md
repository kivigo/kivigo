---
sidebar_position: 2
---

# Encoders and Decoders

KiviGo supports pluggable encoders that handle the serialization and deserialization of your data. This allows you to choose the best encoding format for your use case while maintaining the same API.

## Available Encoders

KiviGo comes with built-in support for several encoding formats:

### JSON Encoder (Default)

```go
import "github.com/kivigo/kivigo/pkg/encoder"

client, err := client.New(kvStore, client.Option{
    Encoder: encoder.JSON, // This is the default
})
```

### YAML Encoder

```go
import "github.com/kivigo/kivigo/pkg/encoder"

client, err := client.New(kvStore, client.Option{
    Encoder: encoder.YAML,
})
```

## Using Different Encoders

### JSON Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/kivigo/kivigo"
    "github.com/kivigo/kivigo/pkg/encoder"
    "github.com/kivigo/kivigo/backend/badger"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    kvStore, _ := badger.New(badger.DefaultOptions("./data"))
    defer kvStore.Close()
    
    // JSON encoder (default)
    jsonClient, err := client.New(kvStore, client.Option{
        Encoder: encoder.JSON,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    user := User{ID: 1, Name: "John", Email: "john@example.com"}
    
    // Store with JSON encoding
    err = jsonClient.Set(ctx, "user:1", user)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve with JSON decoding
    var retrievedUser User
    err = jsonClient.Get(ctx, "user:1", &retrievedUser)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Retrieved user: %+v\n", retrievedUser)
}
```

### YAML Example

```go
// YAML encoder
yamlClient, err := client.New(kvStore, client.Option{
    Encoder: encoder.YAML,
})

// Same API, different encoding
err = yamlClient.Set(ctx, "config", ConfigStruct{...})
```

## Creating Custom Encoders

You can create custom encoders by implementing the `Encoder` interface:

```go
package main

import (
    "encoding/xml"
    
    "github.com/kivigo/kivigo/pkg/models"
)

// Custom XML encoder
type XMLEncoder struct{}

func (e XMLEncoder) Encode(v interface{}) ([]byte, error) {
    return xml.Marshal(v)
}

func (e XMLEncoder) Decode(data []byte, v interface{}) error {
    return xml.Unmarshal(data, v)
}

// Ensure it implements the interface
var _ models.Encoder = (*XMLEncoder)(nil)

func main() {
    // Use custom XML encoder
    xmlClient, err := client.New(kvStore, client.Option{
        Encoder: XMLEncoder{},
    })
    
    // Use the same API with XML encoding
    err = xmlClient.Set(ctx, "data", MyStruct{...})
}
```

## Advanced Encoder Features

### Compression Encoder

Wrap existing encoders with compression:

```go
import (
    "bytes"
    "compress/gzip"
    "io"
)

type CompressedEncoder struct {
    inner models.Encoder
}

func (e CompressedEncoder) Encode(v interface{}) ([]byte, error) {
    // First encode with inner encoder
    data, err := e.inner.Encode(v)
    if err != nil {
        return nil, err
    }
    
    // Then compress
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    
    _, err = gz.Write(data)
    if err != nil {
        return nil, err
    }
    
    err = gz.Close()
    if err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (e CompressedEncoder) Decode(data []byte, v interface{}) error {
    // First decompress
    gz, err := gzip.NewReader(bytes.NewReader(data))
    if err != nil {
        return err
    }
    defer gz.Close()
    
    decompressed, err := io.ReadAll(gz)
    if err != nil {
        return err
    }
    
    // Then decode with inner encoder
    return e.inner.Decode(decompressed, v)
}

// Usage
compressedJSON := CompressedEncoder{inner: encoder.JSON}
client, err := client.New(kvStore, client.Option{
    Encoder: compressedJSON,
})
```

### Encryption Encoder

Add encryption to your data:

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

type EncryptedEncoder struct {
    inner models.Encoder
    gcm   cipher.AEAD
}

func NewEncryptedEncoder(inner models.Encoder, key []byte) (*EncryptedEncoder, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    return &EncryptedEncoder{
        inner: inner,
        gcm:   gcm,
    }, nil
}

func (e *EncryptedEncoder) Encode(v interface{}) ([]byte, error) {
    // First encode with inner encoder
    data, err := e.inner.Encode(v)
    if err != nil {
        return nil, err
    }
    
    // Create random nonce
    nonce := make([]byte, e.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    // Encrypt data
    ciphertext := e.gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}

func (e *EncryptedEncoder) Decode(data []byte, v interface{}) error {
    // Extract nonce and ciphertext
    nonceSize := e.gcm.NonceSize()
    if len(data) < nonceSize {
        return fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    
    // Decrypt
    plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return err
    }
    
    // Decode with inner encoder
    return e.inner.Decode(plaintext, v)
}

// Usage
key := []byte("32-byte-long-key-for-aes-256-gcm!") // 32 bytes for AES-256
encryptedJSON, err := NewEncryptedEncoder(encoder.JSON, key)
if err != nil {
    log.Fatal(err)
}

client, err := client.New(kvStore, client.Option{
    Encoder: encryptedJSON,
})
```

## Performance Considerations

### Encoding Benchmarks

Different encoders have different performance characteristics:

```go
func BenchmarkEncoders(b *testing.B) {
    data := LargeStruct{...}
    
    encoders := map[string]models.Encoder{
        "JSON": encoder.JSON,
        "YAML": encoder.YAML,
        "XML":  XMLEncoder{},
    }
    
    for name, enc := range encoders {
        b.Run(name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                encoded, _ := enc.Encode(data)
                var decoded LargeStruct
                enc.Decode(encoded, &decoded)
            }
        })
    }
}
```

### Choosing the Right Encoder

| Encoder | Speed | Size | Human Readable | Schema Evolution |
|---------|-------|------|----------------|------------------|
| JSON | Fast | Medium | ✅ | Good |
| YAML | Slower | Larger | ✅ | Good |
| XML | Slower | Largest | ✅ | Excellent |
| Binary | Fastest | Smallest | ❌ | Limited |
| Compressed | Medium | Smallest | ❌ | Same as inner |

## Error Handling

Handle encoding/decoding errors appropriately:

```go
import "github.com/kivigo/kivigo/pkg/errs"

func safeEncodeDecode() {
    var result MyStruct
    err := client.Get(ctx, "key", &result)
    
    if err != nil {
        switch {
        case errors.Is(err, errs.ErrNotFound):
            log.Println("Key not found")
        case errors.Is(err, errs.ErrInvalidData):
            log.Println("Data could not be decoded")
        default:
            log.Printf("Other error: %v", err)
        }
    }
}
```

## Best Practices

### 1. Consistent Encoding

Use the same encoder for a given key across your application:

```go
// Bad: Different encoders for same logical data
jsonClient.Set(ctx, "user:1", user)
yamlClient.Get(ctx, "user:1", &user) // Will fail!

// Good: Same encoder throughout
jsonClient.Set(ctx, "user:1", user)
jsonClient.Get(ctx, "user:1", &user) // Works!
```

### 2. Schema Evolution

Design your structs for backward compatibility:

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    
    // New fields should be optional with omitempty
    Phone string `json:"phone,omitempty"`
    Age   int    `json:"age,omitempty"`
}
```

### 3. Performance Testing

Always benchmark your encoder choice with real data:

```go
func TestEncoderPerformance(t *testing.T) {
    realData := getRealWorldData()
    
    start := time.Now()
    encoded, err := encoder.JSON.Encode(realData)
    encodeTime := time.Since(start)
    
    start = time.Now()
    var decoded MyStruct
    err = encoder.JSON.Decode(encoded, &decoded)
    decodeTime := time.Since(start)
    
    t.Logf("JSON: Encode=%v, Decode=%v, Size=%d bytes", 
           encodeTime, decodeTime, len(encoded))
}
```

### 4. Validation

Validate data after decoding:

```go
type ValidatedStruct struct {
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func (v *ValidatedStruct) Validate() error {
    if v.Email == "" {
        return errors.New("email is required")
    }
    if v.Age < 0 || v.Age > 150 {
        return errors.New("invalid age")
    }
    return nil
}

// Usage
var data ValidatedStruct
err := client.Get(ctx, "key", &data)
if err != nil {
    return err
}

if err := data.Validate(); err != nil {
    return fmt.Errorf("invalid data: %w", err)
}
```

Encoders provide flexibility in how your data is stored and transmitted. Choose the right encoder for your use case and always consider performance, size, and compatibility requirements.
