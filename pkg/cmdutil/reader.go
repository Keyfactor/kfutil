package cmdutil

import (
	"bytes"
	"log"
	"net/url"
	"os"
)

// ReaderBuilder is a type that aggregates multiple readers and
// sequentially reads data from each one when its Read() method is called.
type ReaderBuilder struct {
	readers []Reader // List of readers to fetch data from.
}

// Reader is an interface for types that can produce byte slices.
type Reader interface {
	Read() ([]byte, error) // Read should return the data as a byte slice.
}

// NewReaderBuilder initializes a new empty ReaderBuilder.
func NewReaderBuilder() *ReaderBuilder {
	return &ReaderBuilder{}
}

// Read fetches data from all added readers and combines it.
func (b *ReaderBuilder) Read() ([]byte, error) {
	var readBytes []byte
	for _, r := range b.readers {
		read, err := r.Read()
		if err != nil {
			return nil, err // Return the error immediately if any reader fails.
		}

		// Append data from the current reader to the combined result.
		readBytes = append(readBytes, read...)
	}
	return readBytes, nil
}

// Stdin adds a standard input reader to the builder.
func (b *ReaderBuilder) Stdin() {
	b.readers = append(b.readers, &StdinReader{})
}

// URL adds one or more URL readers to the builder.
func (b *ReaderBuilder) URL(urls ...*url.URL) {
	for _, u := range urls {
		b.readers = append(b.readers, &URLReader{
			url: u,
		})
	}
}

// Path adds one or more path readers to the builder.
func (b *ReaderBuilder) Path(paths ...string) {
	for _, path := range paths {
		b.readers = append(b.readers, &PathReader{
			path: path,
		})
	}
}

// URLReader is a type that fetches data from a URL.
type URLReader struct {
	url *url.URL
}

// Read fetches data from the URL using a simple REST client.
func (r *URLReader) Read() ([]byte, error) {
	return NewSimpleRestClient().Get(r.url.String())
}

// PathReader is a type that fetches data from a local file.
type PathReader struct {
	path string
}

// Read reads data from a local file specified by path.
func (r *PathReader) Read() ([]byte, error) {
	return os.ReadFile(r.path)
}

// StdinReader is a type that reads data from the standard input.
type StdinReader struct {
}

// Read reads data from the standard input until EOF is reached.
func (r *StdinReader) Read() ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(os.Stdin)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Read %d bytes from stdin:\n%s", buf.Len(), buf.String())
	return buf.Bytes(), nil
}
