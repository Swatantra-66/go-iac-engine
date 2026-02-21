package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Swatantra-66/go-iac-tool/internal/parser"
	"github.com/Swatantra-66/go-iac-tool/internal/provider"
	"github.com/Swatantra-66/go-iac-tool/internal/state"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const stateBucket = "swatantra-iac-remote-state-999"
const stateKey = "state.json"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables")
	}

	r := gin.Default()

	r.Use(APIKeyMiddleware())

	r.POST("/deploy", handleDeploy)
	r.DELETE("/destroy", handleDestroy)

	fmt.Println("IaC Engine API starting on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleDeploy(c *gin.Context) {
	currentState, err := state.LoadState(stateBucket, stateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load remote state"})
		return
	}

	config, err := parser.ParseConfig("main.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse main.yaml"})
		return
	}

	deployed := []string{}
	updated := []string{}

	for _, res := range config.Resources {
		if existing, exists := currentState.Resources[res.Name]; exists {
			if res.Type == "aws_ec2_instance" {
				fmt.Printf("Drift detected on %s. Triggering update...\n", res.Name)
				err := provider.UpdateEC2Instance(res, existing.ProviderID)
				if err != nil {
					log.Printf("Error updating %s: %v\n", res.Name, err)
					continue
				}
				updated = append(updated, res.Name)
			}
			continue
		}

		providerID, err := provider.DeployResource(res)
		if err != nil {
			log.Printf("Error deploying %s: %v\n", res.Name, err)
			continue
		}

		currentState.Resources[res.Name] = state.ResourceState{
			Type:       res.Type,
			Name:       res.Name,
			ProviderID: providerID,
		}
		deployed = append(deployed, res.Name)
	}

	state.SaveState(stateBucket, stateKey, currentState)

	c.JSON(http.StatusOK, gin.H{
		"status":   "Success",
		"deployed": deployed,
		"updated":  updated,
	})
}

func handleDestroy(c *gin.Context) {
	currentState, err := state.LoadState(stateBucket, stateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load remote state"})
		return
	}

	destroyed := []string{}
	for name, res := range currentState.Resources {
		err := provider.DestroyResource(res.Type, res.ProviderID, "us-east-1")
		if err != nil {
			log.Printf("Error destroying %s: %v\n", name, err)
			continue
		}
		delete(currentState.Resources, name)
		destroyed = append(destroyed, name)
	}

	state.SaveState(stateBucket, stateKey, currentState)

	c.JSON(http.StatusOK, gin.H{
		"status":    "Success",
		"destroyed": destroyed,
	})
}

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secretKey := os.Getenv("IAC_API_KEY")

		if secretKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Server configuration error: API Key not set.",
			})
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey != secretKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized. Please provide a valid X-API-Key header.",
			})
			return
		}
		c.Next()
	}
}
