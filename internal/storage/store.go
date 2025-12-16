package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/7StaSH7/gometrics/internal/model"
)

func (s *MemStorage) Store() error {
	totalMetrics := len(s.gauges) + len(s.counter)
	metrics := make([]model.Metrics, 0, totalMetrics)

	for name, value := range s.gauges {
		v := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &v,
		})
	}
	for name, value := range s.counter {
		d := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: model.Counter,
			Delta: &d,
		})
	}

	if err := s.write(metrics); err != nil {
		return err
	}

	return nil
}

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

func (s *MemStorage) write(metrics []model.Metrics) error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}
	defer file.Close()

	r := bufio.NewWriter(file)

	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	if _, err := r.Write(data); err != nil {
		return err
	}

	if err := r.Flush(); err != nil {
		return err
	}

	return nil
}
