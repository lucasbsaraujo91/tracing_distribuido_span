// weather_usecase.go
package usecase

import (
	"context"
	"temperatura_por_cep/internal/entity"
	"temperatura_por_cep/internal/infra/api_busca_temperatura/service"
	"temperatura_por_cep/internal/utils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// DTOs de entrada e saída

type GetWeatherOutputDTO struct {
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

type WeatherUseCase struct {
	AddressUseCase *AddressUseCase
	WeatherService *service.WeatherService
}

func (u *WeatherUseCase) GetWeatherByZipCode(ctx context.Context, zipCode string) (*GetWeatherOutputDTO, error) {
	// Inicia um span para a busca de endereço pelo CEP
	ctx, span := otel.Tracer("WeatherUseCase").Start(ctx, "GetWeatherByZipCode")
	defer span.End()

	// Busca o endereço pelo CEP
	address, err := u.getAddressByZipCode(ctx, zipCode)
	if err != nil {
		return nil, err
	}

	// Busca o clima pela cidade do endereço
	weather, err := u.getWeatherByCity(ctx, address.City, address.State)
	if err != nil {
		return nil, err
	}

	return weather, nil
}

func (u *WeatherUseCase) getAddressByZipCode(ctx context.Context, zipCode string) (*entity.Address, error) {
	// Inicia um span para rastrear a função getAddressByZipCode
	ctx, span := otel.Tracer("WeatherUseCase").Start(ctx, "Serviço A")
	defer span.End()

	// Cria o DTO de entrada
	input := GetAddressInputDTO{ZipCode: zipCode}

	// Chama o método GetAddressByZipCode
	addressOutput, err := u.AddressUseCase.GetAddressByZipCode(input)
	if err != nil {
		return nil, err
	}

	// Adiciona informações do retorno ao span
	span.SetAttributes(
		attribute.String("address.street", addressOutput.Street),
		attribute.String("address.city", addressOutput.City),
		attribute.String("address.state", addressOutput.State),
		attribute.String("address.zipcode", addressOutput.ZipCode),
	)

	address := &entity.Address{
		Street:  addressOutput.Street,
		City:    addressOutput.City,
		State:   addressOutput.State,
		ZipCode: addressOutput.ZipCode,
	}
	return address, nil
}

func (u *WeatherUseCase) getWeatherByCity(ctx context.Context, city string, state string) (*GetWeatherOutputDTO, error) {
	// Inicia um span para rastrear a função getWeatherByCity
	ctx, span := otel.Tracer("WeatherUseCase").Start(ctx, "Serviço B")
	defer span.End()

	newCity := utils.SanitizeCity(utils.RemoveAccents(city))

	fullWeather, err := u.WeatherService.FetchWeatherByCity(newCity, state)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("weather.city", newCity),
		attribute.String("weather.state", state),
		attribute.Float64("weather.temp_c", fullWeather.Current.TempC),
		attribute.Float64("weather.temp_f", utils.ConvertCelsiusToFahrenheit(fullWeather.Current.TempC)),
		attribute.Float64("weather.temp_k", utils.ConvertCelsiusToKelvin(fullWeather.Current.TempC)),
	)

	// Converte a resposta para o formato desejado
	weather := &GetWeatherOutputDTO{
		TempC: fullWeather.Current.TempC,
		TempF: utils.ConvertCelsiusToFahrenheit(fullWeather.Current.TempC),
		TempK: utils.ConvertCelsiusToKelvin(fullWeather.Current.TempC),
	}

	return weather, nil
}
