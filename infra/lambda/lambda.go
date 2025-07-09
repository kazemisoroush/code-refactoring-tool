package lambda

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jackc/pgx/v5"
)

type Event struct {
	RequestType           string                 `json:"RequestType"` // Create, Update, Delete
	ServiceToken          string                 `json:"ServiceToken"`
	ResponseURL           string                 `json:"ResponseURL"`
	StackId               string                 `json:"StackId"`
	RequestId             string                 `json:"RequestId"`
	LogicalResourceId     string                 `json:"LogicalResourceId"`
	ResourceType          string                 `json:"ResourceType"`
	ResourceProperties    map[string]interface{} `json:"ResourceProperties"`
	OldResourceProperties map[string]interface{} `json:"OldResourceProperties"`
}

// Function to send a response back to CloudFormation
func sendResponse(event Event, status string, reason string) error {
	responseBody := map[string]interface{}{
		"Status":             status,
		"Reason":             reason,
		"PhysicalResourceId": event.LogicalResourceId, // Or a more specific ID if needed
		"StackId":            event.StackId,
		"RequestId":          event.RequestId,
		"LogicalResourceId":  event.LogicalResourceId,
	}
	jsonBody, _ := json.Marshal(responseBody)

	req, err := http.NewRequest("PUT", event.ResponseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request to CloudFormation: %w", err)
	}
	req.Header.Set("Content-Type", "") // Required for S3 pre-signed URL

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send response to CloudFormation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CloudFormation responded with status %d", resp.StatusCode)
	}
	return nil
}

// getSecretValue fetches a secret from AWS Secrets Manager
func getSecretValue(secretARN string) (map[string]string, error) {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}
	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("secret string is nil")
	}

	var secretData map[string]string
	err = json.Unmarshal([]byte(*result.SecretString), &secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret string: %w", err)
	}
	return secretData, nil
}

// Your Schema Management Logic (similar to your storage package)
func ensureSchema(ctx context.Context, conn *pgx.Conn, tableName string) error {
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id VARCHAR(255) PRIMARY KEY,
            text TEXT,
            embedding VECTOR,
            metadata JSON
        )`, tableName)

	_, err := conn.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create/check table '%s': %w", tableName, err)
	}
	fmt.Printf("Schema for table '%s' ensured successfully.\n", tableName)
	return nil
}

// Your handler for the Lambda Custom Resource
func handler(ctx context.Context, event Event) error {
	fmt.Printf("Received event: %+v\n", event)

	// Extract connection details from environment variables
	secretARN := os.Getenv("DB_SECRET_ARN")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	tableName := event.ResourceProperties["TableName"].(string) // Get tableName from Custom Resource properties

	if event.RequestType == "Delete" {
		// Optionally handle deletion. For schema migrations, you typically do nothing
		// or perform a cleanup (e.g., DropSchema). Be careful with dropping in production!
		fmt.Println("Delete request received. Skipping schema drop for safety.")
		return sendResponse(event, "SUCCESS", "Delete request handled.")
	}

	// Fetch database credentials from Secrets Manager
	secretData, err := getSecretValue(secretARN)
	if err != nil {
		fmt.Printf("Error getting secret: %v\n", err)
		sendResponse(event, "FAILED", fmt.Sprintf("Error getting secret: %v", err))
		return err
	}

	username := secretData["username"]
	password := secretData["password"]

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s", // Use %s for port as it's from env variable
		username,
		password,
		dbHost,
		dbPort,
		dbName,
	)

	var pgConn *pgx.Conn
	// Retry connection as RDS might not be immediately available
	for i := 0; i < 5; i++ {
		pgConn, err = pgx.Connect(ctx, connStr)
		if err == nil {
			break
		}
		fmt.Printf("Failed to connect to Postgres (attempt %d): %v. Retrying...\n", i+1, err)
		time.Sleep(5 * time.Second) // Wait before retrying
	}
	if err != nil {
		sendResponse(event, "FAILED", fmt.Sprintf("Failed to connect to Postgres after retries: %v", err))
		return fmt.Errorf("failed to connect to Postgres after retries: %w", err)
	}
	defer pgConn.Close(ctx)

	// Execute schema migration
	err = ensureSchema(ctx, pgConn, tableName)
	if err != nil {
		fmt.Printf("Error ensuring schema: %v\n", err)
		sendResponse(event, "FAILED", fmt.Sprintf("Error ensuring schema: %v", err))
		return err
	}

	fmt.Println("Schema migration successful.")
	return sendResponse(event, "SUCCESS", "Schema migration successful.")
}

func main() {
	lambda.Start(handler)
}
