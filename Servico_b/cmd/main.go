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
	apiTemperatura "temperatura_por_cep/internal/infra/api_busca_temperatura/api"
	"temperatura_por_cep/internal/infra/api_busca_temperatura/service"
	"temperatura_por_cep/internal/usecase"
)

func main() {
	// Configuração do exportador OTLP
	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("172.23.0.2:4317"), otlptracegrpc.WithInsecure())
	if err != nil {
		log.Fatalf("Erro ao criar o exportador OTLP: %v", err)
	}

	// Configuração do provedor de traços com o exportador OTLP
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("servico-temperatura-cep"),
		)),
	)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)

	// Configuração das rotas e handlers
	r := chi.NewRouter()
	fetcher := &api.DefaultAddressFetcher{}
	apiKey := "4457481f588941748b8232000240910" // Substitua pela chave da API
	weatherFetcher := apiTemperatura.NewWeatherFetcher(apiKey)
	weatherService := service.NewWeatherService(apiKey, weatherFetcher)

	addressUseCase := usecase.NewAddressUseCase(fetcher)
	weatherUseCase := &usecase.WeatherUseCase{
		AddressUseCase: addressUseCase,
		WeatherService: weatherService,
	}

	// Rota para busca de clima com traços distribuídos
	r.Get("/weather/zipcode/{zipCode}", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("servico-temperatura").Start(r.Context(), "handleWeatherByZipCode")
		defer span.End()

		zipCode := chi.URLParam(r, "zipCode")
		handleWeatherByZipCode(ctx, w, zipCode, weatherUseCase)
	})

	log.Fatal(http.ListenAndServe(":8082", r))
}

func handleWeatherByZipCode(ctx context.Context, w http.ResponseWriter, zipCode string, weatherUseCase *usecase.WeatherUseCase) {
	// Span para medir o tempo de resposta da busca de clima
	ctx, span := otel.Tracer("servico-temperatura").Start(ctx, "BuscarClima")
	defer span.End()

	weather, err := weatherUseCase.GetWeatherByZipCode(ctx, zipCode)
	if err != nil {
		handleError(w, err)
		return
	}
	sendJSONResponse(w, weather)
}

func handleError(w http.ResponseWriter, err error) {
	if err.Error() == "invalid zipcode" {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
	} else if err.Error() == "can not find zipcode" {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
	} else {
		http.Error(w, fmt.Sprintf("Error fetching weather: %v", err), http.StatusInternalServerError)
	}
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
