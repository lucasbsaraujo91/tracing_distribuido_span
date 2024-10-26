package service

import "temperatura_por_cep/internal/infra/api_busca_cep/entity"

type AddressFetcher interface {
	FetchFromServicoA(zipCode string) (entity.ServicoAddress, error)
}

type AddressData struct {
	CEP          string
	Street       string
	Neighborhood string
	City         string
	State        string
}
