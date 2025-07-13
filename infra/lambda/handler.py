"""
Handler create schema for Lambda.
"""
import os
import json
import time
import urllib.request
import urllib.error
import psycopg2
import boto3
import traceback


def send_response(event, status, reason):
    """
    Send response to CloudFormation
    """
    response_body = {
        "Status": status,
        "Reason": reason,
        "PhysicalResourceId": event.get("LogicalResourceId", "unknown"),
        "StackId": event.get("StackId"),
        "RequestId": event.get("RequestId"),
        "LogicalResourceId": event.get("LogicalResourceId")
    }

    data = json.dumps(response_body).encode("utf-8")
    req = urllib.request.Request(event["ResponseURL"], data=data, method="PUT")
    req.add_header("Content-Type", "")

    try:
        with urllib.request.urlopen(req) as response:
            print(f"CloudFormation response sent with status: {response.status}")
            print(f"CloudFormation response reason: {reason}")
    except urllib.error.HTTPError as e:
        print(f"HTTPError sending response to CloudFormation: {e.code} - {e.reason}")
        print("Response body:", e.read().decode())
        traceback.print_exc()
    except urllib.error.URLError as e:
        print(f"URLError sending response to CloudFormation: {e.reason}")
        traceback.print_exc()
    except Exception as e:
        print(f"Unexpected error sending response to CloudFormation: {e}")
        traceback.print_exc()


def get_secret_value(secret_arn):
    """
    Get secret value from ARN.
    """
    print(f"Fetching secret from Secrets Manager: {secret_arn}")
    try:
        client = boto3.client("secretsmanager")
        response = client.get_secret_value(SecretId=secret_arn)
        secret_string = response.get("SecretString")
        if not secret_string:
            raise ValueError("SecretString is None")
        return json.loads(secret_string)
    except Exception as e:
        print("Error fetching secret:")
        traceback.print_exc()
        raise RuntimeError(f"Error getting secret: {e}")


def ensure_schema(conn, table_name):
    print(f"Ensuring schema for table: {table_name}")
    try:
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
    except Exception:
        print("Error creating table schema:")
        traceback.print_exc()
        raise


def lambda_handler(event, _):
    """
    Lambda handler function.
    """
    try:
        print("===== Lambda Invocation Start =====")
        print("Received event:", json.dumps(event, indent=2))

        request_type = event.get("RequestType")
        if not request_type:
            raise ValueError("Missing RequestType in event")

        table_name = event.get("ResourceProperties", {}).get("TableName")
        if not table_name:
            raise ValueError("Missing TableName in ResourceProperties")

        if request_type == "Delete":
            print("Delete request received. Skipping schema drop for safety.")
            send_response(event, "SUCCESS", "Delete request handled.")
            return None

        print("Reading environment variables...")
        try:
            secret_arn = os.environ["DB_SECRET_ARN"]
            db_host = os.environ["DB_HOST"]
            db_port = os.environ["DB_PORT"]
            db_name = os.environ["DB_NAME"]
        except KeyError as e:
            print(f"Missing environment variable: {e}")
            raise

        print("Fetching DB secret...")
        secret_data = get_secret_value(secret_arn)
        print("Secret fetched.")
        username = secret_data["username"]
        password = secret_data["password"]

        conn_str = f"dbname={db_name} user={username} password=**** host={db_host} port={db_port}"
        print(f"Connecting to Postgres with: {conn_str}")

        conn = None
        try:
            conn = psycopg2.connect(
                dbname=db_name,
                user=username,
                password=password,
                host=db_host,
                port=db_port,
                connect_timeout=10
            )
            print("Postgres connection successful.")
        except Exception as e:
            print(f"Postgres connection failed: {e}")
            traceback.print_exc()
            time.sleep(5)

        if conn is None:
            raise ConnectionError("Failed to connect to Postgres after retries.")

        ensure_schema(conn, table_name)
        conn.close()
        send_response(event, "SUCCESS", "Schema migration successful.")
        print("===== Lambda Invocation Complete =====")
        return None

    except Exception as e:
        print("===== Lambda Unhandled Exception =====")
        traceback.print_exc()
        send_response(event, "FAILED", f"Exception: {str(e)}")
        return None
