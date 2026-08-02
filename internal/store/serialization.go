package store

import (
	"encoding/binary"
	"fmt"

	"github.com/SaDMikaSa/UPass/internal/domain"
)

func marshalVaultBinary(records map[string]domain.Record) ([]byte, error) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(records)))

	for _, rec := range records {
		buf = appendBinaryField(buf, rec.Service)
		buf = appendBinaryField(buf, rec.Login)
		buf = appendBinaryField(buf, rec.Password)
		buf = appendBinaryField(buf, rec.Note)
	}
	return buf, nil
}

func unmarshalVaultBinary(data []byte) (map[string]domain.Record, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid binary payload: too short")
	}

	count := binary.BigEndian.Uint32(data[:4])
	records := make(map[string]domain.Record, count)
	offset := 4

	for i := uint32(0); i < count; i++ {
		service, n, err := readBinaryField(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += n

		login, n, err := readBinaryField(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += n

		password, n, err := readBinaryField(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += n

		note, n, err := readBinaryField(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += n

		key := string(service)
		records[key] = domain.Record{
			Service:  service,
			Login:    login,
			Password: password,
			Note:     note,
		}
	}
	return records, nil
}

func appendBinaryField(buf []byte, data []byte) []byte {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(data)))
	buf = append(buf, lenBuf...)
	buf = append(buf, data...)
	return buf
}

func readBinaryField(data []byte) ([]byte, int, error) {
	if len(data) < 2 {
		return nil, 0, fmt.Errorf("truncated field length")
	}
	length := int(binary.BigEndian.Uint16(data[:2]))
	if len(data) < 2+length {
		return nil, 0, fmt.Errorf("truncated field data")
	}

	field := make([]byte, length)
	copy(field, data[2:2+length])
	return field, 2 + length, nil
}
