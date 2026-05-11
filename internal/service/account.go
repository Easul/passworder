package service

import (
	"time"

	"passworder/internal/crypto"
	"passworder/internal/model"
	"passworder/internal/repository"
)

type AccountService struct {
	repo      *repository.AccountRepository
	cryptoKey []byte
}

func NewAccountService(repo *repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) SetCryptoKey(key []byte) {
	s.cryptoKey = key
}

func (s *AccountService) Create(a *model.Account, plainPassword string) error {
	encrypted, err := crypto.Encrypt(plainPassword, s.cryptoKey)
	if err != nil {
		return err
	}
	a.PasswordEncrypted = []byte(encrypted)
	return s.repo.Create(a)
}

func (s *AccountService) CreateImported(a *model.Account) error {
	return s.repo.Create(a)
}

func (s *AccountService) Update(a *model.Account, plainPassword string) error {
	if plainPassword != "" {
		encrypted, err := crypto.Encrypt(plainPassword, s.cryptoKey)
		if err != nil {
			return err
		}
		a.PasswordEncrypted = []byte(encrypted)
	}
	return s.repo.Update(a)
}

func (s *AccountService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *AccountService) Get(id int64) (*model.Account, error) {
	acc, err := s.repo.Get(id, time.Now().Unix())
	if err != nil || acc == nil {
		return acc, err
	}
	password, err := crypto.Decrypt(string(acc.PasswordEncrypted), s.cryptoKey)
	if err != nil {
		return nil, err
	}
	acc.Password = password
	return acc, nil
}

func (s *AccountService) List() ([]model.Account, error) {
	return s.repo.List(time.Now().Unix())
}

func (s *AccountService) Search(query string) ([]model.Account, error) {
	return s.repo.Search(query, time.Now().Unix())
}

func (s *AccountService) ListByCategory(categoryID int64) ([]model.Account, error) {
	return s.repo.ListByCategory(categoryID, time.Now().Unix())
}
