"""
Lambda to ensure Postgres schema is created.
"""
import json
import os
import time
import traceback
from dataclasses import dataclass
from typing import Any, Dict

import boto3
import psycopg2


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


class SecretManagerError(Exception):
    """Custom exception for Secrets Manager-related errors."""


class ConfigurationError(Exception):
    """Custom exception for configuration-related errors."""


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
    except ValueError as port_error:
        raise ConfigurationError(f"Invalid port number in DB_PORT: {port_error}") from port_error
    except Exception as config_error:
        raise ConfigurationError(f"Failed to load database configuration: {config_error}") from config_error


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

    except json.JSONDecodeError as json_error:
        raise SecretManagerError(f"Failed to parse secret as JSON: {json_error}") from json_error
    except Exception as secret_error:
        print("Error fetching secret:")
        traceback.print_exc()
        raise SecretManagerError(f"Error getting secret from Secrets Manager: {secret_error}") from secret_error


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

    except psycopg2.Error as db_error:
        raise DatabaseError(f"PostgreSQL error while ensuring database exists: {db_error}") from db_error
    except Exception as general_error:
        print(f"Error ensuring database exists: {general_error}")
        traceback.print_exc()
        raise DatabaseError(f"Unexpected error while ensuring database exists: {general_error}") from general_error
    finally:
        if conn:
            conn.close()


def ensure_schema(config: DatabaseConfig, username: str, password: str, table_name: str) -> None:
    """
    Ensure the schema/table exists in the database with proper migration support.

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
            # First, create the pgvector extension if it doesn't exist
            cursor.execute("CREATE EXTENSION IF NOT EXISTS vector;")

            # Get embedding dimensions from environment variable
            embedding_dimensions = os.getenv("EMBEDDING_DIMENSIONS", "1536")

            # Check if table exists and get its current schema
            cursor.execute("""
                SELECT column_name, data_type, udt_name
                FROM information_schema.columns
                WHERE table_name = %s
                ORDER BY ordinal_position
            """, (table_name,))

            columns = cursor.fetchall()

            if columns:
                print(f"Table {table_name} exists with {len(columns)} columns")
                column_info = {col[0]: (col[1], col[2]) for col in columns}

                needs_migration = False
                migration_issues = []

                # Check ID column
                if 'id' in column_info:
                    _, id_udt_name = column_info['id']
                    if id_udt_name != 'uuid':
                        needs_migration = True
                        migration_issues.append(f"ID column type: {id_udt_name} (needs uuid)")
                else:
                    needs_migration = True
                    migration_issues.append("Missing ID column")

                # Check embedding column
                if 'embedding' in column_info:
                    _, emb_udt_name = column_info['embedding']
                    if emb_udt_name != 'vector':
                        needs_migration = True
                        migration_issues.append(f"Embedding column type: {emb_udt_name} (needs vector)")
                else:
                    needs_migration = True
                    migration_issues.append("Missing embedding column")

                # Check metadata column
                if 'metadata' not in column_info:
                    needs_migration = True
                    migration_issues.append("Missing metadata column")

                # Always check for required indexes, even if schema seems correct
                needs_index_update = False
                index_issues = []

                # Check if the text search index exists
                cursor.execute("""
                    SELECT indexname FROM pg_indexes
                    WHERE tablename = %s AND indexdef LIKE '%%gin%%to_tsvector%%'
                """, (table_name,))
                
                text_index_exists = cursor.fetchone()
                if not text_index_exists:
                    needs_index_update = True
                    index_issues.append("Missing text search index")

                # Check if any vector index exists (HNSW or IVFFlat)
                cursor.execute("""
                    SELECT indexname FROM pg_indexes
                    WHERE tablename = %s AND (
                        indexdef LIKE '%%hnsw%%vector_cosine_ops%%' OR 
                        indexdef LIKE '%%ivfflat%%vector_cosine_ops%%'
                    )
                """, (table_name,))
                
                vector_index_exists = cursor.fetchone()
                if not vector_index_exists:
                    needs_index_update = True
                    index_issues.append("Missing vector index")

                if needs_migration:
                    print(f"Table {table_name} needs migration. Issues: {migration_issues}")

                    # For testing/development, we can perform automatic migration
                    # In production, you might want to be more careful
                    auto_migrate = os.getenv("AUTO_MIGRATE_SCHEMA", "false").lower() == "true"

                    if auto_migrate:
                        print(f"Performing automatic migration for table {table_name}")
                        _migrate_table_schema(cursor, table_name, embedding_dimensions, column_info)
                    else:
                        print(f"Auto-migration disabled. Table {table_name} requires manual migration.")
                        # Log each issue for clarity
                        for issue in migration_issues:
                            print(f"Migration issue: {issue}")
                        conn.commit()
                        return
                elif needs_index_update:
                    print(f"Table {table_name} has correct schema but needs index updates. Issues: {index_issues}")
                    
                    # Create missing indexes
                    if not text_index_exists:
                        print(f"Creating missing text search index for table {table_name}")
                        create_index_sql = f"""
                            CREATE INDEX IF NOT EXISTS {table_name}_text_gin_idx
                            ON {table_name} USING gin (to_tsvector('simple', text))
                        """
                        cursor.execute(create_index_sql)
                        print(f"Created text search index for table {table_name}")

                    if not vector_index_exists:
                        print(f"Creating missing vector index for table {table_name}")
                        # Try HNSW first, fallback to IVFFlat if HNSW is not available
                        try:
                            create_vector_index_sql = f"""
                                CREATE INDEX IF NOT EXISTS {table_name}_embedding_hnsw_idx
                                ON {table_name} USING hnsw (embedding vector_cosine_ops)
                            """
                            cursor.execute(create_vector_index_sql)
                            print(f"Created HNSW vector index for table {table_name}")
                        except psycopg2.Error as hnsw_error:
                            # Rollback the failed transaction before trying the fallback
                            conn.rollback()
                            if "access method" in str(hnsw_error) and "hnsw" in str(hnsw_error):
                                print(f"HNSW not available, falling back to IVFFlat: {hnsw_error}")
                                try:
                                    create_vector_index_sql = f"""
                                        CREATE INDEX IF NOT EXISTS {table_name}_embedding_ivfflat_idx
                                        ON {table_name} USING ivfflat (embedding vector_cosine_ops)
                                        WITH (lists = 100)
                                    """
                                    cursor.execute(create_vector_index_sql)
                                    print(f"Created IVFFlat vector index for table {table_name}")
                                except psycopg2.Error as ivf_error:
                                    conn.rollback()
                                    print(f"Both HNSW and IVFFlat failed, skipping vector index: {ivf_error}")
                                    # Continue without vector index - it's not critical for basic functionality
                            else:
                                raise hnsw_error

                    conn.commit()
                    return
            else:
                print(f"Table {table_name} does not exist, creating with proper schema")
                # Table doesn't exist, create it with proper vector type
                create_table_sql = f"""
                    CREATE TABLE {table_name} (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        text TEXT,
                        embedding vector({embedding_dimensions}),
                        metadata JSONB
                    )
                """
                cursor.execute(create_table_sql)

                # Create the required text search index for Bedrock Knowledge Base
                create_index_sql = f"""
                    CREATE INDEX IF NOT EXISTS {table_name}_text_gin_idx
                    ON {table_name} USING gin (to_tsvector('simple', text))
                """
                cursor.execute(create_index_sql)
                print(f"Created text search index for table {table_name}")

                # Create the required vector index for Bedrock Knowledge Base
                # Try HNSW first, fallback to IVFFlat if HNSW is not available
                try:
                    create_vector_index_sql = f"""
                        CREATE INDEX IF NOT EXISTS {table_name}_embedding_hnsw_idx
                        ON {table_name} USING hnsw (embedding vector_cosine_ops)
                    """
                    cursor.execute(create_vector_index_sql)
                    print(f"Created HNSW vector index for table {table_name}")
                except psycopg2.Error as hnsw_error:
                    # Rollback the failed transaction before trying the fallback
                    conn.rollback()
                    if "access method" in str(hnsw_error) and "hnsw" in str(hnsw_error):
                        print(f"HNSW not available, falling back to IVFFlat: {hnsw_error}")
                        try:
                            # Need to recreate the table and text index since we rolled back
                            create_table_sql = f"""
                                CREATE TABLE {table_name} (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    text TEXT,
                                    embedding vector({embedding_dimensions}),
                                    metadata JSONB
                                )
                            """
                            cursor.execute(create_table_sql)
                            
                            # Recreate the text index
                            create_index_sql = f"""
                                CREATE INDEX IF NOT EXISTS {table_name}_text_gin_idx
                                ON {table_name} USING gin (to_tsvector('simple', text))
                            """
                            cursor.execute(create_index_sql)
                            print(f"Recreated text search index for table {table_name}")
                            
                            # Now create the IVFFlat vector index
                            create_vector_index_sql = f"""
                                CREATE INDEX IF NOT EXISTS {table_name}_embedding_ivfflat_idx
                                ON {table_name} USING ivfflat (embedding vector_cosine_ops)
                                WITH (lists = 100)
                            """
                            cursor.execute(create_vector_index_sql)
                            print(f"Created IVFFlat vector index for table {table_name}")
                        except psycopg2.Error as ivf_error:
                            conn.rollback()
                            print(f"Both HNSW and IVFFlat failed, skipping vector index: {ivf_error}")
                            # Recreate at least the basic table without vector index
                            try:
                                create_table_sql = f"""
                                    CREATE TABLE {table_name} (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                        text TEXT,
                                        embeddings vector({embedding_dimensions}),
                                        metadata JSONB
                                    )
                                """
                                cursor.execute(create_table_sql)
                                
                                # Create the text index
                                create_index_sql = f"""
                                    CREATE INDEX IF NOT EXISTS {table_name}_text_gin_idx
                                    ON {table_name} USING gin (to_tsvector('simple', text))
                                """
                                cursor.execute(create_index_sql)
                                print(f"Created table and text index without vector index for {table_name}")
                            except psycopg2.Error as basic_error:
                                print(f"Failed to create basic table structure: {basic_error}")
                                raise basic_error
                    else:
                        raise hnsw_error

        conn.commit()
        print(f"Schema ensured for table: {table_name}")

    except psycopg2.Error as db_error:
        raise DatabaseError(f"PostgreSQL error while creating schema: {db_error}") from db_error
    except Exception as schema_error:
        print("Error creating table schema:")
        traceback.print_exc()
        raise DatabaseError(f"Unexpected error while creating schema: {schema_error}") from schema_error
    finally:
        if conn:
            conn.close()


def _migrate_table_schema(cursor, table_name: str, embedding_dimensions: str, existing_columns: dict) -> None:
    """
    Migrate existing table schema to the required format for Bedrock Knowledge Base.

    Args:
        cursor: Database cursor
        table_name: Name of the table to migrate
        embedding_dimensions: Target embedding dimensions
        existing_columns: Dictionary of existing column info

    Raises:
        DatabaseError: If migration fails
    """
    print(f"Starting schema migration for table {table_name}")

    try:
        # Begin transaction for safe migration
        cursor.execute("BEGIN;")

        # Create a backup table name
        backup_table = f"{table_name}_backup_{int(time.time())}"

        # Step 1: Check if there's existing data
        cursor.execute(f"SELECT COUNT(*) FROM {table_name}")
        row_count = cursor.fetchone()[0]

        if row_count > 0:
            print(f"Found {row_count} rows in {table_name}, creating backup")
            # Create backup of existing data
            cursor.execute(f"CREATE TABLE {backup_table} AS SELECT * FROM {table_name}")
            print(f"Backup created as {backup_table}")

        # Step 2: Drop existing table (we have backup if needed)
        cursor.execute(f"DROP TABLE {table_name}")
        print(f"Dropped existing table {table_name}")

        # Step 3: Create new table with correct schema
        create_table_sql = f"""
            CREATE TABLE {table_name} (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                text TEXT,
                embedding vector({embedding_dimensions}),
                metadata JSONB
            )
        """
        cursor.execute(create_table_sql)
        print(f"Created new table {table_name} with correct schema")

        # Create the required text search index for Bedrock Knowledge Base
        create_index_sql = f"""
            CREATE INDEX IF NOT EXISTS {table_name}_text_gin_idx
            ON {table_name} USING gin (to_tsvector('simple', text))
        """
        cursor.execute(create_index_sql)
        print(f"Created text search index for table {table_name}")

        # Create the required vector index for Bedrock Knowledge Base
        # Try HNSW first, fallback to IVFFlat if HNSW is not available
        try:
            create_vector_index_sql = f"""
                CREATE INDEX IF NOT EXISTS {table_name}_embedding_hnsw_idx
                ON {table_name} USING hnsw (embedding vector_cosine_ops)
            """
            cursor.execute(create_vector_index_sql)
            print(f"Created HNSW vector index for table {table_name}")
        except psycopg2.Error as hnsw_error:
            # For migration, we're already in a transaction, so no manual rollback needed
            if "access method" in str(hnsw_error) and "hnsw" in str(hnsw_error):
                print(f"HNSW not available, falling back to IVFFlat: {hnsw_error}")
                try:
                    create_vector_index_sql = f"""
                        CREATE INDEX IF NOT EXISTS {table_name}_embedding_ivfflat_idx
                        ON {table_name} USING ivfflat (embedding vector_cosine_ops)
                        WITH (lists = 100)
                    """
                    cursor.execute(create_vector_index_sql)
                    print(f"Created IVFFlat vector index for table {table_name}")
                except psycopg2.Error as ivf_error:
                    print(f"Both HNSW and IVFFlat failed, continuing without vector index: {ivf_error}")
                    # Continue migration without vector index - it's not critical
            else:
                raise hnsw_error

        # Step 4: Migrate data if there was any
        if row_count > 0:
            # Try to migrate compatible data
            try:
                # Check what columns exist in backup
                if 'text' in existing_columns:
                    if 'metadata' in existing_columns:
                        # Migrate text and metadata, skip embedding (will need to be regenerated)
                        cursor.execute(f"""
                            INSERT INTO {table_name} (text, metadata)
                            SELECT text, metadata FROM {backup_table}
                        """)
                    else:
                        # Only migrate text
                        cursor.execute(f"""
                            INSERT INTO {table_name} (text)
                            SELECT text FROM {backup_table}
                        """)
                    print("Migrated text data from backup table")
                else:
                    print("No compatible text data found in backup table")

                print(f"Data migration completed. Backup table {backup_table} retained for safety")
            except (psycopg2.Error, ValueError, KeyError) as migration_error:
                print(f"Warning: Could not migrate data: {migration_error}")
                print(f"Backup table {backup_table} contains original data")

        # Commit the transaction
        cursor.execute("COMMIT;")
        print(f"Migration completed successfully for table {table_name}")

    except Exception as migration_error:
        # Rollback on error
        cursor.execute("ROLLBACK;")
        raise DatabaseError(f"Schema migration failed: {migration_error}") from migration_error


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


def lambda_handler(event: Dict[str, Any], context: Any) -> Dict[str, str]:  # pylint: disable=unused-argument
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

    except Exception as e:  # pylint: disable=broad-exception-caught
        print("===== Lambda Unhandled Exception =====")
        traceback.print_exc()
        response = LambdaResponse(status="error", message=f"Unexpected error: {e}")
        return response.to_dict()
