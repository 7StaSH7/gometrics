package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/7StaSH7/gometrics/internal/model"
)

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
