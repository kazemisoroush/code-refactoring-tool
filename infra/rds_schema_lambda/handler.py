"""
Lambda to ensure Postgres schema is created.
"""
import os
import json
import psycopg2
import boto3
import traceback


def get_secret_value(secret_arn):
    """
    Get secret value from Secrets Manager.
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
    """
    Ensure the schema/table exists in the database.
    """
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
    Expects event like: { "table": "your_table_name" }
    """
    try:
        print("===== Lambda Invocation Start =====")
        print("Received event:", json.dumps(event, indent=2))

        table_name = event.get("table")
        if not table_name:
            raise ValueError("Missing 'table' in event")

        print("Reading environment variables...")
        secret_arn = os.environ["DB_SECRET_ARN"]
        print(f"port: {os.environ["DB_PORT"]}")
        db_host = os.environ["DB_HOST"]
        db_port = int(os.environ["DB_PORT"])
        db_name = os.environ["DB_NAME"]

        print("Fetching DB secret...")
        secret_data = get_secret_value(secret_arn)
        print("Secret fetched.")
        username = secret_data["username"]
        password = secret_data["password"]

        print("Connecting to Postgres...")
        conn = psycopg2.connect(
            dbname=db_name,
            user=username,
            password=password,
            host=db_host,
            port=db_port,
            connect_timeout=10
        )
        print("Postgres connection successful.")

        ensure_schema(conn, table_name)
        conn.close()

        print("===== Lambda Invocation Complete =====")
        return {
            "status": "success",
            "message": f"Schema ensured for table {table_name}"
        }

    except Exception as e:
        print("===== Lambda Unhandled Exception =====")
        traceback.print_exc()
        return {
            "status": "error",
            "message": str(e)
        }
