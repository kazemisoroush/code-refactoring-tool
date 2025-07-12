"""
Handler test package.
"""
import unittest
from unittest.mock import patch, MagicMock
import json
import os
from botocore.exceptions import ClientError

import handler


class TestSendResponse(unittest.TestCase):
    """
    Test send response.
    """

    @patch("urllib.request.urlopen")
    def test_send_response_success(self, mock_urlopen):
        """
        Test send response success
        """
        mock_response = MagicMock()
        mock_response.status = 200
        mock_urlopen.return_value.__enter__.return_value = mock_response

        event = {
            "ResponseURL": "http://example.com",
            "StackId": "stack-id",
            "RequestId": "request-id",
            "LogicalResourceId": "logical-id"
        }

        handler.send_response(event, "SUCCESS", "Reason goes here")
        mock_urlopen.assert_called_once()


class TestGetSecretValue(unittest.TestCase):
    """
    Test get secret value.
    """

    @patch("boto3.client")
    def test_get_secret_value_success(self, mock_boto_client):
        """
        Test get secret value success.
        """
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
        """
        Test get secret value missing.
        """
        mock_secrets = MagicMock()
        mock_boto_client.return_value = mock_secrets
        mock_secrets.get_secret_value.return_value = {"SecretString": None}

        with self.assertRaises(RuntimeError) as context:
            handler.get_secret_value("arn:missing")

        self.assertIn("SecretString is None", str(context.exception))

    @patch("boto3.client")
    def test_get_secret_value_error(self, mock_boto_client):
        """
        Test get secret value error.
        """
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
    Test ensure schema.
    """

    def test_ensure_schema_success(self):
        """
        Test ensure schema success.
        """
        conn = MagicMock()
        cursor = MagicMock()
        conn.cursor.return_value.__enter__.return_value = cursor

        handler.ensure_schema(conn, "my_table")

        cursor.execute.assert_called_once()
        conn.commit.assert_called_once()


class TestLambdaHandler(unittest.TestCase):
    """
    Test Lambda handler.
    """

    @patch("handler.send_response")
    def test_lambda_handler_delete(self, mock_send_response):
        """
        Test Lambda handler.
        """
        event = {
            "RequestType": "Delete",
            "ResourceProperties": {"TableName": "my_table"},
            "ResponseURL": "http://example.com",
            "StackId": "stack-id",
            "RequestId": "req-id",
            "LogicalResourceId": "logical-id"
        }

        handler.lambda_handler(event, {})
        mock_send_response.assert_called_once_with(event, "SUCCESS", "Delete request handled.")

    @patch("handler.get_secret_value")
    @patch("handler.psycopg2.connect")
    @patch("handler.send_response")
    def test_lambda_handler_create_success(self, mock_send_response, mock_connect, mock_get_secret):
        """
        Test Lambda handler create success.
        """
        os.environ["DB_SECRET_ARN"] = "arn:secret"
        os.environ["DB_HOST"] = "localhost"
        os.environ["DB_PORT"] = "5432"
        os.environ["DB_NAME"] = "testdb"

        event = {
            "RequestType": "Create",
            "ResourceProperties": {"TableName": "test_table"},
            "ResponseURL": "http://example.com",
            "StackId": "stack-id",
            "RequestId": "req-id",
            "LogicalResourceId": "logical-id"
        }

        mock_get_secret.return_value = {"username": "user", "password": "pass"}
        conn = MagicMock()
        mock_connect.return_value = conn

        handler.lambda_handler(event, {})

        mock_connect.assert_called_once()
        conn.close.assert_called_once()
        mock_send_response.assert_called_with(event, "SUCCESS", "Schema migration successful.")
