package api

import "temperatura_por_cep/internal/infra/api_busca_cep/entity"

type AddressFetcher interface {
	FetchAddressFromServicoA(cep string) (entity.ServicoAddress, error)
}
