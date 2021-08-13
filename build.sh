docker build -t public.ecr.aws/i4r5n0t9/linuxkit-vsphere-config:v5 .
aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin public.ecr.aws
docker push public.ecr.aws/i4r5n0t9/linuxkit-vsphere-config:v5