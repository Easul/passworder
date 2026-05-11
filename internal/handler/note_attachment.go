package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"passworder/internal/model"
	"passworder/internal/service"
)

type NoteAttachmentHandler struct {
	service *service.NoteAttachmentService
}

func NewNoteAttachmentHandler(service *service.NoteAttachmentService) *NoteAttachmentHandler {
	return &NoteAttachmentHandler{service: service}
}

func (h *NoteAttachmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		JSONError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		JSONError(w, http.StatusBadRequest, "未选择文件")
		return
	}

	var created []*model.NoteAttachment
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		a, err := h.service.Create(fileID, header, file)
		if err != nil {
			continue
		}
		created = append(created, a)
	}

	JSONResponse(w, http.StatusCreated, created)
}

func (h *NoteAttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	attachments, err := h.service.ListByFile(fileID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取附件列表失败")
		return
	}
	JSONResponse(w, http.StatusOK, attachments)
}

func (h *NoteAttachmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的附件 ID")
		return
	}

	a, rc, err := h.service.Open(id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "附件不存在")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", a.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(a.OriginalName)))
	w.Header().Set("Content-Length", strconv.FormatInt(a.SizeBytes, 10))
	io.Copy(w, rc)
}

func (h *NoteAttachmentHandler) Preview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的附件 ID")
		return
	}

	a, rc, err := h.service.Open(id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "附件不存在")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", a.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", url.PathEscape(a.OriginalName)))
	io.Copy(w, rc)
}

func (h *NoteAttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的附件 ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "删除附件失败")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}
