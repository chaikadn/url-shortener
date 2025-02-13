package file

import (
	"encoding/json"
	"os"

	"github.com/chaikadn/url-shortener/internal/app/model"
)

type jsonDecoder struct {
	file    *os.File
	decoder *json.Decoder
}

func NewJSONDecoder(filename string) (*jsonDecoder, error) {
	// для увеличения производительности можно использовать буфер: bufio.NewReader(file)
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &jsonDecoder{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (d *jsonDecoder) ReadTo(entry *model.URLEntry) error {
	return d.decoder.Decode(entry)
}

func (d *jsonDecoder) Close() error {
	return d.file.Close()
}

type jsonEncoder struct {
	file    *os.File
	encoder *json.Encoder
}

func NewJSONEncoder(filename string) (*jsonEncoder, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &jsonEncoder{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (e *jsonEncoder) WriteFrom(entry *model.URLEntry) error {
	return e.encoder.Encode(entry)
}

func (e *jsonEncoder) Close() error {
	return e.file.Close()
}
