variable "aws_region" {
  type    = string
  default = "eu-central-1"
}

variable "my_ip" {
  type = string
  description = "Local public IP address for SSH access"
  default = "0.0.0.0/0" # Override locally in terraform.tfvars with "x.x.x.x/32"
}

variable "bucket_name" {
  type        = string
  description = "Unique name for the S3 bucket storing urls.json"
  default     = "my-site-checker-urls-bucket-192837" # Change to a unique name
}

variable "key_name" {
  type = string
  description = "SSH key name"
  default = ""
}

variable "public_key_string" {
  type = string
  description = "SSH public key value"
  default = ""
}