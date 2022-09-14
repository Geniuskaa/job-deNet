package file

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"os"
	"strconv"
)

const constraint = int64(1 << 20) // 1.java мб
const storagePATH = "filesStorage"

type Service struct {
	fileToUUID map[string]string // uuid : название файла
	// Храним офсет для дробленного файла
	offsets map[string]int64 // uuid : отступ
}

func NewService() *Service {
	return &Service{fileToUUID: make(map[string]string), offsets: make(map[string]int64)}
}

func (s *Service) ExecuteFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	id := uuid.NewString()
	s.fileToUUID[id] = header.Filename

	path := fmt.Sprintf("%s/%s", storagePATH, id)
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("ExecuteFile failed during dir creation: %w", err)
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ExecuteFile failed during file reading: %w", err)
	}

	if header.Size < constraint {
		dif := constraint - header.Size
		fileData = addData(fileData, dif)
		err = saveFile(path, fileData, 0)
		if err != nil {
			return "", fmt.Errorf("ExecuteFile failed during saving file part: %w", err)
		}
		s.offsets[id] = dif
	} else {
		wholeParts := header.Size / constraint
		restPart := header.Size - wholeParts*constraint

		for i := int64(0); i < wholeParts; i++ {
			err = saveFile(path, fileData[i*constraint:constraint*(i+1)], int(i))
			if err != nil {
				return "", fmt.Errorf("ExecuteFile failed during saving file parts: %w", err)
			}
		}

		filledData := addData(fileData[constraint*(wholeParts):], constraint-restPart)
		err = saveFile(path, filledData, int(wholeParts))
		if err != nil {
			return "", fmt.Errorf("ExecuteFile failed during saving file parts: %w", err)
		}
		s.offsets[id] = restPart
	}

	return id, nil
}

func addData(arr []byte, dif int64) []byte {
	newArr := make([]byte, dif)
	newArr = append(newArr, arr...)
	return newArr
}

func saveFile(path string, file []byte, num int) error {
	strNum := strconv.Itoa(num)
	err := os.WriteFile(fmt.Sprintf("%s/%s", path, strNum), file, os.ModePerm)
	return err
}

func (s *Service) ConcatenateFile(fileId string) ([]byte, string, error) {
	filename, exists := s.fileToUUID[fileId]
	if !exists {
		return nil, "", errors.New("ConcatenateFile failed: such file not exists")
	}

	offset, exists := s.offsets[fileId]

	filesPath := fmt.Sprintf("%s/%s", storagePATH, fileId)

	arr, err := os.ReadDir(filesPath)
	if err != nil {
		return nil, "", fmt.Errorf("ConcatenateFile failed: %w", err)
	}

	countOfFiles := len(arr)
	buf := bytes.NewBuffer(make([]byte, 0, countOfFiles*int(constraint)))

	for i := 0; i < countOfFiles; i++ {
		name := strconv.Itoa(i)
		data, err := os.ReadFile(fmt.Sprintf("%s/%s", filesPath, name))
		if err != nil {
			return nil, "", fmt.Errorf("ConcatenateFile failed during opening file: %w", err)
		}
		if i == countOfFiles-1 {
			buf.Write(data[offset:])
		} else {
			buf.Write(data)
		}
	}

	return buf.Bytes(), filename, nil
}
