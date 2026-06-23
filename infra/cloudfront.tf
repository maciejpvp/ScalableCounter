variable "Environment" {
  type        = string
  description = "The environment name"
  default     = "production"
}

resource "random_password" "cloudfront_secret" {
  length  = 32
  special = false
}

resource "aws_cloudfront_function" "strip_prefix" {
  name    = "strip-prefix-${var.Environment}"
  runtime = "cloudfront-js-2.0"
  comment = "Strips /api prefix from the URI"
  publish = true

  lifecycle {
    create_before_destroy = true
  }

  code = <<EOF
function handler(event) {
    var request = event.request;
    var uri = request.uri;

    if (uri.startsWith('/api')) {
        uri = uri.slice(4); 
    }

    // Ensure the URI always begins with a single '/'
    if (!uri.startsWith('/')) {
        uri = '/' + uri;
    }

    request.uri = uri;
    return request;
}
EOF
}

data "aws_cloudfront_cache_policy" "disabled" {
  name = "Managed-CachingDisabled"
}

data "aws_cloudfront_origin_request_policy" "all_viewer" {
  name = "Managed-AllViewer"
}

module "cloudfront" {
  source  = "terraform-aws-modules/cloudfront/aws"
  version = "~> 3.0"

  comment             = "CloudFront distribution for ScalableCounter application"
  enabled             = true
  price_class         = "PriceClass_100"
  retain_on_delete    = false
  wait_for_deployment = false

  create_origin_access_identity = false

  origin = {
    alb = {
      domain_name = aws_lb.app_alb.dns_name
      custom_origin_config = {
        http_port              = 80
        https_port             = 443
        origin_protocol_policy = "http-only"
        origin_ssl_protocols   = ["TLSv1.2"]
      }
      custom_header = [
        {
          name  = "X-From-CloudFront"
          value = random_password.cloudfront_secret.result
        }
      ]
    }
  }

  default_cache_behavior = {
    target_origin_id       = "alb"
    viewer_protocol_policy = "redirect-to-https"

    allowed_methods          = ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"]
    cached_methods           = ["GET", "HEAD"]
    compress                 = true
    use_forwarded_values     = false

    cache_policy_id          = data.aws_cloudfront_cache_policy.disabled.id
    origin_request_policy_id = data.aws_cloudfront_origin_request_policy.all_viewer.id

    function_association = {
      viewer-request = {
        function_arn = aws_cloudfront_function.strip_prefix.arn
      }
    }
  }

  viewer_certificate = {
    cloudfront_default_certificate = true
  }

  tags = {
    Environment = "production"
    Project     = "ScalableCounter"
  }
}

output "cloudfront_domain_name" {
  description = "The domain name of the CloudFront distribution"
  value       = module.cloudfront.cloudfront_distribution_domain_name
}
