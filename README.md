# serverless-lambda-s3-trial

Read gziped JSONs from S3, decode them and write a gzip file to S3.

Python version and Go version.

## QA

How do I put files to S3 ?
The answer is below.

```powershell
Write-S3Object -Recurse -BucketName serverless-s3-input-bucket -KeyPrefix download -Folder ./
```
