package file

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
)

type handler struct {
	ctx     context.Context
	log     *zap.SugaredLogger
	service *Service
}

func NewHandler(ctx context.Context, log *zap.SugaredLogger, service *Service) *handler {
	return &handler{ctx: ctx, log: log, service: service}
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/upload", h.FileSplit)
	r.Get("/download/{name}", h.FileUnion)

	return r
}

func (h *handler) FileSplit(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseMultipartForm(90 << 20) // Выставление максимального размера файла на прием. Сейчас 90 Мб
	if err != nil {
		h.log.Error(fmt.Errorf("FileSplit failed: %w", err))
		http.Error(writer, "Error setting the file size", http.StatusInternalServerError)
		return
	}

	file, hed, err := request.FormFile("file")
	if err != nil {
		h.log.Error(fmt.Errorf("FileSplit failed during file getting: %w", err))
		http.Error(writer, "Error retrieving the File, use 'file' key", http.StatusUnprocessableEntity)
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			h.log.Error(fmt.Errorf("FileSplit failed: %w", err))
		}
	}(file)

	id, err := h.service.ExecuteFile(file, hed)
	if err != nil {
		h.log.Error(fmt.Errorf("FileSplit failed: %w", err))
		http.Error(writer, "Error parsing file", http.StatusInternalServerError)
		return
	}

	success := &SuccessUpload{
		Text:   "Successfully saved your file!",
		FileId: id,
	}

	data, err := json.Marshal(success)
	if err != nil {
		h.log.Error(fmt.Errorf("FileSplit failed during marshaling: %w", err))
		http.Error(writer, "Error parsing file", http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Type", "application/json")
	writer.Write(data)

	return
}

func (h *handler) FileUnion(writer http.ResponseWriter, request *http.Request) {
	fileName := chi.URLParam(request, "name")
	file, fileName, err := h.service.ConcatenateFile(fileName)
	if err != nil {
		h.log.Error(fmt.Errorf("FileSplit failed: %w", err))
		http.Error(writer, "Error parsing file", http.StatusInternalServerError)
		return
	}

	writer.Header().Add("Content-Disposition", fmt.Sprintf("form-data; name=\"file\";filename=\"%s\"", fileName))
	writer.Header().Add("Content-Type", "multipart/form-data")
	writer.Write(file)

	return
}
