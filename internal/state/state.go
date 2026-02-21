package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ResourceState struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	ProviderID string `json:"provider_id"`
}

type State struct {
	Resources map[string]ResourceState `json:"resources"`
}

// LoadState fetches the state.json directly from your S3 bucket
func LoadState(bucketName string, key string) (*State, error) {
	state := &State{Resources: make(map[string]ResourceState)}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for state: %w", err)
	}
	client := s3.NewFromConfig(cfg)

	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Println("No remote state found. Starting fresh!")
		return state, nil
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote state body: %w", err)
	}

	err = json.Unmarshal(data, state)
	return state, err
}

// SaveState uploads the updated state back to your S3 bucket
func SaveState(bucketName string, key string, state *State) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return err
}
