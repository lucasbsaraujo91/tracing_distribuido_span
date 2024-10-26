package entity

import (
	"errors"
	"regexp"
)

type Address struct {
	ZipCode      string // Para ambos os CEPs
	Street       string // Para Logradouro e Rua
	Neighborhood string // Para Bairro
	City         string // Para Cidade
	State        string // Para Estado
	Complemento  string // Para Complemento (se necessário)
	IBGE         string // Para IBGE (se necessário)
	GIA          string // Para GIA (se necessário)
	DDD          string // Para DDD (se necessário)
	SIAFI        string // Para SIAFI (se necessário)
}

func NewConsultZipCode(zipCode string, street string, neighborhood string, city string, state string) (*Address, error) {
	address := &Address{
		ZipCode:      zipCode,
		Street:       street,
		Neighborhood: neighborhood,
		City:         city,
		State:        state,
	}

	err := address.IsValid()
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (a *Address) IsValid() error {
	if len(a.ZipCode) != 8 {
		return errors.New("invalid zipcode: must be 8 digits")
	}
	if matched, _ := regexp.MatchString("^[0-9]+$", a.ZipCode); !matched {
		return errors.New("invalid zipcode: must contain only digits")
	}
	return nil
}
