"""
Test suite for Lambda handler functions that ensure Postgres schema creation.
"""

import unittest
from unittest.mock import patch, MagicMock
import json
import os
from botocore.exceptions import ClientError

import handler


class TestGetSecretValue(unittest.TestCase):
    """
    Test get_secret_value function.
    """

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
        """Should raise RuntimeError when SecretString is None."""
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets
        mock_secrets.get_secret_value.return_value = {"SecretString": None}

        with self.assertRaises(RuntimeError) as context:
            handler.get_secret_value("arn:missing")

        self.assertIn("SecretString is None", str(context.exception))

    @patch("boto3.client")
    def test_get_secret_value_error(self, mock_boto_client):
        """Should raise RuntimeError when Secrets Manager call fails."""
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets
        mock_secrets.get_secret_value.side_effect = ClientError(
            {"Error": {"Code": "AccessDeniedException", "Message": "Denied!"}},
            "GetSecretValue"
        )

        with self.assertRaises(RuntimeError) as context:
            handler.get_secret_value("arn:denied")

        self.assertIn("Error getting secret", str(context.exception))


class TestEnsureSchema(unittest.TestCase):
    """
    Test ensure_schema function.
    """

    def test_ensure_schema_success(self):
        """Should call execute and commit when schema creation is successful."""
        conn = MagicMock()
        cursor = MagicMock()
        conn.cursor.return_value.__enter__.return_value = cursor

        handler.ensure_schema(conn, "my_table")

        cursor.execute.assert_called_once()
        conn.commit.assert_called_once()


class TestLambdaHandler(unittest.TestCase):
    """
    Test lambda_handler function.
    """

    @patch("handler.get_secret_value")
    @patch("handler.psycopg2.connect")
    @patch.dict(os.environ, {
        "DB_SECRET_ARN": "arn:secret",
        "DB_HOST": "localhost",
        "DB_PORT": "5432",
        "DB_NAME": "testdb"
    })
    def test_lambda_handler_success(self, mock_connect, mock_get_secret):
        """Should return success when table creation is successful."""
        mock_get_secret.return_value = {"username": "user", "password": "pass"}
        
        # Create mock connections for both database existence check and schema creation
        postgres_conn = MagicMock()
        target_conn = MagicMock()
        
        # Mock the cursor for database existence check to return that database exists
        cursor_mock = MagicMock()
        cursor_mock.fetchone.return_value = True  # Database exists
        postgres_conn.cursor.return_value.__enter__.return_value = cursor_mock
        
        # Set up mock_connect to return different connections for different calls
        mock_connect.side_effect = [postgres_conn, target_conn]

        event = {"table": "my_table"}

        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "success")
        self.assertIn("Schema ensured", result["message"])
        
        # Verify connect was called twice: once for postgres db, once for target db
        self.assertEqual(mock_connect.call_count, 2)
        
        # Verify the calls were made with correct parameters
        calls = mock_connect.call_args_list
        # First call should be to 'postgres' database
        self.assertEqual(calls[0][1]['dbname'], 'postgres')
        # Second call should be to target database
        self.assertEqual(calls[1][1]['dbname'], 'testdb')
        
        # Verify both connections were closed
        postgres_conn.close.assert_called_once()
        target_conn.close.assert_called_once()

    @patch("handler.get_secret_value")
    @patch("handler.psycopg2.connect", side_effect=Exception("Connection failed"))
    @patch.dict(os.environ, {
        "DB_SECRET_ARN": "arn:secret",
        "DB_HOST": "localhost",
        "DB_PORT": "5432",
        "DB_NAME": "testdb"
    })
    def test_lambda_handler_connection_failure(self, mock_connect, mock_get_secret):
        """Should return error when Postgres connection fails."""
        mock_get_secret.return_value = {"username": "user", "password": "pass"}

        event = {"table": "fail_table"}

        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "error")
        self.assertIn("Connection failed", result["message"])

    def test_lambda_handler_missing_table(self):
        """Should return error when 'table' is missing in event."""
        event = {}  # No "table" key
        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "error")
        self.assertIn("Missing 'table'", result["message"])

    @patch("handler.get_secret_value")
    @patch("handler.psycopg2.connect")
    @patch.dict(os.environ, {
        "DB_SECRET_ARN": "arn:secret",
        "DB_HOST": "localhost",
        "DB_PORT": "5432",
        "DB_NAME": "newdb"
    })
    def test_lambda_handler_creates_database(self, mock_connect, mock_get_secret):
        """Should create database when it doesn't exist."""
        mock_get_secret.return_value = {"username": "user", "password": "pass"}
        
        # Create mock connections
        postgres_conn = MagicMock()
        target_conn = MagicMock()
        
        # Mock the cursor for database existence check to return that database doesn't exist
        cursor_mock = MagicMock()
        cursor_mock.fetchone.return_value = None  # Database doesn't exist
        postgres_conn.cursor.return_value.__enter__.return_value = cursor_mock
        
        # Set up mock_connect to return different connections for different calls
        mock_connect.side_effect = [postgres_conn, target_conn]

        event = {"table": "my_table"}

        result = handler.lambda_handler(event, {})
        self.assertEqual(result["status"], "success")
        self.assertIn("Schema ensured", result["message"])
        
        # Verify connect was called twice
        self.assertEqual(mock_connect.call_count, 2)
        
        # Verify CREATE DATABASE was called
        cursor_mock.execute.assert_any_call('CREATE DATABASE "newdb"')
