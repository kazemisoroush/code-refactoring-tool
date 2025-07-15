"""
Lambda to ensure Postgres schema is created.
"""
import os
import json
import psycopg2
import boto3
import traceback
from typing import Dict, Any, Optional
from dataclasses import dataclass


@dataclass
class DatabaseConfig:
    """Configuration for database connection."""
    host: str
    port: int
    name: str
    secret_arn: str


@dataclass
class LambdaResponse:
    """Standardized Lambda response."""
    status: str
    message: str
    
    def to_dict(self) -> Dict[str, str]:
        return {"status": self.status, "message": self.message}


class DatabaseError(Exception):
    """Custom exception for database-related errors."""
    pass


class SecretManagerError(Exception):
    """Custom exception for Secrets Manager-related errors."""
    pass


class ConfigurationError(Exception):
    """Custom exception for configuration-related errors."""
    pass


def load_database_config() -> DatabaseConfig:
    """
    Load database configuration from environment variables.
    
    Returns:
        DatabaseConfig: Database configuration object
        
    Raises:
        ConfigurationError: If required environment variables are missing
    """
    try:
        required_vars = ["DB_SECRET_ARN", "DB_HOST", "DB_PORT", "DB_NAME"]
        missing_vars = [var for var in required_vars if not os.environ.get(var)]
        
        if missing_vars:
            raise ConfigurationError(f"Missing required environment variables: {', '.join(missing_vars)}")
        
        return DatabaseConfig(
            host=os.environ["DB_HOST"],
            port=int(os.environ["DB_PORT"]),
            name=os.environ["DB_NAME"],
            secret_arn=os.environ["DB_SECRET_ARN"]
        )
    except ValueError as e:
        raise ConfigurationError(f"Invalid port number in DB_PORT: {e}")
    except Exception as e:
        raise ConfigurationError(f"Failed to load database configuration: {e}")


def get_secret_value(secret_arn: str) -> Dict[str, str]:
    """
    Get secret value from Secrets Manager.
    
    Args:
        secret_arn: ARN of the secret to retrieve
        
    Returns:
        Dict containing username and password
        
    Raises:
        SecretManagerError: If secret retrieval fails
    """
    print(f"Fetching secret from Secrets Manager: {secret_arn}")
    try:
        client = boto3.client("secretsmanager")
        response = client.get_secret_value(SecretId=secret_arn)
        secret_string = response.get("SecretString")
        
        if not secret_string:
            raise SecretManagerError("SecretString is None - secret may be corrupted or empty")
        
        secret_data = json.loads(secret_string)
        
        # Validate required fields
        required_fields = ["username", "password"]
        missing_fields = [field for field in required_fields if field not in secret_data]
        if missing_fields:
            raise SecretManagerError(f"Secret missing required fields: {', '.join(missing_fields)}")
        
        return secret_data
        
    except json.JSONDecodeError as e:
        raise SecretManagerError(f"Failed to parse secret as JSON: {e}")
    except Exception as e:
        print("Error fetching secret:")
        traceback.print_exc()
        raise SecretManagerError(f"Error getting secret from Secrets Manager: {e}")


def ensure_database_exists(config: DatabaseConfig, username: str, password: str) -> None:
    """
    Ensure the target database exists, create it if it doesn't.
    
    Args:
        config: Database configuration
        username: Database username
        password: Database password
        
    Raises:
        DatabaseError: If database operations fail
    """
    print(f"Ensuring database '{config.name}' exists...")
    conn = None
    try:
        # Connect to the default 'postgres' database first
        conn = psycopg2.connect(
            dbname="postgres",
            user=username,
            password=password,
            host=config.host,
            port=config.port,
            connect_timeout=10
        )
        conn.autocommit = True  # Required for CREATE DATABASE
        
        with conn.cursor() as cursor:
            # Check if database exists
            cursor.execute("SELECT 1 FROM pg_database WHERE datname = %s", (config.name,))
            exists = cursor.fetchone()
            
            if not exists:
                print(f"Database '{config.name}' does not exist. Creating it...")
                # Use identifier quoting to handle special characters
                cursor.execute(f'CREATE DATABASE "{config.name}"')
                print(f"Database '{config.name}' created successfully.")
            else:
                print(f"Database '{config.name}' already exists.")
                
    except psycopg2.Error as e:
        raise DatabaseError(f"PostgreSQL error while ensuring database exists: {e}")
    except Exception as e:
        print(f"Error ensuring database exists: {e}")
        traceback.print_exc()
        raise DatabaseError(f"Unexpected error while ensuring database exists: {e}")
    finally:
        if conn:
            conn.close()


def ensure_schema(config: DatabaseConfig, username: str, password: str, table_name: str) -> None:
    """
    Ensure the schema/table exists in the database.
    
    Args:
        config: Database configuration
        username: Database username
        password: Database password
        table_name: Name of the table to create
        
    Raises:
        DatabaseError: If schema operations fail
    """
    print(f"Ensuring schema for table: {table_name}")
    conn = None
    try:
        conn = psycopg2.connect(
            dbname=config.name,
            user=username,
            password=password,
            host=config.host,
            port=config.port,
            connect_timeout=10
        )
        
        with conn.cursor() as cursor:
            create_table_sql = f"""
                CREATE TABLE IF NOT EXISTS {table_name} (
                    id VARCHAR(255) PRIMARY KEY,
                    text TEXT,
                    embedding FLOAT8[],
                    metadata JSON
                )
            """
            cursor.execute(create_table_sql)
        
        conn.commit()
        print(f"Schema ensured for table: {table_name}")
        
    except psycopg2.Error as e:
        raise DatabaseError(f"PostgreSQL error while creating schema: {e}")
    except Exception as e:
        print("Error creating table schema:")
        traceback.print_exc()
        raise DatabaseError(f"Unexpected error while creating schema: {e}")
    finally:
        if conn:
            conn.close()


def validate_event(event: Dict[str, Any]) -> str:
    """
    Validate the incoming Lambda event.
    
    Args:
        event: Lambda event dictionary
        
    Returns:
        str: Table name from the event
        
    Raises:
        ConfigurationError: If event is invalid
    """
    if not isinstance(event, dict):
        raise ConfigurationError("Event must be a dictionary")
    
    table_name = event.get("table")
    if table_name is None:
        raise ConfigurationError("Missing 'table' in event")
    
    if not isinstance(table_name, str) or not table_name.strip():
        raise ConfigurationError("Table name must be a non-empty string")
    
    return table_name.strip()


def lambda_handler(event: Dict[str, Any], context: Any) -> Dict[str, str]:
    """
    Lambda handler function with improved error handling and modularity.
    
    Args:
        event: Lambda event dictionary, expects {"table": "table_name"}
        context: Lambda context (unused)
        
    Returns:
        Dict containing status and message
    """
    try:
        print("===== Lambda Invocation Start =====")
        print("Received event:", json.dumps(event, indent=2))

        # Step 1: Validate event
        table_name = validate_event(event)
        print(f"Validated table name: {table_name}")

        # Step 2: Load configuration
        config = load_database_config()
        print(f"Loaded database config for host: {config.host}")

        # Step 3: Get database credentials
        secret_data = get_secret_value(config.secret_arn)
        print("Database credentials retrieved successfully")

        # Step 4: Ensure database exists
        ensure_database_exists(config, secret_data["username"], secret_data["password"])

        # Step 5: Ensure schema exists
        ensure_schema(config, secret_data["username"], secret_data["password"], table_name)

        print("===== Lambda Invocation Complete =====")
        response = LambdaResponse(
            status="success",
            message=f"Schema ensured for table {table_name}"
        )
        return response.to_dict()

    except ConfigurationError as e:
        print(f"Configuration error: {e}")
        response = LambdaResponse(status="error", message=f"Configuration error: {e}")
        return response.to_dict()

    except SecretManagerError as e:
        print(f"Secrets Manager error: {e}")
        response = LambdaResponse(status="error", message=f"Secrets Manager error: {e}")
        return response.to_dict()

    except DatabaseError as e:
        print(f"Database error: {e}")
        response = LambdaResponse(status="error", message=f"Database error: {e}")
        return response.to_dict()

    except Exception as e:
        print("===== Lambda Unhandled Exception =====")
        traceback.print_exc()
        response = LambdaResponse(status="error", message=f"Unexpected error: {e}")
        return response.to_dict()
