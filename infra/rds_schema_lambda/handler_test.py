"""
Test suite for Lambda handler functions that ensure Postgres schema creation.
"""
# pylint: disable=invalid-name  # Test method names follow unittest conventions

import unittest
from unittest.mock import patch, MagicMock
import json
import os
from botocore.exceptions import ClientError

import handler


class TestDatabaseConfig(unittest.TestCase):
    """Test DatabaseConfig loading."""

    @patch.dict(os.environ, {
        "DB_SECRET_ARN": "arn:secret",
        "DB_HOST": "localhost",
        "DB_PORT": "5432",
        "DB_NAME": "testdb"
    })
    def test_load_database_config_success(self):
        """Should load config successfully with all required vars."""
        config = handler.load_database_config()
        self.assertEqual(config.host, "localhost")
        self.assertEqual(config.port, 5432)
        self.assertEqual(config.name, "testdb")
        self.assertEqual(config.secret_arn, "arn:secret")

    def test_load_database_config_missing_vars(self):
        """Should raise ConfigurationError when vars are missing."""
        with patch.dict(os.environ, {}, clear=True):
            with self.assertRaises(handler.ConfigurationError) as context:
                handler.load_database_config()
            self.assertIn("Missing required environment variables", str(context.exception))

    @patch.dict(os.environ, {
        "DB_SECRET_ARN": "arn:secret",
        "DB_HOST": "localhost",
        "DB_PORT": "invalid",
        "DB_NAME": "testdb"
    })
    def test_load_database_config_invalid_port(self):
        """Should raise ConfigurationError for invalid port."""
        with self.assertRaises(handler.ConfigurationError) as context:
            handler.load_database_config()
        self.assertIn("Invalid port number", str(context.exception))


class TestGetSecretValue(unittest.TestCase):
    """Test get_secret_value function."""

    @patch("boto3.client")
    def test_get_secret_value_success(self, mock_boto_client):
        """Should return secret data when SecretString is present."""
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets

        secret_data = {"username": "test-user", "password": "test-pass"}
        mock_secrets.get_secret_value.return_value = {
            "SecretString": json.dumps(secret_data)
        }

        result = handler.get_secret_value("arn:aws:secretsmanager:xyz")
        self.assertEqual(result["username"], "test-user")
        self.assertEqual(result["password"], "test-pass")

    @patch("boto3.client")
    def test_get_secret_value_missing(self, mock_boto_client):
        """Should raise SecretManagerError when SecretString is None."""
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets
        mock_secrets.get_secret_value.return_value = {"SecretString": None}

        with self.assertRaises(handler.SecretManagerError) as context:
            handler.get_secret_value("arn:missing")

        self.assertIn("SecretString is None", str(context.exception))

    @patch("boto3.client")
    def test_get_secret_value_missing_fields(self, mock_boto_client):
        """Should raise SecretManagerError when required fields are missing."""
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets
        mock_secrets.get_secret_value.return_value = {
            "SecretString": json.dumps({"username": "test-user"})  # missing password
        }

        with self.assertRaises(handler.SecretManagerError) as context:
            handler.get_secret_value("arn:missing")

        self.assertIn("Secret missing required fields", str(context.exception))

    @patch("boto3.client")
    def test_get_secret_value_error(self, mock_boto_client):
        """Should raise SecretManagerError when Secrets Manager call fails."""
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets
        mock_secrets.get_secret_value.side_effect = ClientError(
            {"Error": {"Code": "AccessDeniedException", "Message": "Denied!"}},
            "GetSecretValue"
        )

        with self.assertRaises(handler.SecretManagerError) as context:
            handler.get_secret_value("arn:denied")

        self.assertIn("Error getting secret from Secrets Manager", str(context.exception))


class TestValidateEvent(unittest.TestCase):
    """Test validate_event function."""

    def test_validate_event_success(self):
        """Should return table name for valid event."""
        event = {"table": "my_table"}
        result = handler.validate_event(event)
        self.assertEqual(result, "my_table")

    def test_validate_event_missing_table(self):
        """Should raise ConfigurationError when table is missing."""
        event = {}
        with self.assertRaises(handler.ConfigurationError) as context:
            handler.validate_event(event)
        self.assertIn("Missing 'table' in event", str(context.exception))

    def test_validate_event_empty_table(self):
        """Should raise ConfigurationError when table is empty."""
        event = {"table": ""}
        with self.assertRaises(handler.ConfigurationError) as context:
            handler.validate_event(event)
        self.assertIn("Table name must be a non-empty string", str(context.exception))

    def test_validate_event_invalid_type(self):
        """Should raise ConfigurationError when event is not dict."""
        with self.assertRaises(handler.ConfigurationError) as context:
            handler.validate_event("invalid")
        self.assertIn("Event must be a dictionary", str(context.exception))


class TestEnsureSchema(unittest.TestCase):
    """Test ensure_schema function."""

    @patch.dict(os.environ, {"EMBEDDING_DIMENSIONS": "1024"})
    @patch.dict(os.environ, {"EMBEDDING_DIMENSIONS": "1536"})
    @patch("handler.psycopg2.connect")
    def test_ensure_schema_success(self, mock_connect):
        """Should call execute and commit when schema creation is successful."""
        mock_conn = MagicMock()
        mock_connect.return_value = mock_conn
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__.return_value = mock_cursor

        # Mock that table doesn't exist (no columns found)
        mock_cursor.fetchall.return_value = []

        config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        handler.ensure_schema(config, "user", "pass", "my_table")

        # Should execute extension creation, schema check, table creation, text index, and vector index creation
        assert mock_cursor.execute.call_count == 5

        # Check that the extension is created first
        first_call = mock_cursor.execute.call_args_list[0][0][0]
        assert "CREATE EXTENSION IF NOT EXISTS vector" in first_call

        # Check that it queries for existing schema
        second_call = mock_cursor.execute.call_args_list[1][0][0]
        assert "information_schema.columns" in second_call

        # Check that the table creation uses UUID and vector types
        third_call = mock_cursor.execute.call_args_list[2][0][0]
        assert "id UUID PRIMARY KEY" in third_call
        assert "vector(1536)" in third_call
        assert "metadata JSONB" in third_call
        assert "my_table" in third_call

        # Check that the text search index is created
        fourth_call = mock_cursor.execute.call_args_list[3][0][0]
        assert "CREATE INDEX" in fourth_call
        assert "gin" in fourth_call
        assert "to_tsvector" in fourth_call
        assert "text" in fourth_call

        # Check that the vector index is created
        fifth_call = mock_cursor.execute.call_args_list[4][0][0]
        assert "CREATE INDEX" in fifth_call
        assert "hnsw" in fifth_call
        assert "vector_cosine_ops" in fifth_call
        assert "embedding" in fifth_call

        mock_conn.commit.assert_called_once()
        mock_conn.close.assert_called_once()

    @patch("handler.psycopg2.connect")
    def test_ensure_schema_table_exists_correct_schema(self, mock_connect):
        """Should skip table creation when table exists with correct vector schema but check for index."""
        mock_conn = MagicMock()
        mock_connect.return_value = mock_conn
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__.return_value = mock_cursor

        # Mock that table exists with correct schema (id: uuid, embedding: vector)
        # Use side_effect to return different results for different queries
        def mock_fetchall_side_effect():
            if "information_schema.columns" in mock_cursor.execute.call_args_list[-1][0][0]:
                return [
                    ("embedding", "USER-DEFINED", "vector"),
                    ("id", "USER-DEFINED", "uuid"),
                    ("metadata", "jsonb", "jsonb"),
                    ("text", "text", "text")
                ]
            if "pg_indexes" in mock_cursor.execute.call_args_list[-1][0][0]:
                return [("my_table_text_gin_idx",)]  # Index exists
            return []

        def mock_fetchone_side_effect():
            # Check the specific query to return appropriate index info
            if "pg_indexes" in mock_cursor.execute.call_args_list[-1][0][0]:
                query = mock_cursor.execute.call_args_list[-1][0][0]
                if "gin" in query and "to_tsvector" in query:
                    return ("my_table_text_gin_idx",)  # Text index exists
                elif ("hnsw" in query or "ivfflat" in query) and "vector_cosine_ops" in query:
                    return ("my_table_embedding_hnsw_idx",)  # Vector index exists
            return None

        mock_cursor.fetchall.side_effect = mock_fetchall_side_effect
        mock_cursor.fetchone.side_effect = mock_fetchone_side_effect

        config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        handler.ensure_schema(config, "user", "pass", "my_table")

        # Should execute extension creation, schema check, text index check, and vector index check
        assert mock_cursor.execute.call_count == 4

        # Check that the extension is created first
        first_call = mock_cursor.execute.call_args_list[0][0][0]
        assert "CREATE EXTENSION IF NOT EXISTS vector" in first_call

        # Check that it queries for existing schema
        second_call = mock_cursor.execute.call_args_list[1][0][0]
        assert "information_schema.columns" in second_call

        # Check that it queries for existing text index
        third_call = mock_cursor.execute.call_args_list[2][0][0]
        assert "pg_indexes" in third_call
        assert "gin" in third_call
        assert "to_tsvector" in third_call

        # Check that it queries for existing vector index
        fourth_call = mock_cursor.execute.call_args_list[3][0][0]
        assert "pg_indexes" in fourth_call
        assert ("hnsw" in fourth_call or "ivfflat" in fourth_call)
        assert "vector_cosine_ops" in fourth_call

        mock_conn.commit.assert_called_once()
        mock_conn.close.assert_called_once()

    @patch("handler.psycopg2.connect")
    def test_ensure_schema_table_exists_wrong_schema(self, mock_connect):
        """Should skip table creation when table exists with wrong schema to avoid data loss."""
        mock_conn = MagicMock()
        mock_connect.return_value = mock_conn
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__.return_value = mock_cursor

        # Mock that table exists with wrong type (old array type and varchar id)
        mock_cursor.fetchall.return_value = [
            ("embedding", "ARRAY", "_float8"),
            ("id", "character varying", "varchar")
        ]

        config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        handler.ensure_schema(config, "user", "pass", "my_table")

        # Should execute extension creation, schema check, and index checks (no table creation)
        assert mock_cursor.execute.call_count == 4

        # Check that the extension is created first
        first_call = mock_cursor.execute.call_args_list[0][0][0]
        assert "CREATE EXTENSION IF NOT EXISTS vector" in first_call

        # Check that it queries for existing schema
        second_call = mock_cursor.execute.call_args_list[1][0][0]
        assert "information_schema.columns" in second_call

        mock_conn.commit.assert_called_once()
        mock_conn.close.assert_called_once()

    @patch("handler.psycopg2.connect")
    def test_ensure_schema_table_exists_wrong_id_type(self, mock_connect):
        """Should skip table creation when table exists with wrong ID column type."""
        mock_conn = MagicMock()
        mock_connect.return_value = mock_conn
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__.return_value = mock_cursor

        # Mock that table exists with correct embedding but wrong ID type
        mock_cursor.fetchall.return_value = [
            ("embedding", "USER-DEFINED", "vector"),
            ("id", "character varying", "varchar")
        ]

        config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        handler.ensure_schema(config, "user", "pass", "my_table")

        # Should execute extension creation, schema check, and index checks (no table creation)
        assert mock_cursor.execute.call_count == 4

        # Check that the extension is created first
        first_call = mock_cursor.execute.call_args_list[0][0][0]
        assert "CREATE EXTENSION IF NOT EXISTS vector" in first_call

        # Check that it queries for existing schema
        second_call = mock_cursor.execute.call_args_list[1][0][0]
        assert "information_schema.columns" in second_call

        mock_conn.commit.assert_called_once()
        mock_conn.close.assert_called_once()


class TestLambdaHandler(unittest.TestCase):
    """Test lambda_handler function."""

    @patch("handler.get_secret_value")
    @patch("handler.ensure_database_exists")
    @patch("handler.ensure_schema")
    @patch("handler.load_database_config")
    def test_lambda_handler_success(self, mock_load_config, mock_ensure_schema, mock_ensure_db, mock_get_secret):
        """Should return success when all operations complete successfully."""
        # Setup mocks
        mock_config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        mock_load_config.return_value = mock_config
        mock_get_secret.return_value = {"username": "user", "password": "pass"}

        event = {"table": "my_table"}
        result = handler.lambda_handler(event, {})

        self.assertEqual(result["status"], "success")
        self.assertIn("Schema ensured", result["message"])

        # Verify all functions were called
        mock_load_config.assert_called_once()
        mock_get_secret.assert_called_once_with("arn:secret")
        mock_ensure_db.assert_called_once_with(mock_config, "user", "pass")
        mock_ensure_schema.assert_called_once_with(mock_config, "user", "pass", "my_table")

    def test_lambda_handler_missing_table(self):
        """Should return error when 'table' is missing in event."""
        event = {}  # No "table" key
        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "error")
        self.assertIn("Configuration error", result["message"])

    @patch("handler.load_database_config", side_effect=handler.ConfigurationError("Config failed"))
    def test_lambda_handler_configuration_error(self, mock_load_config):  # pylint: disable=unused-argument
        """Should return configuration error when config loading fails."""
        event = {"table": "my_table"}
        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "error")
        self.assertIn("Configuration error", result["message"])

    @patch("handler.load_database_config")
    @patch("handler.get_secret_value", side_effect=handler.SecretManagerError("Secret failed"))
    def test_lambda_handler_secret_error(self, mock_get_secret, mock_load_config):  # pylint: disable=unused-argument
        """Should return secret manager error when secret retrieval fails."""
        mock_config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        mock_load_config.return_value = mock_config

        event = {"table": "my_table"}
        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "error")
        self.assertIn("Secrets Manager error", result["message"])

    @patch("handler.load_database_config")
    @patch("handler.get_secret_value")
    @patch("handler.ensure_database_exists", side_effect=handler.DatabaseError("DB failed"))
    def test_lambda_handler_database_error(self, mock_ensure_db, mock_get_secret, mock_load_config):  # pylint: disable=unused-argument
        """Should return database error when database operations fail."""
        mock_config = handler.DatabaseConfig("localhost", 5432, "testdb", "arn:secret")
        mock_load_config.return_value = mock_config
        mock_get_secret.return_value = {"username": "user", "password": "pass"}

        event = {"table": "my_table"}
        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "error")
        self.assertIn("Database error", result["message"])


if __name__ == '__main__':
    unittest.main()
