# Terraform configuration for SLOzy on AWS

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "slozy-terraform-state"
    key    = "production/terraform.tfstate"
    region = var.aws_region
    encrypt = true
  }
}

# Variables
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "domain_name" {
  description = "Domain name for SLOzy"
  type        = string
  default     = "slozy.example.com"
}

variable "environment" {
  description = "Environment"
  type        = string
  default     = "production"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "key_pair_name" {
  description = "SSH key pair name"
  type        = string
}

# VPC Configuration
resource "aws_vpc" "slozy_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "slozy-vpc"
    Environment = var.environment
  }
}

resource "aws_internet_gateway" "slozy_igw" {
  vpc_id = aws_vpc.slozy_vpc.id

  tags = {
    Name        = "slozy-igw"
    Environment = var.environment
  }
}

resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.slozy_vpc.id
  cidr_block              = "10.0.${count.index + 1}.0/24"
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name        = "slozy-public-${count.index}"
    Environment = var.environment
  }
}

resource "aws_default_security_group" "slozy" {
  vpc_id = aws_vpc.slozy_vpc.id

  # HTTP
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "slozy-sg"
    Environment = var.environment
  }
}

# EC2 Instance
resource "aws_instance" "slozy" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_default_security_group.slozy.id]
  key_name              = var.key_pair_name

  user_data = base64encode(templatefile("${path.module}/cloud-init.yaml", {
    domain_name = var.domain_name
    environment = var.environment
  }))

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 20
    delete_on_termination = true
    encrypted            = true
  }

  tags = {
    Name        = "slozy-instance"
    Environment = var.environment
    Service    = "slozy"
  }
}

# Load Balancer
resource "aws_lb" "slozy" {
  name               = "slozy-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_default_security_group.slozy.id]

  subnets = aws_subnet.public[*].id

  tags = {
    Name        = "slozy-lb"
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "slozy" {
  name     = "slozy-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.slozy_vpc.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval           = 30
    matcher            = "200"
    path              = "/health"
    port              = "traffic-port"
    protocol           = "HTTP"
    timeout            = 5
    unhealthy_threshold = 2
  }

  tags = {
    Name        = "slozy-tg"
    Environment = var.environment
  }

  depends_on = [aws_instance.slozy]
}

resource "aws_lb_listener" "slozy_http" {
  load_balancer_arn = aws_lb.slozy.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

data "aws_acm_certificate" "slozy" {
  domain      = var.domain_name
  types       = ["AMAZON_ISSUED"]
  most_recent = true
}

resource "aws_lb_listener" "slozy_https" {
  load_balancer_arn = aws_lb.slozy.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = data.aws_acm_certificate.slozy.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.slozy.arn
  }
}

# Route 53
resource "aws_route53_zone" "main" {
  name = var.domain_name
}

resource "aws_route53_record" "slozy" {
  zone_id = aws_route53_zone.main.zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = aws_lb.slozy.dns_name
    zone_id               = aws_lb.slozy.zone_id
    evaluate_target_health = true
  }
}

# S3 Bucket for data backups
resource "aws_s3_bucket" "slozy_backup" {
  bucket        = "slozy-backups-${var.environment}"
  force_destroy = true

  lifecycle_rule {
    enabled = true

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }

    transition {
      days          = 180
      storage_class = "DEEP_ARCHIVE"
    }

    expiration {
      days = 365
    }
  }

  tags = {
    Name        = "slozy-backup"
    Environment = var.environment
  }
}

# Data sources
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-*-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

# Outputs
output "instance_public_ip" {
  description = "Public IP of the EC2 instance"
  value       = aws_instance.slozy.public_ip
}

output "load_balancer_dns" {
  description = "DNS name of the load balancer"
  value       = aws_lb.slozy.dns_name
}