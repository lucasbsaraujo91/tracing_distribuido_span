package entity

// BrasilAPIAddress representa o formato de endereço retornado pelo Serviço A.
type ServicoAddress struct {
	Zipcode      string `json:"zipcode"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}
