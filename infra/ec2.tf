data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_security_group" "ec2_sg" {
  name        = "scalable-counter-ec2-sg"
  description = "Allow HTTP and SSH access"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description     = "HTTP"
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = [aws_security_group.alb_sg.id]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = "scalable-counter-ec2-sg"
  }
}

resource "aws_instance" "app_server" {
  ami                         = data.aws_ami.al2023.id
  instance_type               = "t3.micro"
  subnet_id                   = element(data.aws_subnets.default.ids, 0)
  vpc_security_group_ids      = [aws_security_group.ec2_sg.id]
  iam_instance_profile        = aws_iam_instance_profile.ec2_profile.name
  associate_public_ip_address = true

  user_data = <<-EOF
              #!/bin/bash
              dnf update -y
              dnf install -y docker awscli

              systemctl start docker
              systemctl enable docker

              usermod -aG docker ec2-user

              aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin 646731024202.dkr.ecr.eu-central-1.amazonaws.com

              docker pull 646731024202.dkr.ecr.eu-central-1.amazonaws.com/scalable-counter:latest

              docker run -d \
                -p 80:8080 \
                -e AWS_REGION=eu-central-1 \
                --name counter-app \
                --restart always \
                646731024202.dkr.ecr.eu-central-1.amazonaws.com/scalable-counter:latest
              EOF

  tags = {
    Name = "scalable-counter-ec2-instance"
  }
}
