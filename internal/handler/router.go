package handler

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"passworder/internal/auth"
	"passworder/internal/config"
	"passworder/internal/model"
	"passworder/internal/repository"
	"passworder/internal/service"
	"passworder/internal/storage"
)

type Router struct {
	*mux.Router
}

func NewRouter(cfg *config.Config, db *sqlx.DB, fileStore *storage.FileStorage, staticFS embed.FS) *Router {
	r := &Router{Router: mux.NewRouter()}

	authRepo := repository.NewAuthRepository(db)
	authService := auth.NewService()

	if storedAuth, err := authRepo.Get(); err == nil {
		authService.Load(storedAuth.PasswordHash, storedAuth.KDFSalt)
	}

	settingService := service.NewSettingService(repository.NewSettingRepository(db))
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	accountRepo := repository.NewAccountRepository(db)
	accountService := service.NewAccountService(accountRepo)
	attRepo := repository.NewAttachmentRepository(db)
	attService := service.NewAttachmentService(attRepo, fileStore)
	remRepo := repository.NewReminderRepository(db)
	remService := service.NewReminderService(remRepo, settingService, service.NewSMTPSender())
	noteAttRepo := repository.NewNoteAttachmentRepository(db)
	personalFileRepo := repository.NewPersonalFileRepository(db)
	personalFileService := service.NewPersonalFileService(personalFileRepo, noteAttRepo, fileStore)
	noteAttService := service.NewNoteAttachmentService(noteAttRepo, fileStore)

	authHandler := NewAuthHandler(authService, authRepo, settingService)
	categoryHandler := NewCategoryHandler(categoryService)
	accountHandler := NewAccountHandler(accountService, remService, authService, categoryService)
	attachmentHandler := NewAttachmentHandler(attService, accountService)
	reminderHandler := NewReminderHandler(remService)
	settingHandler := NewSettingHandler(settingService)
	importExportHandler := NewImportExportHandler(accountService, categoryService, remService, personalFileService, noteAttService, authService, settingService)
	personalFileHandler := NewPersonalFileHandler(personalFileService)
	noteAttachmentHandler := NewNoteAttachmentHandler(noteAttService)

	r.HandleFunc("/api/auth/setup", authHandler.Setup).Methods("POST")
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/api/auth/logout", authHandler.Logout).Methods("POST")
	r.HandleFunc("/api/auth/check", authHandler.Check).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware(authService))

	api.HandleFunc("/categories", categoryHandler.List).Methods("GET")
	api.HandleFunc("/categories", categoryHandler.Create).Methods("POST")
	api.HandleFunc("/categories/{id}", categoryHandler.Get).Methods("GET")
	api.HandleFunc("/categories/{id}", categoryHandler.Update).Methods("PUT")
	api.HandleFunc("/categories/{id}", categoryHandler.Delete).Methods("DELETE")

	api.HandleFunc("/accounts", accountHandler.List).Methods("GET")
	api.HandleFunc("/accounts/search", accountHandler.Search).Methods("GET")
	api.HandleFunc("/accounts", accountHandler.Create).Methods("POST")
	api.HandleFunc("/accounts/{id}", accountHandler.Get).Methods("GET")
	api.HandleFunc("/accounts/{id}", accountHandler.Update).Methods("PUT")
	api.HandleFunc("/accounts/{id}", accountHandler.Delete).Methods("DELETE")
	api.HandleFunc("/accounts/{id}/password", accountHandler.GetPassword).Methods("GET")

	api.HandleFunc("/accounts/{id}/attachments", attachmentHandler.List).Methods("GET")
	api.HandleFunc("/accounts/{id}/attachments", attachmentHandler.Upload).Methods("POST")
	api.HandleFunc("/attachments/{id}", attachmentHandler.Download).Methods("GET")
	api.HandleFunc("/attachments/{id}", attachmentHandler.Delete).Methods("DELETE")

	api.HandleFunc("/reminders", reminderHandler.List).Methods("GET")
	api.HandleFunc("/reminders", reminderHandler.Create).Methods("POST")
	api.HandleFunc("/reminders/{id}", reminderHandler.Delete).Methods("DELETE")
	api.HandleFunc("/reminders/pending", reminderHandler.GetPending).Methods("GET")
	api.HandleFunc("/reminders/send-due", reminderHandler.SendDue).Methods("POST")

	api.HandleFunc("/settings", settingHandler.List).Methods("GET")
	api.HandleFunc("/settings/{key}", settingHandler.Get).Methods("GET")
	api.HandleFunc("/settings/{key}", settingHandler.Set).Methods("PUT")
	api.HandleFunc("/server-config", settingHandler.GetServerConfig).Methods("GET")
	api.HandleFunc("/server-config", settingHandler.SetServerConfig).Methods("PUT")

	api.HandleFunc("/export", importExportHandler.Export).Methods("GET")
	api.HandleFunc("/import", importExportHandler.Import).Methods("POST")

	api.HandleFunc("/files", personalFileHandler.List).Methods("GET")
	api.HandleFunc("/files", personalFileHandler.Create).Methods("POST")
	api.HandleFunc("/files/{id}", personalFileHandler.Get).Methods("GET")
	api.HandleFunc("/files/{id}", personalFileHandler.Update).Methods("PUT")
	api.HandleFunc("/files/{id}/preview", personalFileHandler.Preview).Methods("GET")
	api.HandleFunc("/files/{id}/name", personalFileHandler.UpdateName).Methods("PUT")
	api.HandleFunc("/files/{id}/content", personalFileHandler.UpdateContent).Methods("PUT")
	api.HandleFunc("/files/{id}", personalFileHandler.Delete).Methods("DELETE")

	api.HandleFunc("/notes/trash", personalFileHandler.ListTrash).Methods("GET")
	api.HandleFunc("/notes/trash", personalFileHandler.EmptyTrash).Methods("DELETE")
	api.HandleFunc("/notes", personalFileHandler.List).Methods("GET")
	api.HandleFunc("/notes", personalFileHandler.Create).Methods("POST")
	api.HandleFunc("/notes/{id}", personalFileHandler.Get).Methods("GET")
	api.HandleFunc("/notes/{id}", personalFileHandler.Update).Methods("PUT")
	api.HandleFunc("/notes/{id}/preview", personalFileHandler.Preview).Methods("GET")
	api.HandleFunc("/notes/{id}/name", personalFileHandler.UpdateName).Methods("PUT")
	api.HandleFunc("/notes/{id}/content", personalFileHandler.UpdateContent).Methods("PUT")
	api.HandleFunc("/notes/{id}", personalFileHandler.Delete).Methods("DELETE")
	api.HandleFunc("/notes/{id}/restore", personalFileHandler.Restore).Methods("POST")

	api.HandleFunc("/files/{id}/attachments", noteAttachmentHandler.List).Methods("GET")
	api.HandleFunc("/files/{id}/attachments", noteAttachmentHandler.Create).Methods("POST")
	api.HandleFunc("/note-attachments/{id}", noteAttachmentHandler.Get).Methods("GET")
	api.HandleFunc("/note-attachments/{id}/preview", noteAttachmentHandler.Preview).Methods("GET")
	api.HandleFunc("/note-attachments/{id}", noteAttachmentHandler.Delete).Methods("DELETE")

	api.HandleFunc("/password/generate", PasswordGenerator).Methods("GET")

	staticContent, _ := fs.Sub(staticFS, "static")
	staticHandler := http.FileServer(http.FS(staticContent))
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		staticHandler.ServeHTTP(w, req)
	}))

	return r
}

func authMiddleware(authService *auth.Service) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				JSONError(w, http.StatusUnauthorized, "未登录")
				return
			}
			if !authService.ValidateSession(token) {
				JSONError(w, http.StatusUnauthorized, "会话已过期")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.Response{
		Type:    model.ResponseSuccess,
		Message: "success",
		Data:    data,
	})
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.Response{
		Type:    model.ResponseError,
		Message: message,
	})
}
