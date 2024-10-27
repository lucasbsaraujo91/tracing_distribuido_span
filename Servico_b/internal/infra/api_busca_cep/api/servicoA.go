package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"temperatura_por_cep/internal/infra/api_busca_cep/entity"
)

// AddressFetcherParaServicoA define o novo fetcher para consultar endereços via Serviço A.
type DefaultAddressFetcher struct{}

// FetchAddressFromServicoA consulta o Serviço A para obter o endereço com base no CEP.
func (f *DefaultAddressFetcher) FetchAddressFromServicoA(cep string) (entity.ServicoAddress, error) {
	//time.Sleep(2 * time.Second)
	var address entity.ServicoAddress
	resp, err := http.Get(fmt.Sprintf("http://servico_a:8085/address/%s", cep))
	if err != nil {
		return address, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return address, fmt.Errorf("failed to fetch serviço A: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		return address, err
	}
	return address, nil
}
