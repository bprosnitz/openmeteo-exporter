package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var url = `https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,apparent_temperature,is_day,precipitation,rain,showers,snowfall,weather_code,cloud_cover,pressure_msl,surface_pressure,wind_speed_10m,wind_direction_10m,wind_gusts_10m&daily=weather_code,temperature_2m_max,temperature_2m_min,apparent_temperature_max,apparent_temperature_min,sunrise,sunset,daylight_duration,sunshine_duration,uv_index_max,uv_index_clear_sky_max,precipitation_sum,rain_sum,showers_sum,snowfall_sum,precipitation_hours,precipitation_probability_max,wind_speed_10m_max,wind_gusts_10m_max,wind_direction_10m_dominant,shortwave_radiation_sum,et0_fao_evapotranspiration&timezone=America%%2FLos_Angeles&forecast_days=1`

func main() {
	var pollInterval time.Duration
	var latitude float64
	var longitude float64
	flag.DurationVar(&pollInterval, "poll-interval", time.Minute, "poll frequency")
	flag.Float64Var(&latitude, "latitude", 0, "latitude")
	flag.Float64Var(&longitude, "longitude", 0, "longitude")
	flag.Parse()

	if latitude == 0 && longitude == 0 {
		log.Fatalf("Please specify latitude and longitude")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fullUrl := fmt.Sprintf(url, latitude, longitude)
	log.Printf("full url: %s", fullUrl)

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullUrl, strings.NewReader(""))
		if err != nil {
			log.Fatalf("error constructing request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("error in http request: %v", err)
		}
		var response response
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&response); err != nil {
			log.Fatalf("error decoding response body: %v", err)
		}

		metrics.elevation.Set(response.Elevation)

		metrics.current.timeSinceUpdate.Set(time.Since(response.Current.Time).Seconds())
		metrics.current.temperature2M.Set(response.Current.Temperature2M)
		metrics.current.relativeHumidity2M.Set(response.Current.RelativeHumidity2M)
		metrics.current.apparentTemperature.Set(response.Current.ApparentTemperature)
		metrics.current.precipitation.Set(response.Current.Precipitation)
		metrics.current.rain.Set(response.Current.Rain)
		metrics.current.showers.Set(response.Current.Showers)
		metrics.current.snowfall.Set(response.Current.Snowfall)
		metrics.current.weatherCode.Set(response.Current.WeatherCode)
		metrics.current.cloudCover.Set(response.Current.CloudCover)
		metrics.current.pressureMsl.Set(response.Current.PressureMsl)
		metrics.current.surfacePressure.Set(response.Current.SurfacePressure)
		metrics.current.windSpeed10M.Set(response.Current.WindSpeed10M)
		metrics.current.windDirection10M.Set(response.Current.WindDirection10M)
		metrics.current.windGusts10m.Set(response.Current.WindGusts10m)

		metrics.daily.timeSinceUpdate.Set(time.Since(response.Daily.Time[0]).Seconds())
		metrics.daily.weatherCode.Set(response.Daily.WeatherCode[0])
		metrics.daily.temperature2MMax.Set(response.Daily.Temperature2MMax[0])
		metrics.daily.temperature2MMin.Set(response.Daily.Temperature2MMin[0])
		metrics.daily.apparentTemperatureMax.Set(response.Daily.ApparentTemperatureMax[0])
		metrics.daily.apparentTemperatureMin.Set(response.Daily.ApparentTemperatureMin[0])
		//metrics.daily.sunrise.Set(response.Daily.Sunrise[0])
		//metrics.daily.sunset.Set(response.Daily.sunset)
		metrics.daily.daylightDuraton.Set(response.Daily.DaylightDuraton[0])
		metrics.daily.sunshineDuration.Set(response.Daily.SunshineDuration[0])
		metrics.daily.uvIndexMax.Set(response.Daily.UVIndexMax[0])
		metrics.daily.uvIndexClearSkyMax.Set(response.Daily.UVIndexClearSkyMax[0])
		metrics.daily.precipitationSum.Set(response.Daily.PrecipitationSum[0])
		metrics.daily.rainSum.Set(response.Daily.RainSum[0])
		metrics.daily.showersSum.Set(response.Daily.ShowersSum[0])
		metrics.daily.snowfallSum.Set(response.Daily.SnowfallSum[0])
		metrics.daily.precipitationHours.Set(response.Daily.PrecipitationHours[0])
		metrics.daily.precipitationProbabilityMax.Set(response.Daily.PrecipitationProbabilityMax[0])
		metrics.daily.windSpeed10MMax.Set(response.Daily.WindSpeed10MMax[0])
		metrics.daily.windGusts10MMax.Set(response.Daily.WindGusts10MMax[0])
		metrics.daily.windDirection10MDominant.Set(response.Daily.WindDirection10MDominant[0])
		metrics.daily.shortwaveRadiationSum.Set(response.Daily.ShortwaveRadiationSum[0])
		metrics.daily.et0FAOEvapotranspiration.Set(response.Daily.ET0FAOEvapotranspiration[0])

		select {
		case <-ctx.Done():
			log.Printf("%v", ctx.Err())
			return
		case <-time.After(pollInterval):
		}
	}
}

type response struct {
	Elevation float64
	Current   struct {
		Time                time.Time `json:"time"`
		Temperature2M       float64   `json:"temperature_2m"`
		RelativeHumidity2M  float64   `json:"relative_humidity_2m"`
		ApparentTemperature float64   `json:"apparent_temperature"`
		Precipitation       float64   `json:"precipitation"`
		Rain                float64   `json:"rain"`
		Showers             float64   `json:"showers"`
		Snowfall            float64   `json:"snowfall"`
		WeatherCode         float64   `json:"weather_code"`
		CloudCover          float64   `json:"cloud_cover"`
		PressureMsl         float64   `json:"pressure_msl"`
		SurfacePressure     float64   `json:"surface_pressure"`
		WindSpeed10M        float64   `json:"wind_speed_10m"`
		WindDirection10M    float64   `json:"wind_direction_10m"`
		WindGusts10m        float64   `json:"wind_gusts_10m"`
	} `json:"current"`
	Daily struct {
		Time                        []time.Time `json:"time"`
		WeatherCode                 []float64   `json:"weather_code"`
		Temperature2MMax            []float64   `json:"temperature_2m_max"`
		Temperature2MMin            []float64   `json:"temperature_2m_min"`
		ApparentTemperatureMax      []float64   `json:"apparent_temperature_max"`
		ApparentTemperatureMin      []float64   `json:"apparent_temperature_min"`
		Sunrise                     []time.Time `json:"sunrise"`
		Sunset                      []time.Time `json:"sunset"`
		DaylightDuraton             []float64   `json:"daylight_duration"`
		SunshineDuration            []float64   `json:"sunshine_duration"`
		UVIndexMax                  []float64   `json:"uv_index_max"`
		UVIndexClearSkyMax          []float64   `json:"uv_index_clear_sky_max"`
		PrecipitationSum            []float64   `json:"precipitation_sum"`
		RainSum                     []float64   `json:"rain_sum"`
		ShowersSum                  []float64   `json:"showers_sum"`
		SnowfallSum                 []float64   `json:"snowfall_sum"`
		PrecipitationHours          []float64   `json:"precipitation_hours"`
		PrecipitationProbabilityMax []float64   `json:"precipitation_probability_max"`
		WindSpeed10MMax             []float64   `json:"wind_speed_10m_max"`
		WindGusts10MMax             []float64   `json:"wind_gusts_10m_max"`
		WindDirection10MDominant    []float64   `json:"wind_direction_10m_dominant"`
		ShortwaveRadiationSum       []float64   `json:"shortwave_radiation_sum"`
		ET0FAOEvapotranspiration    []float64   `json:"et0_fao_evapotranspiration"`
	} `json:"daily"`
}

type currentMetrics struct {
	timeSinceUpdate     prometheus.Gauge
	temperature2M       prometheus.Gauge
	relativeHumidity2M  prometheus.Gauge
	apparentTemperature prometheus.Gauge
	precipitation       prometheus.Gauge
	rain                prometheus.Gauge
	showers             prometheus.Gauge
	snowfall            prometheus.Gauge
	weatherCode         prometheus.Gauge
	cloudCover          prometheus.Gauge
	pressureMsl         prometheus.Gauge
	surfacePressure     prometheus.Gauge
	windSpeed10M        prometheus.Gauge
	windDirection10M    prometheus.Gauge
	windGusts10m        prometheus.Gauge
}

type dailyMetrics struct {
	timeSinceUpdate             prometheus.Gauge
	weatherCode                 prometheus.Gauge
	temperature2MMax            prometheus.Gauge
	temperature2MMin            prometheus.Gauge
	apparentTemperatureMax      prometheus.Gauge
	apparentTemperatureMin      prometheus.Gauge
	sunrise                     prometheus.Gauge
	sunset                      prometheus.Gauge
	daylightDuraton             prometheus.Gauge
	sunshineDuration            prometheus.Gauge
	uvIndexMax                  prometheus.Gauge
	uvIndexClearSkyMax          prometheus.Gauge
	precipitationSum            prometheus.Gauge
	rainSum                     prometheus.Gauge
	showersSum                  prometheus.Gauge
	snowfallSum                 prometheus.Gauge
	precipitationHours          prometheus.Gauge
	precipitationProbabilityMax prometheus.Gauge
	windSpeed10MMax             prometheus.Gauge
	windGusts10MMax             prometheus.Gauge
	windDirection10MDominant    prometheus.Gauge
	shortwaveRadiationSum       prometheus.Gauge
	et0FAOEvapotranspiration    prometheus.Gauge
}

var metrics = struct {
	elevation prometheus.Gauge
	current   currentMetrics
	daily     dailyMetrics
}{
	elevation: prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "openmeteo",
		Name:      "elevation",
	}),
	current: currentMetrics{
		timeSinceUpdate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_time_since_last_update",
		}),
		temperature2M: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_temperature_2m",
		}),
		relativeHumidity2M: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_relative_humidity_2m",
		}),
		apparentTemperature: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_apparent_temperature",
		}),
		precipitation: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_precipitation",
		}),
		rain: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_rain",
		}),
		showers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_showers",
		}),
		snowfall: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_snowfall",
		}),
		weatherCode: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_weather_code",
		}),
		cloudCover: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_cloud_cover",
		}),
		pressureMsl: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_pressure_msl",
		}),
		surfacePressure: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_surface_pressure",
		}),
		windSpeed10M: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_wind_speed_10m",
		}),
		windDirection10M: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_wind_direction_10m",
		}),
		windGusts10m: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "current_wind_gusts_10m",
		}),
	},
	daily: dailyMetrics{
		timeSinceUpdate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_time_since_last_update",
		}),
		weatherCode: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_weather_code",
		}),
		temperature2MMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_temperature_2m_max",
		}),
		temperature2MMin: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_temperature_2m_min",
		}),
		apparentTemperatureMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_apparent_temperature_max",
		}),
		apparentTemperatureMin: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_apparent_temperature_min",
		}),
		sunrise: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_sunrise",
		}),
		sunset: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_sunset",
		}),
		daylightDuraton: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_daylight_duration",
		}),
		sunshineDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_sunshine_duration",
		}),
		uvIndexMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_uv_index_max",
		}),
		uvIndexClearSkyMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_uv_index_clear_sky_max",
		}),
		precipitationSum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_precipitation_sum",
		}),
		rainSum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_rain_sum",
		}),
		showersSum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_showers_sum",
		}),
		snowfallSum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_snowfall_sum",
		}),
		precipitationHours: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_precipitation_hours",
		}),
		precipitationProbabilityMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_precipitation_probability_max",
		}),
		windSpeed10MMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_wind_speed_10m_max",
		}),
		windGusts10MMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_wind_gusts_10m_max",
		}),
		windDirection10MDominant: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_wind_direction_10m_dominant",
		}),
		shortwaveRadiationSum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_shortwave_radiation_sum",
		}),
		et0FAOEvapotranspiration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "openmeteo",
			Name:      "daily_et0_fao_evapotranspiration",
		}),
	},
}

func init() {
	prometheus.Register(metrics.elevation)
	prometheus.Register(metrics.current.timeSinceUpdate)
	prometheus.Register(metrics.current.temperature2M)
	prometheus.Register(metrics.current.relativeHumidity2M)
	prometheus.Register(metrics.current.apparentTemperature)
	prometheus.Register(metrics.current.precipitation)
	prometheus.Register(metrics.current.rain)
	prometheus.Register(metrics.current.showers)
	prometheus.Register(metrics.current.snowfall)
	prometheus.Register(metrics.current.weatherCode)
	prometheus.Register(metrics.current.cloudCover)
	prometheus.Register(metrics.current.pressureMsl)
	prometheus.Register(metrics.current.surfacePressure)
	prometheus.Register(metrics.current.windSpeed10M)
	prometheus.Register(metrics.current.windDirection10M)
	prometheus.Register(metrics.current.windGusts10m)
	prometheus.Register(metrics.daily.timeSinceUpdate)
	prometheus.Register(metrics.daily.weatherCode)
	prometheus.Register(metrics.daily.temperature2MMax)
	prometheus.Register(metrics.daily.temperature2MMin)
	prometheus.Register(metrics.daily.apparentTemperatureMax)
	prometheus.Register(metrics.daily.apparentTemperatureMin)
	prometheus.Register(metrics.daily.sunrise)
	prometheus.Register(metrics.daily.sunset)
	prometheus.Register(metrics.daily.daylightDuraton)
	prometheus.Register(metrics.daily.sunshineDuration)
	prometheus.Register(metrics.daily.uvIndexMax)
	prometheus.Register(metrics.daily.uvIndexClearSkyMax)
	prometheus.Register(metrics.daily.precipitationSum)
	prometheus.Register(metrics.daily.rainSum)
	prometheus.Register(metrics.daily.showersSum)
	prometheus.Register(metrics.daily.snowfallSum)
	prometheus.Register(metrics.daily.precipitationHours)
	prometheus.Register(metrics.daily.precipitationProbabilityMax)
	prometheus.Register(metrics.daily.windSpeed10MMax)
	prometheus.Register(metrics.daily.windGusts10MMax)
	prometheus.Register(metrics.daily.windDirection10MDominant)
	prometheus.Register(metrics.daily.shortwaveRadiationSum)
	prometheus.Register(metrics.daily.et0FAOEvapotranspiration)
}
