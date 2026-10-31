output "s3_bucket_name" {
  value       = aws_s3_bucket.urls_bucket.id
  description = "The S3 bucket name created for urls.json"
}

output "security_group_id" {
  value       = aws_security_group.ec2_sg.id
  description = "The public IP of the EC2 instance"
}

output "ec2_public_ip" {
  value       = aws_instance.site_checker_vm.public_ip
  description = "The ID of the EC2 Security Group"
}