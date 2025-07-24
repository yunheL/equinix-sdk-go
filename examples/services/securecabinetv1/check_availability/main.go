package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/equinix/equinix-sdk-go/services/securecabinetv1"

	"github.com/equinix/equinix-sdk-go/extensions/equinixoauth2"
)

func main() {
	// Parse command line flags
	var accountNumber string
	flag.StringVar(&accountNumber, "account", "", "Billing account number")
	flag.Parse()

	if accountNumber == "" {
		log.Fatal("Please provide an account number using -account flag")
	}

	// Initialize OAuth2 configuration using environment variables
	ctx := context.Background()
	clientId := os.Getenv("EQUINIX_API_CLIENTID")
	clientSecret := os.Getenv("EQUINIX_API_CLIENTSECRET")

	if clientId == "" || clientSecret == "" {
		log.Fatal("Please set EQUINIX_API_CLIENTID and EQUINIX_API_CLIENTSECRET environment variables")
	}

	baseURL := "https://api.equinix.com"
	authConfig := equinixoauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		BaseURL:      baseURL,
	}
	authTransport := authConfig.New()

	// Configure the Secure Cabinet client
	configuration := securecabinetv1.NewConfiguration()
	configuration.HTTPClient = &http.Client{
		Transport: authTransport,
	}
	configuration.AddDefaultHeader("X-SOURCE", "API")
	configuration.Servers = securecabinetv1.ServerConfigurations{
		{
			URL:         "https://api.equinix.com",
			Description: "Equinix API",
		},
	}
	client := securecabinetv1.NewAPIClient(configuration)

	// Call the availability endpoint
	fmt.Printf("Checking Secure Cabinet availability for account: %s\n", accountNumber)

	availability, resp, err := client.AvailabilityApi.GetProductsAvailability(ctx, accountNumber).Execute()
	if err != nil {
		log.Printf("Error checking availability: %v", err)
		if resp != nil {
			log.Printf("Response status: %s", resp.Status)
			log.Printf("Response body: %s", resp.Body)
		}
		os.Exit(1)
	}

	// Print the results
	fmt.Printf("\nFound %d IBX locations with Secure Cabinet availability:\n\n", len(availability))

	for _, product := range availability {
		fmt.Printf("IBX: %s\n", product.GetIbx())
		fmt.Printf("  Max cabinets per order: %d\n", product.GetMaximumNumberOfCabinetsToOrder())
		fmt.Printf("  Min power draw per cabinet: %.2f kW\n", product.GetMinimumDrawCapacityPerCabinet())
		fmt.Printf("  Max power draw per cabinet: %.2f kW\n", product.GetMaximumDrawCapacityPerCabinet())

		dimensions := product.GetCabinetDimensions()
		fmt.Printf("  Cabinet dimensions:\n")
		fmt.Printf("    Width: %d %s\n", dimensions.Width.GetValue(), dimensions.Width.GetUnit())
		fmt.Printf("    Depth: %d %s\n", dimensions.Depth.GetValue(), dimensions.Depth.GetUnit())
		fmt.Printf("    Height: %d %s\n", dimensions.Height.GetValue(), dimensions.Height.GetUnit())

		if pdu, ok := product.GetPduConfigurationOk(); ok && pdu != nil {
			fmt.Printf("  PDU available: Yes\n")
		} else {
			fmt.Printf("  PDU available: No\n")
		}

		if fabric, ok := product.GetFabricPortSpeedOk(); ok && fabric != nil {
			fmt.Printf("  Fabric port available: Yes\n")
		} else {
			fmt.Printf("  Fabric port available: No\n")
		}

		fmt.Println()
	}
}
