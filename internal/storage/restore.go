package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/7StaSH7/gometrics/internal/model"
)

func (s *MemStorage) Restore() error {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}
	defer file.Close()

	r := bufio.NewReader(file)

	data, _, err := r.ReadLine()
	if err != nil {
		if err != io.EOF {
			return err
		}
	}

	metrics := []model.Metrics{}
	json.Unmarshal(data, &metrics)

	for _, metric := range metrics {
		switch metric.MType {
		case model.Counter:
			{
				s.counter[metric.ID] = *metric.Delta
			}
		case model.Gauge:
			{
				s.gauges[metric.ID] = *metric.Value
			}
		}
	}

	return nil
}
