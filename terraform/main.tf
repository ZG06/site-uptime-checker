# Create S3 Bucket for urls.json
resource "aws_s3_bucket" "urls_bucket" {
  bucket = var.bucket_name
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "urls_bucket_privacy" {
  bucket = aws_s3_bucket.urls_bucket.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# IAM Role for EC2 Instance
resource "aws_iam_role" "ec2_s3_read_role" {
  name = "site_checker_ec2_s3_read_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

# Attach S3 Read-Only Policy to the Role
resource "aws_iam_role_policy_attachment" "s3_read_attach" {
  role       = aws_iam_role.ec2_s3_read_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
}

# Attach ECR Read-Only Policy so EC2 can pull images from ECR
resource "aws_iam_role_policy_attachment" "ecr_read_attach" {
  role       = aws_iam_role.ec2_s3_read_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

# Create an Instance Profile for EC2
resource "aws_iam_instance_profile" "ec2_profile" {
  name = "site_checker_ec2_profile"
  role = aws_iam_role.ec2_s3_read_role.name
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

# Create Security Group allowing SSH (Port 22) and Outbound internet traffic
resource "aws_security_group" "ec2_sg" {
  name        = "site_checker_sg"
  description = "Allow inbound SSH access"

  # Inbound rule for SSH
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.my_ip]
  }

  # Outbound rule (allows EC2 to reach ECR, S3, and public URLs)
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "deployer" {
  key_name = var.key_name
  public_key = var.public_key_string
}

# Spin up an EC2 Instance
resource "aws_instance" "site_checker_vm" {
  ami = data.aws_ami.amazon_linux.id
  instance_type = "t3.micro"
  iam_instance_profile = aws_iam_instance_profile.ec2_profile.name
  vpc_security_group_ids = [aws_security_group.ec2_sg.id]
  key_name = aws_key_pair.deployer.key_name

  # Boot script that runs as soon as the VM boots up
  user_data = <<-EOF
              #!/bin/bash
              sudo dnf update -y
              sudo dnf install docker -y
              sudo systemctl start docker
              sudo systemctl enable docker
              EOF

  tags = {
    Name = "Site-Checker-EC2"
  }
}
