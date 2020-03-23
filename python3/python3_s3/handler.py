import decimal
import gzip
import io
import json
import os
from logging import getLogger
from time import sleep

import boto3
from aws_lambda_context import LambdaContext, LambdaDict

LOGGER = getLogger()
LOGGER.setLevel("INFO")


def entry(event: LambdaDict, context: LambdaContext):
    LOGGER.info("entry point.")

    LOGGER.info("cleanup output bucket.")
    output_bucket_name = os.getenv("OutputBucketName")
    s3 = boto3.resource("s3")
    output_bucket = s3.Bucket(output_bucket_name)
    output_bucket.objects.delete()

    LOGGER.info("end entry point.")
    return []


def read_gzip_content(bytes, encoding="utf8"):
    json_file = gzip.open(io.BytesIO(bytes), "rt", encoding=encoding)
    return json_file.read()


def make_gzip_content(results, encoding="utf8"):
    json_bytes = bytes(json.dumps(results, default=decimal.Decimal), encoding)
    return gzip.compress(json_bytes)


def read_write(event: LambdaDict, context: LambdaContext):
    LOGGER.info("read.")

    LOGGER.info("read bucket objects.")
    input_bucket_name = os.getenv("InputBucketName")
    output_bucket_name = os.getenv("OutputBucketName")
    s3 = boto3.resource("s3")
    input_bucket = s3.Bucket(input_bucket_name)
    results = []
    for object_summary in input_bucket.objects.all():
        response = object_summary.get()
        body = response["Body"].read()
        json_dict = json.loads(read_gzip_content(body))
        results.extend(json_dict)

    output_bucket = s3.Bucket(output_bucket_name)
    output_bucket.put_object(Key="python.json.gz", Body=make_gzip_content(results))

    LOGGER.info("end read.")
    return {
        "message": "Go Serverless v1.0! Your function executed successfully!",
        "event": event,
    }
