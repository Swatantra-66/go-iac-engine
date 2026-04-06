package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/Swatantra-66/go-iac-tool/internal/parser"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsTypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
)

func DeployResource(res parser.Resource) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(res.Region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	switch res.Type {
	case "aws_s3_bucket":
		return createS3Bucket(cfg, res)
	case "aws_ec2_instance":
		return createEC2Instance(cfg, res.Name, res.AMI, res.InstanceType)
	default:
		return "", fmt.Errorf("unsupported resource type: %s", res.Type)
	}
}

func createS3Bucket(cfg aws.Config, res parser.Resource) (string, error) {
	client := s3.NewFromConfig(cfg)
	bucketName := res.Name

	fmt.Printf("Provisioning S3 bucket '%s'...\n", bucketName)

	// 1. Create the base bucket
	_, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create bucket: %w", err)
	}
	fmt.Printf("Successfully created base S3 bucket: %s\n", bucketName)

	// 2. Apply Security Configurations if specified
	if res.Encryption != "" {
		fmt.Println("Applying KMS Encryption...")
		keyId, err := CreateKMSKey(context.TODO(), cfg, bucketName)
		if err != nil {
			return bucketName, fmt.Errorf("bucket created, but KMS key failed: %w", err)
		}

		_, err = client.PutBucketEncryption(context.TODO(), &s3.PutBucketEncryptionInput{
			Bucket: aws.String(bucketName),
			ServerSideEncryptionConfiguration: &s3Types.ServerSideEncryptionConfiguration{
				Rules: []s3Types.ServerSideEncryptionRule{
					{
						ApplyServerSideEncryptionByDefault: &s3Types.ServerSideEncryptionByDefault{
							SSEAlgorithm:   s3Types.ServerSideEncryptionAwsKms,
							KMSMasterKeyID: aws.String(keyId),
						},
						BucketKeyEnabled: aws.Bool(true),
					},
				},
			},
		})
		if err != nil {
			return bucketName, fmt.Errorf("failed to apply encryption: %w", err)
		}
	}

	if res.BlockPublicAccess {
		fmt.Println("Enabling Block Public Access...")
		_, err = client.PutPublicAccessBlock(context.TODO(), &s3.PutPublicAccessBlockInput{
			Bucket: aws.String(bucketName),
			PublicAccessBlockConfiguration: &s3Types.PublicAccessBlockConfiguration{
				BlockPublicAcls:       aws.Bool(true),
				IgnorePublicAcls:      aws.Bool(true),
				BlockPublicPolicy:     aws.Bool(true),
				RestrictPublicBuckets: aws.Bool(true),
			},
		})
		if err != nil {
			return bucketName, fmt.Errorf("failed to block public access: %w", err)
		}
	}

	if res.BucketPolicy != "" {
		fmt.Println("Attaching IAM Bucket Policy...")
		_, err = client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
			Bucket: aws.String(bucketName),
			Policy: aws.String(res.BucketPolicy),
		})
		if err != nil {
			return bucketName, fmt.Errorf("failed to attach bucket policy: %w", err)
		}
	}

	fmt.Printf("Successfully secured S3 bucket: %s\n", bucketName)
	return bucketName, nil
}

// DestroyResource routes the state information to the correct AWS delete function
func DestroyResource(resourceType string, providerID string, region string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	switch resourceType {
	case "aws_s3_bucket":
		return deleteS3Bucket(cfg, providerID)
	case "aws_ec2_instance":
		return deleteEC2Instance(cfg, providerID)
	default:
		return fmt.Errorf("unsupported resource type for deletion: %s", resourceType)
	}
}

// deleteS3Bucket handles the specific API call to delete an S3 bucket
func deleteS3Bucket(cfg aws.Config, bucketName string) error {
	client := s3.NewFromConfig(cfg)

	fmt.Printf("Destroying S3 bucket '%s'...\n", bucketName)

	_, err := client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	fmt.Printf("Successfully destroyed S3 bucket: %s\n", bucketName)
	return nil
}

func createEC2Instance(cfg aws.Config, name, ami, instanceType string) (string, error) {
	client := ec2.NewFromConfig(cfg)
	fmt.Printf("Provisioning EC2 Instance '%s' (Type: %s)...\n", name, instanceType)

	output, err := client.RunInstances(context.TODO(), &ec2.RunInstancesInput{
		ImageId:      aws.String(ami),
		InstanceType: ec2Types.InstanceType(instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create EC2 instance: %w", err)
	}

	instanceID := *output.Instances[0].InstanceId
	fmt.Printf("Successfully created EC2 Instance! ID: %s\n", instanceID)
	return instanceID, nil
}

func deleteEC2Instance(cfg aws.Config, instanceID string) error {
	client := ec2.NewFromConfig(cfg)
	fmt.Printf("Destroying EC2 Instance '%s'...\n", instanceID)

	_, err := client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}
	fmt.Printf("Successfully triggered termination for EC2: %s\n", instanceID)
	return nil
}

func UpdateEC2Instance(res parser.Resource, instanceID string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(res.Region))
	if err != nil {
		return err
	}
	client := ec2.NewFromConfig(cfg)

	fmt.Printf("Stopping Instance %s for hardware upgrade...\n", instanceID)
	_, err = client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return err
	}

	fmt.Println("Waiting for instance to reach stopped state...")
	waiter := ec2.NewInstanceStoppedWaiter(client)
	err = waiter.Wait(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 2*time.Minute)
	if err != nil {
		return err
	}

	fmt.Printf("Applying new instance type: %s\n", res.InstanceType)
	_, err = client.ModifyInstanceAttribute(context.TODO(), &ec2.ModifyInstanceAttributeInput{
		InstanceId:   aws.String(instanceID),
		InstanceType: &ec2Types.AttributeValue{Value: aws.String(res.InstanceType)},
	})
	if err != nil {
		return err
	}

	fmt.Println("Restarting instance with new hardware...")
	_, err = client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})

	return err
}

func CreateKMSKey(ctx context.Context, cfg aws.Config, bucketName string) (string, error) {
	client := kms.NewFromConfig(cfg)

	desc := fmt.Sprintf("Encryption key for IaC bucket: %s", bucketName)

	input := &kms.CreateKeyInput{
		Description: aws.String(desc),
		KeyUsage:    kmsTypes.KeyUsageTypeEncryptDecrypt,
		Origin:      kmsTypes.OriginTypeAwsKms,
	}

	output, err := client.CreateKey(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create KMS key: %w", err)
	}

	keyId := *output.KeyMetadata.KeyId
	fmt.Printf("Successfully generated KMS Key: %s\n", keyId)

	return keyId, nil
}
