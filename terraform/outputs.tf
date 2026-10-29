output "s3_bucket_name" {
  value       = aws_s3_bucket.urls_bucket.id
  description = "The S3 bucket name created for urls.json"
}