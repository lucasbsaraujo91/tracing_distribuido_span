package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"temperatura_por_cep/internal/infra/api_busca_cep/api"
	"temperatura_por_cep/internal/usecase"
)

// Inicialização do OpenTelemetry
func initOpenTelemetry() (*trace.TracerProvider, error) {
	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("172.23.0.2:4317"), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("erro ao criar o exportador OTLP: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("servico-cep"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	// Configuração do OpenTelemetry
	tp, err := initOpenTelemetry()
	if err != nil {
		log.Fatalf("Erro ao inicializar OpenTelemetry: %v", err)
	}
	defer func() {
		// Encerra o provedor de traços
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Erro ao encerrar o provedor de traços: %v", err)
		}
	}()

	// Configuração do servidor e rotas
	r := chi.NewRouter()
	fetcher := &api.DefaultAddressFetcher{}
	r.Get("/address/{cep}", AddressHandler(fetcher))

	// Inicia o servidor
	log.Println("Servidor rodando na porta 8081")
	http.ListenAndServe(":8081", r)
}

// Handler para tratar as requisições de endereço
func AddressHandler(fetcher api.AddressFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("servico-cep").Start(r.Context(), "AddressHandler")
		defer span.End()

		cep := chi.URLParam(r, "cep")
		handleAddressRequest(w, ctx, cep, fetcher)
	}
}

func handleAddressRequest(w http.ResponseWriter, ctx context.Context, cep string, fetcher api.AddressFetcher) {
	_, span := otel.Tracer("servico-cep").Start(ctx, "handleAddressRequest")
	defer span.End()

	addressUseCase := usecase.NewAddressUseCase(fetcher)
	address, err := addressUseCase.GetAddressByZipCode(usecase.GetAddressInputDTO{ZipCode: cep})
	if err != nil {
		switch err.Error() {
		case "invalid zipcode":
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		case "can not find zipcode":
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default:
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	sendJSONResponse(w, address)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshaling data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
