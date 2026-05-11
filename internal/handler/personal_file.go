package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"passworder/internal/service"
)

type PersonalFileHandler struct {
	service *service.PersonalFileService
}

func NewPersonalFileHandler(service *service.PersonalFileService) *PersonalFileHandler {
	return &PersonalFileHandler{service: service}
}

func (h *PersonalFileHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		JSONError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	remarks := strings.TrimSpace(r.FormValue("remarks"))
	body := r.FormValue("body")
	bodyFormat := strings.TrimSpace(r.FormValue("bodyFormat"))
	if title == "" {
		JSONError(w, http.StatusBadRequest, "标题不能为空")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		JSONError(w, http.StatusBadRequest, "文件上传失败")
		return
	}
	if file != nil {
		defer file.Close()
	}

	f, err := h.service.Create(title, remarks, body, bodyFormat, header, file)
	if err != nil {
		if strings.Contains(err.Error(), "正文格式") {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, "保存文件失败")
		return
	}
	JSONResponse(w, http.StatusCreated, f)
}

func (h *PersonalFileHandler) List(w http.ResponseWriter, r *http.Request) {
	fileType := r.URL.Query().Get("type")
	var files interface{}
	var err error

	if fileType != "" {
		files, err = h.service.ListByType(fileType)
	} else {
		files, err = h.service.List()
	}

	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取文件列表失败")
		return
	}
	JSONResponse(w, http.StatusOK, files)
}

func (h *PersonalFileHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	f, rc, err := h.service.Open(id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "文件不存在")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", f.OriginalName))
	w.Header().Set("Content-Length", strconv.FormatInt(f.SizeBytes, 10))
	io.Copy(w, rc)
}

func (h *PersonalFileHandler) Preview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	f, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取文件失败")
		return
	}
	if f == nil {
		JSONError(w, http.StatusNotFound, "文件不存在")
		return
	}

	if f.StoredName == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(f.Body))
		return
	}

	if f.FileType == "markdown" {
		content, err := h.service.ReadContent(id)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "读取文件失败")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(content))
		return
	}

	_, rc, err := h.service.Open(id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "文件不存在")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, rc)
}

func (h *PersonalFileHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	var req struct {
		Title      string `json:"title"`
		Remarks    string `json:"remarks"`
		Body       string `json:"body"`
		BodyFormat string `json:"bodyFormat"`
	}
	if err := JSONRequest(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "请求参数错误")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		JSONError(w, http.StatusBadRequest, "标题不能为空")
		return
	}

	if err := h.service.Update(id, strings.TrimSpace(req.Title), strings.TrimSpace(req.Remarks), req.Body, req.BodyFormat); err != nil {
		if strings.Contains(err.Error(), "正文格式") {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, "更新文件失败")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *PersonalFileHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := JSONRequest(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	f, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取文件失败")
		return
	}
	if f == nil {
		JSONError(w, http.StatusNotFound, "文件不存在")
		return
	}

	if err := h.service.Update(id, strings.TrimSpace(req.Name), f.Remarks, f.Body, f.BodyFormat); err != nil {
		if strings.Contains(err.Error(), "正文格式") {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, "更新文件失败")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *PersonalFileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	if err := h.service.DeleteNote(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "删除文件失败")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *PersonalFileHandler) ListTrash(w http.ResponseWriter, r *http.Request) {
	files, err := h.service.ListDeletedNotes()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取回收站失败")
		return
	}
	JSONResponse(w, http.StatusOK, files)
}

func (h *PersonalFileHandler) Restore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	if err := h.service.RestoreNote(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "恢复文件失败")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *PersonalFileHandler) EmptyTrash(w http.ResponseWriter, r *http.Request) {
	if err := h.service.EmptyTrash(); err != nil {
		JSONError(w, http.StatusInternalServerError, "清空回收站失败")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *PersonalFileHandler) UpdateContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := JSONRequest(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	f, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取文件失败")
		return
	}
	if f == nil {
		JSONError(w, http.StatusNotFound, "文件不存在")
		return
	}
	if f.FileType != "markdown" || f.StoredName == "" {
		JSONError(w, http.StatusBadRequest, "非 Markdown 文件，无法编辑内容")
		return
	}

	if err := h.service.UpdateContent(id, req.Content); err != nil {
		JSONError(w, http.StatusInternalServerError, "更新文件内容失败")
		return
	}

	updatedFile, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "获取更新后的文件失败")
		return
	}
	JSONResponse(w, http.StatusOK, updatedFile)
}

func JSONRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
